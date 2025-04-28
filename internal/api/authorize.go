package api

import (
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/httpclient"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authorize(appCtx *context.AppContext) gin.HandlerFunc {
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims := accessToken.Claims.(*jwthandler.JWTClaim)
		userIdStr, err := claims.GetSubject()
		if err != nil {
			appCtx.Logger.Errorfln("GetSubject(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		userId, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			appCtx.Logger.Errorfln("ParseInt(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// get user information
		url := fmt.Sprintf("%s/users/%d", appCtx.UserServiceURL, userId)
		header := http.Header{
			"Authorization": []string{authHeader},
		}
		resp, err := appCtx.Client.Get(url, header)
		if err != nil {
			appCtx.Logger.Errorfln("error when sending GET request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if resp.StatusCode != http.StatusOK {
			appCtx.Logger.Errorfln("error GET request status: %d", resp.StatusCode)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		body, err := httpclient.ReadResponse(resp)
		if err != nil {
			appCtx.Logger.Errorfln("ReadResponse(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		var user model.ChatUser
		err = helper.ParseAsJson(body, &user)
		if err != nil {
			appCtx.Logger.Errorfln("ParseAsJson(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("response body from GET request: %#v", user)

		c.JSON(http.StatusOK, gin.H{"user": user})
		return
	}
}
