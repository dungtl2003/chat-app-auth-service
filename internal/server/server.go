package server

import (
	"context"
	"dungtl2003/chat-app-auth-service/internal/api"
	"dungtl2003/chat-app-auth-service/internal/cache"
	"dungtl2003/chat-app-auth-service/internal/config"
	"dungtl2003/chat-app-auth-service/internal/logging"
	"dungtl2003/chat-app-auth-service/internal/password"
	"dungtl2003/chat-app-auth-service/internal/router"
	"dungtl2003/chat-app-auth-service/internal/services"
	"dungtl2003/chat-app-auth-service/internal/services/idgen"
	"dungtl2003/chat-app-auth-service/internal/services/mailer"
	"dungtl2003/chat-app-auth-service/internal/services/user"
	"dungtl2003/chat-app-auth-service/internal/validate"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type AuthServer struct {
	srv    *http.Server
	logger *logging.LoggerWrapper

	// Lifecycle management
	services []services.Service      // Things that need Close()
	workers  []func(context.Context) // Background tasks (Kafka consumers, etc)

	// Internal lifecycle management
	shutdownOnce sync.Once
	ctx          context.Context
	cancel       context.CancelFunc // To stop background workers
	wg           sync.WaitGroup     // To wait for background workers

}

type AuthServerOptions struct {
	UserService        user.UserService
	MailerService      mailer.MailerService
	IdGeneratorService idgen.IdGeneratorService
}

// New creates a new AuthServer instance. The function will load the configuration
// and set up all necessary components. Call Run() to start the server. This function
// will exit the program if there is an error when creating.
func New(opts *AuthServerOptions) (*AuthServer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	s := &AuthServer{
		services: make([]services.Service, 0),
		workers:  make([]func(context.Context), 0),
		cancel:   cancel,
		ctx:      ctx,
	}

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
	s.logger = loggerWrapper

	loggerWrapper.Infofln("Creating validator")
	validator := validate.NewValidator()

	loggerWrapper.Infofln("Creating password manager")
	pm, err := password.NewBcryptPasswordManager(config.PasswordManagerConfig.Cost)
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
		userService, err = user.NewUserServiceV1(config.UserServiceConfig.URL, &user.UserServiceV1Options{
			Logger: loggerWrapper,
		})
		if err != nil {
			return nil, fmt.Errorf("error when creating user service: %w", err)
		}
	}

	var mailerService mailer.MailerService
	if opts != nil && opts.MailerService != nil {
		loggerWrapper.Infofln("Using provided mailer service")
		mailerService = opts.MailerService
	} else {
		loggerWrapper.Infofln("Creating mailer service")
		mailerService, err = mailer.NewSmtpMailerService(&mailer.SmtpMailerOptions{
			Logger: loggerWrapper,
			Host:   config.SmtpConfig.Host,
			Port:   config.SmtpConfig.Port,
		})
	}

	redisClient, err := cache.NewRedisClient(ctx, cache.Config{
		Addrs:    config.RedisConfig.Addrs,
		Password: config.RedisConfig.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("error when creating redis client: %w", err)
	}
	loggerWrapper.Infofln("Redis client created successfully")

	loggerWrapper.Infofln("Creating application context")
	handlerDeps := &api.HandlerDeps{
		Logger:          loggerWrapper,
		Config:          config,
		Validator:       validator,
		PasswordManager: pm,
		RedisClient:     redisClient,

		IdGeneratorService: idGeneratorService,
		UserService:        userService,
		MailerService:      mailerService,
	}
	loggerWrapper.Infofln("Application context: %#v", handlerDeps)

	services := []services.Service{
		idGeneratorService,
		userService,
		mailerService,
	}
	s.services = services

	loggerWrapper.Infofln("Creating API handlers")
	handlers := []router.Handler{
		{
			Method: router.GET,
			Path:   "/healthcheck",
			H:      api.HealthCheck(handlerDeps),
		},
		{
			Method: router.GET,
			Path:   "/auth/check",
			H:      api.Authorize(handlerDeps),
		},
		{
			Method: router.POST,
			Path:   "/auth/refresh",
			H:      api.Refresh(handlerDeps),
		},
		{
			Method: router.POST,
			Path:   "/auth/logout",
			H:      api.Logout(handlerDeps),
		},

		{
			Method: router.POST,
			Path:   "/auth/login",
			H:      api.Login(handlerDeps),
		},
		{
			Method: router.POST,
			Path:   "/auth/signup",
			H:      api.SignUp(handlerDeps),
		},
		{
			Method: router.POST,
			Path:   "/auth/password-reset/request",
			H:      api.RequestPasswordReset(handlerDeps),
		},
		{
			Method: router.POST,
			Path:   "/auth/password-reset/confirm",
			H:      api.ResetPassword(handlerDeps),
		},
	}
	loggerWrapper.Infofln("Creating router with %d handlers", len(handlers))
	r := router.New(logger, handlers...)

	loggerWrapper.Infofln("Creating HTTP server")
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", config.ServerPort),
		Handler: r,
	}
	s.srv = srv

	return s, nil

}

// Run starts the server. It listens for incoming HTTP requests and handles
// them according to the defined routes. Remember to call Close() to shut down
// the server gracefully.
func (s *AuthServer) Run() error {
	// Start background workers
	for _, w := range s.workers {
		worker := w
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			worker(s.ctx)
		}()
	}

	// Start HTTP server
	go func() {
		s.logger.Infofln("HTTP server listening on %s", s.srv.Addr)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Errorfln("HTTP server error: %v", err)
			s.cancel() // stop workers if server fails
		}
	}()

	// Wait for OS signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	s.logger.Info("Shutdown signal received")

	return s.Close()
}

// Close shuts down the server and closes all services.
func (s *AuthServer) Close() error {
	var finalErr error

	// Ensure we only close once
	s.shutdownOnce.Do(func() {
		s.logger.Info("Starting graceful shutdown sequence...")

		// Stop Background Workers
		s.logger.Debug("Stopping background workers...")
		s.cancel()  // Cancel the context passed to workers
		s.wg.Wait() // Wait for them to finish their current task

		// Shutdown HTTP Server
		s.logger.Debug("Shutting down HTTP server...")

		// Create a timeout context specifically for the shutdown procedure
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			s.logger.Errorfln("HTTP shutdown error: %v", err)
			// We don't return immediately; we still want to close DB/Kafka
			finalErr = err
		}

		// Close External Resources (DB, Kafka, etc.)
		s.logger.Debug("Closing external services...")
		for _, service := range s.services {
			if err := service.Close(); err != nil {
				s.logger.Errorfln("Error closing service [%s]: %v", service.Name(), err)
				if finalErr == nil {
					finalErr = err
				}
			} else {
				s.logger.Debugfln("Service [%s] closed", service.Name())
			}
		}

		// Close channels if strictly necessary (usually not needed if writers are stopped)

		s.logger.Info("Server shutdown complete.")
	})

	return finalErr
}
