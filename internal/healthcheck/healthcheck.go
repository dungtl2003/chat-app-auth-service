package healthcheck

import (
	"dungtl2003/chat-app-auth-service/internal/services"

	"github.com/gin-gonic/gin"
)

func HealthCheck(a *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		a.Logger.Info("server are running")
		c.JSON(200, "UP")
	}
}
