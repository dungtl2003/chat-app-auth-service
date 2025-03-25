package v1

import (
	"dungtl2003/chat-app-auth-service/internal/helper"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context, logger *slog.Logger) {
	l := helper.NewLoggerWrapper(logger)
	l.Info("Login endpoint")
}
