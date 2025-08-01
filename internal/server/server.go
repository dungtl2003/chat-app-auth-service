package server

import (
	"context"
	"dungtl2003/chat-app-auth-service/internal/api"
	"dungtl2003/chat-app-auth-service/internal/config"
	ctx "dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/healthcheck"
	"dungtl2003/chat-app-auth-service/internal/httpclient"
	"dungtl2003/chat-app-auth-service/internal/logging"
	"dungtl2003/chat-app-auth-service/internal/password"
	"dungtl2003/chat-app-auth-service/internal/router"
	"dungtl2003/chat-app-auth-service/internal/services"
	"dungtl2003/chat-app-auth-service/internal/services/snowflake"
	"dungtl2003/chat-app-auth-service/internal/validate"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	srv    *http.Server
	appCtx *ctx.AppContext
}

// New creates a new AuthServer instance. The function will load the configuration
// and set up all necessary components. Call Run() to start the server. This function
// will exit the program if there is an error when creating.
func New() *Server {
	log.Println("Loading configuration")
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("NewConfig(): %v", err)
	}
	log.Printf("Configuration: %s\n", config)
	logger, err := logging.NewLogger(config.LogConfig.Level, config.LogConfig.Kind)
	if err != nil {
		log.Fatalf("failed to create logger, error: %v", err)
	}
	loggerWrapper := logging.NewLoggerWrapper(logger)

	log.Println("Creating validator")
	validator := validate.NewValidator()

	log.Println("Creating http client")
	client := httpclient.New()

	log.Println("Creating password manager")
	pm, err := password.NewBcryptPasswordManager(config.Cost)
	if err != nil {
		log.Fatalf("NewBcryptPasswordManager(): %v", err)
	}

	// Create a new ID generator service
	idGeneratorService, err := snowflake.New(config.SnowflakeConfig.Addr, config.SnowflakeConfig.CertDir, loggerWrapper)
	if err != nil {
		loggerWrapper.Errorfln("snowflake.New(): %v", err)
		os.Exit(1)
	}

	appCtx := ctx.NewAppCtx(loggerWrapper, validator, client, config.UserServiceURL, *config.JwtTokenConfig, config.DomainName, pm, idGeneratorService, []services.Service{idGeneratorService})

	log.Println("Creating router")
	handlers := []router.Handler{
		{
			Method: router.GET,
			Path:   "/healthcheck",
			H:      healthcheck.HealthCheck(appCtx),
		},
		{
			Method: router.GET,
			Path:   "/auth/check",
			H:      api.Authorize(appCtx),
		},
		{
			Method: router.POST,
			Path:   "/auth/refresh",
			H:      api.Refresh(appCtx),
		},
		{
			Method: router.POST,
			Path:   "/auth/logout",
			H:      api.Logout(appCtx),
		},

		{
			Method: router.POST,
			Path:   "/auth/login",
			H:      api.Login(appCtx),
		},
		{
			Method: router.POST,
			Path:   "/auth/signup",
			H:      api.SignUp(appCtx),
		},
	}
	r := router.New(logger, handlers...)

	log.Println("Creating server")
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", config.ServerPort),
		Handler: r,
	}

	return &Server{
		appCtx: appCtx,
		srv:    srv,
	}
}

// Run starts the server. The function will start the server and listen for signals
// to shut down the server. The function will exit the program if there is an error
// when starting the server. Call Close() to shut down the server.
func (s *Server) Run() {
	logger := s.appCtx.Logger

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("error when starting server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("received signal to shut down server")
	s.Close()
}

// Close shuts down the server. The function will close the database connection and
// shut down the server. The function will exit the program with status code 0 if
// the server is shut down successfully. The function will exit the program with
// status code 1 if there is an error when shutting down the server.
func (s *Server) Close() {
	s.appCtx.Logger.Info("shutting down server")

	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
		if err != nil {
			os.Exit(1)
		}

		os.Exit(0)
	}()

	for _, service := range s.appCtx.Services {
		if err = service.Close(); err != nil {
			s.appCtx.Logger.Errorfln("error when closing service [%s]: %v", service.Name(), err)
		} else {
			s.appCtx.Logger.Debugfln("service [%s] closed successfully", service.Name())
		}
	}

	if err = s.srv.Shutdown(ctx); err != nil {
		s.appCtx.Logger.Errorfln("error when shutting down server: %v", err)
	} else {
		s.appCtx.Logger.Infofln("server shut down")
	}
}
