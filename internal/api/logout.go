package api

import (
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Logout(appCtx *context.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bearer <token>
		authHeader := c.GetHeader("Authorization")
		appCtx.Logger.Debugfln("authorization header: %s", authHeader)
		if authHeader == "" {
			appCtx.Logger.Debugfln("missing authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			appCtx.Logger.Debugfln("authorization header should have 2 parts")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header should have 2 parts"})
			c.Abort()
			return
		}

		if parts[0] != "Bearer" {
			appCtx.Logger.Debugfln("authorization header should start with `Bearer`")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header should start with `Bearer`"})
			c.Abort()
			return
		}

		accessTokenString := parts[1]
		accessToken, err := jwthandler.DecodeToken(appCtx.JwtConfig.JwtSecret, accessTokenString)
		if err != nil {
			appCtx.Logger.Debugfln("DecodeToken(): %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
			c.Abort()
			return
		}

		parsedToken, err := helper.ParseToken(accessToken)
		if err != nil {
			appCtx.Logger.Errorfln("ParseToken(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		userId := parsedToken.UserId
		sessId := parsedToken.SessionId

		// revoke session
		ok := helper.HandleRevokeSession(appCtx, c, userId, sessId)
		if ok {
			c.JSON(http.StatusOK, gin.H{"message": "Logout successfully"})
			c.Abort()
		}
	}
}
