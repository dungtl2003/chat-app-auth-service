package api

import (
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthorizeResponseBody struct {
	User model.ChatUser `json:"user"`
}

func Authorize(appCtx *context.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bearer <token>
		authHeader := c.GetHeader("Authorization")
		appCtx.Logger.Debugfln("Authorization header: %s", authHeader)
		if authHeader == "" {
			appCtx.Logger.Debugfln("Missing authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			appCtx.Logger.Debugfln("Authorization header should have 2 parts")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header should have 2 parts"})
			c.Abort()
			return
		}

		if parts[0] != "Bearer" {
			appCtx.Logger.Debugfln("Authorization header should start with `Bearer`")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header should start with `Bearer`"})
			c.Abort()
			return
		}

		accessTokenString := parts[1]
		accessToken, err := jwthandler.DecodeToken(appCtx.JwtConfig.JwtSecret, accessTokenString)
		if err != nil {
			appCtx.Logger.Debugfln("DecodeToken(): %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// make sure token is not internal token
		audiences, err := accessToken.Claims.GetAudience()
		if err != nil {
			appCtx.Logger.Debugfln("GetAudience(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if len(audiences) == 0 {
			appCtx.Logger.Debugfln("token has no audience")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		if audiences[0] == jwthandler.INTERNAL_AUDIENCE {
			appCtx.Logger.Debugfln("token is internal token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
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

		// create internal token
		internalToken, err := jwthandler.CreateInternalToken(appCtx.JwtConfig.JwtSecret, appCtx.JwtConfig.ATDurationMs, userId)
		if err != nil {
			appCtx.Logger.Errorfln("CreateInternalToken(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// get user information
		userResponse, err := appCtx.UserService.GetUserById(c, userId)
		if err != nil {
			appCtx.Logger.Errorfln("GetUserById(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if userResponse.Body.Error != "" {
			appCtx.Logger.Errorfln("GetUserById() returned error: %v", userResponse.Body.Error)
			c.JSON(userResponse.StatusCode, gin.H{"error": userResponse.Body.Error})
			c.Abort()
			return
		}

		user := userResponse.Body.User
		// sanitize user data
		if user.Password != "" {
			appCtx.Logger.Warnfln("User password should not be returned in response")
			user.Password = ""
		}
		// replace user token with internal token for internal service
		c.Header("Authorization", fmt.Sprintf("Bearer %s", internalToken))
		c.JSON(http.StatusOK, AuthorizeResponseBody{
			User: user,
		})
	}
}
