package server

import (
	"context"
	"dungtl2003/chat-app-auth-service/internal/api"
	"dungtl2003/chat-app-auth-service/internal/config"
	ctx "dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/logging"
	"dungtl2003/chat-app-auth-service/internal/password"
	"dungtl2003/chat-app-auth-service/internal/router"
	"dungtl2003/chat-app-auth-service/internal/services"
	"dungtl2003/chat-app-auth-service/internal/services/idgen"
	"dungtl2003/chat-app-auth-service/internal/services/user"
	"dungtl2003/chat-app-auth-service/internal/validate"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type AuthServer struct {
	srv    *http.Server
	appCtx *ctx.AppContext
}

type AuthServerOptions struct {
	UserService        user.UserService
	IdGeneratorService idgen.IdGeneratorService // Optional ID generator service
}

// New creates a new AuthServer instance. The function will load the configuration
// and set up all necessary components. Call Run() to start the server. This function
// will exit the program if there is an error when creating.
func New(opts *AuthServerOptions) (*AuthServer, error) {
	log.Println("Loading configuration")
	config, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("error when loading configuration: %w", err)
	}
	log.Printf("Configuration: %s\n", config)

	log.Println("Creating logger")
	logger, err := logging.NewLogger(config.LogConfig.Level, config.LogConfig.Kind)
	if err != nil {
		return nil, fmt.Errorf("error when creating logger: %w", err)
	}
	loggerWrapper := logging.NewLoggerWrapper(logger)
	loggerWrapper.Infofln("Logger created successfully, switch to this logger")

	loggerWrapper.Infofln("Creating validator")
	validator := validate.NewValidator()

	loggerWrapper.Infofln("Creating password manager")
	pm, err := password.NewBcryptPasswordManager(config.Cost)
	if err != nil {
		log.Fatalf("NewBcryptPasswordManager(): %v", err)
	}

	var idGeneratorService idgen.IdGeneratorService
	if opts != nil && opts.IdGeneratorService != nil {
		loggerWrapper.Infofln("Using provided ID generator service")
		idGeneratorService = opts.IdGeneratorService
	} else {
		loggerWrapper.Infofln("Creating ID generator service")
		idGeneratorService, err = idgen.NewSnowflakeService(config.IdGeneratorConfig.Addr, &idgen.SnowflakeServiceOptions{
			Logger:  loggerWrapper,
			CertDir: config.IdGeneratorConfig.CertDir,
		})
		if err != nil {
			return nil, fmt.Errorf("error when creating ID generator service: %w", err)
		}
	}

	var userService user.UserService
	if opts != nil && opts.UserService != nil {
		loggerWrapper.Infofln("Using provided user service")
		userService = opts.UserService
	} else {
		loggerWrapper.Infofln("Creating user service")
		userService, err = user.NewUSerServiceV1(config.UserServiceConfig.URL, &user.UserServiceV1Options{
			Logger: loggerWrapper,
		})
		if err != nil {
			return nil, fmt.Errorf("error when creating user service: %w", err)
		}
	}

	loggerWrapper.Infofln("Creating application context")
	appCtx := &ctx.AppContext{
		Logger:          loggerWrapper,
		Validator:       validator,
		DomainName:      config.DomainName,
		PasswordManager: pm,
		JwtConfig:       config.JwtTokenConfig,

		IdGeneratorService: idGeneratorService,
		UserService:        userService,
		Services: []services.Service{
			idGeneratorService,
			userService,
		},
	}
	loggerWrapper.Infofln("Application context: %#v", appCtx)

	loggerWrapper.Infofln("Creating API handlers")
	handlers := []router.Handler{
		{
			Method: router.GET,
			Path:   "/healthcheck",
			H:      api.HealthCheck(appCtx),
		},
		{
			Method: router.GET,
			Path:   "/check",
			H:      api.Authorize(appCtx),
		},
		{
			Method: router.POST,
			Path:   "/refresh",
			H:      api.Refresh(appCtx),
		},
		{
			Method: router.POST,
			Path:   "/logout",
			H:      api.Logout(appCtx),
		},

		{
			Method: router.POST,
			Path:   "/login",
			H:      api.Login(appCtx),
		},
		{
			Method: router.POST,
			Path:   "/signup",
			H:      api.SignUp(appCtx),
		},
	}
	loggerWrapper.Infofln("Creating router with %d handlers", len(handlers))
	r := router.New(logger, handlers...)

	loggerWrapper.Infofln("Creating HTTP server")
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", config.ServerPort),
		Handler: r,
	}

	return &AuthServer{
		appCtx: appCtx,
		srv:    srv,
	}, nil
}

// Run starts the server. It listens for incoming HTTP requests and handles
// them according to the defined routes. Remember to call Close() to shut down
// the server gracefully.
func (s *AuthServer) Run() error {
	errSignal := make(chan error, 1)
	quit := make(chan os.Signal, 1)

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.appCtx.Logger.Error("error when starting server", "error", err)
			errSignal <- fmt.Errorf("error when starting server: %w", err)
		}
	}()
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errSignal:
		return err
	case <-quit:
		s.appCtx.Logger.Info("Received signal to shut down server")
		return s.Close()
	}

}

// Close shuts down the server and closes all services.
func (s *AuthServer) Close() error {
	s.appCtx.Logger.Info("shutting down server")

	var err error

	for _, service := range s.appCtx.Services {
		if err = service.Close(); err != nil {
			s.appCtx.Logger.Errorfln("Error when closing service [%s]: %v", service.Name(), err)
			return err
		}

		s.appCtx.Logger.Debugfln("Service [%s] closed successfully", service.Name())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = s.srv.Shutdown(ctx); err != nil {
		s.appCtx.Logger.Errorfln("Error when shutting down server: %v", err)
		return err
	}

	s.appCtx.Logger.Infofln("Server shut down")
	return nil
}
