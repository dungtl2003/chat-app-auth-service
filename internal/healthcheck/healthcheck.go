package healthcheck

import (
	"dungtl2003/chat-app-auth-service/internal/helper"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context, l *slog.Logger) {
	logger := helper.NewLoggerWrapper(l)

	logger.Info("server are running")
	c.JSON(200, "UP")
}
