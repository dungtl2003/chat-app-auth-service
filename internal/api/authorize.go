package api

import (
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authorize(a *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bearer <token>
		authHeader := c.GetHeader("Authorization")
		a.Logger.Debugf("authorization header: %s", authHeader)
		if authHeader == "" {
			a.Logger.Errorf("missing authorization header")
			c.JSON(401, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			a.Logger.Errorf("authorization header should have 2 parts")
			c.JSON(401, "authorization header should have 2 parts")
			return
		}

		if parts[0] != "Bearer" {
			a.Logger.Errorf("authorization header should start with `Bearer`")
			c.JSON(401, "authorization header should start with `Bearer`")
			return
		}

		accessTokenString := parts[1]
		_, err := jwthandler.DecodeToken(a.JwtConfig.JwtSecret, accessTokenString)
		if err != nil {
			a.Logger.Debugf("DecodeToken(): %v", err)
			c.JSON(401, "invalid token")
			return
		}

		c.JSON(200, "authorized")
		return
	}
}
