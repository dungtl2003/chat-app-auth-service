package context

import (
	"dungtl2003/chat-app-auth-service/internal/config"
	"dungtl2003/chat-app-auth-service/internal/httpclient"
	"dungtl2003/chat-app-auth-service/internal/logging"
	"dungtl2003/chat-app-auth-service/internal/password"
	"dungtl2003/chat-app-auth-service/internal/services"
	"dungtl2003/chat-app-auth-service/internal/services/snowflake"
	"dungtl2003/chat-app-auth-service/internal/validate"
)

type AppContext struct {
	Logger          *logging.LoggerWrapper
	Validator       *validate.Validator
	Client          *httpclient.HttpClient
	UserServiceURL  string
	JwtConfig       config.JwtTokenConfig
	DomainName      string
	PasswordManager password.PasswordManager

	IdGeneratorService *snowflake.IdGeneratorService
	Services           []services.Service
}

func NewAppCtx(logger *logging.LoggerWrapper, validator *validate.Validator, client *httpclient.HttpClient, userServiceURL string, jwtConfig config.JwtTokenConfig, domainName string, pm password.PasswordManager, idGeneratorService *snowflake.IdGeneratorService, services []services.Service) *AppContext {
	return &AppContext{
		IdGeneratorService: idGeneratorService,
		Logger:             logger,
		Validator:          validator,
		Client:             client,
		UserServiceURL:     userServiceURL,
		JwtConfig:          jwtConfig,
		DomainName:         domainName,
		PasswordManager:    pm,
		Services:           services,
	}
}
