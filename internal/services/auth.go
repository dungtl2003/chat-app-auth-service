package services

import (
	"dungtl2003/chat-app-auth-service/internal/config"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/httpclient"
	"dungtl2003/chat-app-auth-service/internal/password"
	"log/slog"
)

type AuthService struct {
	Logger          *helper.LoggerWrapper
	Validator       *helper.Validator
	Client          *httpclient.HttpClient
	UserServiceURL  string
	JwtConfig       config.JwtTokenConfig
	DomainName      string
	PasswordManager password.PasswordManager
}

func NewAuthService(logger *slog.Logger, validator *helper.Validator, client *httpclient.HttpClient, userServiceURL string, jwtConfig config.JwtTokenConfig, domainName string, pm password.PasswordManager) *AuthService {
	l := helper.NewLoggerWrapper(logger)
	return &AuthService{
		Logger:          &l,
		Validator:       validator,
		Client:          client,
		UserServiceURL:  userServiceURL,
		JwtConfig:       jwtConfig,
		DomainName:      domainName,
		PasswordManager: pm,
	}
}
