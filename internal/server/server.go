package server

import (
	"context"
	v1 "dungtl2003/chat-app-auth-service/internal/api/v1"
	"dungtl2003/chat-app-auth-service/internal/config"
	"dungtl2003/chat-app-auth-service/internal/healthcheck"
	"dungtl2003/chat-app-auth-service/internal/router"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthServer struct {
	config *config.Config
	srv    *http.Server
}

// New creates a new AuthServer instance. The function will load the configuration
// and set up all necessary components. Call Run() to start the server. This function
// will exit the program if there is an error when creating.
func New() *AuthServer {
	log.Println("Loading configuration")
	config := config.New()
	config.Load()
	log.Printf("Configuration: %s\n", config)
	logger := config.LogConfig.Logger

	// log.Println("Creating validator")
	// validator := helper.NewValidator()

	log.Println("Creating router")
	handlers := []router.Handler{
		{
			Method: router.GET,
			Path:   "/healthcheck",
			H: func(c *gin.Context) {
				healthcheck.HealthCheck(c, logger)
			},
		},
		{
			Method: router.GET,
			Path:   "/api/v1/authorize",
			H: func(c *gin.Context) {
				v1.Authorize(c, logger)
			},
		},
		{
			Method: router.GET,
			Path:   "/api/v1/refresh",
			H: func(c *gin.Context) {
				v1.Refresh(c, logger)
			},
		},
		{
			Method: router.GET,
			Path:   "/api/v1/logout",
			H: func(c *gin.Context) {
				v1.Logout(c, logger)
			},
		},

		{
			Method: router.POST,
			Path:   "/api/v1/login",
			H: func(c *gin.Context) {
				v1.Login(c, logger)
			},
		},
	}
	r := router.New(logger, handlers...)

	log.Println("Creating server")
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", config.ServerPort),
		Handler: r,
	}

	return &AuthServer{
		config: config,
		srv:    srv,
	}
}

// Run starts the server. The function will start the server and listen for signals
// to shut down the server. The function will exit the program if there is an error
// when starting the server. Call Close() to shut down the server.
func (s *AuthServer) Run() {
	logger := s.config.LogConfig.Logger

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
func (s *AuthServer) Close() {
	logger := s.config.LogConfig.Logger
	logger.Info("shutting down server")

	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
		if err != nil {
			os.Exit(1)
		}

		os.Exit(0)
	}()

	if err = s.srv.Shutdown(ctx); err != nil {
		logger.Error("error when shutting down server", "error", err)
	} else {
		logger.Info("server shut down")
	}
}
