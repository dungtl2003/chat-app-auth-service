package api

import (
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/httpclient"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/services"
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authorize(a *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bearer <token>
		authHeader := c.GetHeader("Authorization")
		a.Logger.Debugf("authorization header: %s", authHeader)
		if authHeader == "" {
			a.Logger.Debugf("missing authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			a.Logger.Debugf("authorization header should have 2 parts")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header should have 2 parts"})
			c.Abort()
			return
		}

		if parts[0] != "Bearer" {
			a.Logger.Debugf("authorization header should start with `Bearer`")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header should start with `Bearer`"})
			c.Abort()
			return
		}

		accessTokenString := parts[1]
		accessToken, err := jwthandler.DecodeToken(a.JwtConfig.JwtSecret, accessTokenString)
		if err != nil {
			a.Logger.Debugf("DecodeToken(): %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims := accessToken.Claims.(*jwthandler.JWTClaim)
		userIdStr, err := claims.GetSubject()
		if err != nil {
			a.Logger.Errorf("GetSubject(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		userId, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			a.Logger.Errorf("ParseInt(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		url := fmt.Sprintf("%s/users/%d", a.UserServiceURL, userId)
		header := http.Header{
			"Authorization": []string{authHeader},
		}
		resp, err := a.Client.Get(url, header)
		if err != nil {
			a.Logger.Errorf("error when sending GET request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if resp.StatusCode != http.StatusOK {
			a.Logger.Errorf("error GET request status: %d", resp.StatusCode)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		body, err := httpclient.ReadResponse(resp)
		if err != nil {
			a.Logger.Errorf("ReadResponse(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		var user model.ChatUser
		err = helper.ParseAsJson(body, &user)
		if err != nil {
			a.Logger.Errorf("ParseAsJson(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		a.Logger.Debugf("response body from GET request: %#v", user)

		// security
		for i := range user.Devices {
			user.Devices[i].RefreshToken = ""
		}
		user.SessionVersion = types.NewJsonInt64(-1)

		c.JSON(http.StatusOK, gin.H{"user": user})
		return
	}
}
