package context

import (
	"dungtl2003/chat-app-auth-service/internal/config"
	"dungtl2003/chat-app-auth-service/internal/logging"
	"dungtl2003/chat-app-auth-service/internal/password"
	"dungtl2003/chat-app-auth-service/internal/services"
	"dungtl2003/chat-app-auth-service/internal/services/idgen"
	"dungtl2003/chat-app-auth-service/internal/services/user"
	"dungtl2003/chat-app-auth-service/internal/validate"
)

type AppContext struct {
	Logger          *logging.LoggerWrapper
	Validator       *validate.Validator
	UserService     user.UserService
	JwtConfig       config.JwtTokenConfig
	DomainName      string
	PasswordManager password.PasswordManager

	IdGeneratorService idgen.IdGeneratorService
	Services           []services.Service
}
