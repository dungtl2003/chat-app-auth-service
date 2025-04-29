package api

import (
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
			c.SetCookie("refresh_token", "", -1, "/", appCtx.DomainName, false, true)
			c.Abort()
			return
		}

		claims := accessToken.Claims.(*jwthandler.JWTClaim)
		sessId, err := claims.GetSessId()
		if err != nil {
			appCtx.Logger.Errorfln("GetSessId(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		userIdStr, err := claims.GetSubject()
		if err != nil {
			appCtx.Logger.Errorfln("GetSubject(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		userId, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			appCtx.Logger.Errorfln("strconv.ParseInt(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// revoke session
		url := helper.EncodeURLPath(fmt.Sprintf("%s/users/%d/sessions/%d/revoke", appCtx.UserServiceURL, userId, sessId))
		appCtx.Logger.Debugfln("sending POST request to %s", url)
		resp, err := appCtx.Client.Post(url, nil, nil)
		if err != nil {
			appCtx.Logger.Errorfln("error when sending POST request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.Status(resp.StatusCode)
			_, err = io.Copy(c.Writer, resp.Body)
			if err != nil {
				appCtx.Logger.Errorfln("Copy(): %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
			}
			return
		}

		c.SetCookie("refresh_token", "", -1, "/", appCtx.DomainName, false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Logout successfully"})
		return
	}
}
