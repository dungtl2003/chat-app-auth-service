package v1

import (
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/services"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Logout(a *services.AuthService) gin.HandlerFunc {
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
		accessToken, err := jwthandler.DecodeToken(a.JwtConfig.JwtSecret, accessTokenString)
		if err != nil {
			a.Logger.Debugf("DecodeToken(): %v", err)

			if err == jwt.ErrTokenExpired {
				c.JSON(401, "token expired")
			} else {
				c.JSON(401, "invalid token")
			}

			return
		}

		claims := accessToken.Claims.(*jwthandler.JWTClaim)
		deviceId, err := claims.GetDeviceId()
		if err != nil {
			a.Logger.Debugf("GetDeviceId(): %v", err)
			c.AbortWithStatus(500)
			return
		}

		// update device's token
		url := fmt.Sprintf("%s/%d/token", a.DeviceURL, deviceId)
		a.Logger.Debugf("sending DELETE request to %s", url)
		resp, err := a.Client.Delete(url, nil)
		if err != nil {
			a.Logger.Errorf("error when sending DELETE request: %v", err)
			c.AbortWithStatus(500)
			return
		}
		if resp.StatusCode != 200 {
			c.Status(resp.StatusCode)
			_, err = io.Copy(c.Writer, resp.Body)
			if err != nil {
				a.Logger.Errorf("Copy(): %v", err)
				c.AbortWithStatus(500)
			}

			return
		}

		c.SetCookie("refresh_token", "", 0, "/", a.DomainName, false, true)

		c.JSON(200, "logout successfully")
		return
	}
}
