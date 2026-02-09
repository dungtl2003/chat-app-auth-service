package api

import (
	"dungtl2003/chat-app-auth-service/internal/cache"
	"dungtl2003/chat-app-auth-service/internal/config"
	"dungtl2003/chat-app-auth-service/internal/logging"
	"dungtl2003/chat-app-auth-service/internal/password"
	"dungtl2003/chat-app-auth-service/internal/services/idgen"
	"dungtl2003/chat-app-auth-service/internal/services/mailer"
	"dungtl2003/chat-app-auth-service/internal/services/user"
	"dungtl2003/chat-app-auth-service/internal/validate"
)

type HandlerDeps struct {
	Logger          *logging.LoggerWrapper
	Config          *config.Config
	Validator       *validate.Validator
	PasswordManager password.PasswordManager
	RedisClient     *cache.Client

	UserService        user.UserService
	MailerService      mailer.MailerService
	IdGeneratorService idgen.IdGeneratorService
}
