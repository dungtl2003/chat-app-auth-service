package api

import (
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/services"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func Logout(a *services.AuthService) gin.HandlerFunc {
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
			c.SetCookie("refresh_token", "", -1, "/", a.DomainName, false, true)
			c.Abort()
			return
		}

		claims := accessToken.Claims.(*jwthandler.JWTClaim)
		deviceId, err := claims.GetDeviceId()
		if err != nil {
			a.Logger.Errorf("GetDeviceId(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		userIdStr, err := claims.GetSubject()
		if err != nil {
			a.Logger.Errorf("GetSubject(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		userId, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			a.Logger.Errorf("strconv.ParseInt(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// update device's token
		url := fmt.Sprintf("%s/users/%d/devices/%d/token", a.UserServiceURL, userId, deviceId)
		a.Logger.Debugf("sending DELETE request to %s", url)
		resp, err := a.Client.Delete(url, nil)
		if err != nil {
			a.Logger.Errorf("error when sending DELETE request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.Status(resp.StatusCode)
			_, err = io.Copy(c.Writer, resp.Body)
			if err != nil {
				a.Logger.Errorf("Copy(): %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
			}

			return
		}

		c.SetCookie("refresh_token", "", -1, "/", a.DomainName, false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Logout successfully"})
		return
	}
}
