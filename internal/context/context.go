package context

import (
	"dungtl2003/chat-app-auth-service/internal/cache"
	"dungtl2003/chat-app-auth-service/internal/config"
	"dungtl2003/chat-app-auth-service/internal/logging"
	"dungtl2003/chat-app-auth-service/internal/password"
	"dungtl2003/chat-app-auth-service/internal/services"
	"dungtl2003/chat-app-auth-service/internal/services/idgen"
	"dungtl2003/chat-app-auth-service/internal/services/mailer"
	"dungtl2003/chat-app-auth-service/internal/services/user"
	"dungtl2003/chat-app-auth-service/internal/validate"
)

type AppContext struct {
	Env             config.Env
	Logger          *logging.LoggerWrapper
	Validator       *validate.Validator
	UserService     user.UserService
	MailerService   mailer.MailerService
	JwtConfig       config.JwtTokenConfig
	IdGenConfig     config.IdGeneratorConfig
	DomainName      string
	PasswordManager password.PasswordManager
	RedisClient     *cache.Client

	IdGeneratorService idgen.IdGeneratorService
	Services           []services.Service
}
