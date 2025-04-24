package api

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/services"
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SignUp(a *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// create account
		url := fmt.Sprintf("%s/users", a.UserServiceURL)
		a.Logger.Debugf("sending POST request to %s", url)
		resp, err := a.Client.Post(url, nil, c.Request.Body)
		if err != nil {
			a.Logger.Errorf("error when sending POST request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		if resp.StatusCode != http.StatusCreated {
			c.Status(resp.StatusCode)
			_, err = io.Copy(c.Writer, resp.Body)
			if err != nil {
				a.Logger.Errorf("Copy(): %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
			}

			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			a.Logger.Errorf("ReadAll(): %v", err)
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

		// when creating account, the user will have only one device
		deviceId := user.Devices[0].Id.Int64()

		// create tokens
		accessToken, err := jwthandler.CreateToken(a.JwtConfig.JwtSecret, user, a.JwtConfig.ATDurationMs, deviceId)
		if err != nil {
			a.Logger.Errorf("CreateToken(): error creating access token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		a.Logger.Debugf("AT: %s", accessToken)
		refreshToken, err := jwthandler.CreateToken(a.JwtConfig.JwtSecret, user, a.JwtConfig.RTDurationMs, deviceId)
		if err != nil {
			a.Logger.Errorf("CreateToken(): error creating refresh token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		a.Logger.Debugf("RT: %s", refreshToken)

		// update device's token
		url = fmt.Sprintf("%s/users/%d/devices/%d/token", a.UserServiceURL, user.Id.Int64(), deviceId)
		a.Logger.Debugf("sending PATCH request to %s", url)
		payload := fmt.Appendf(nil, `
		{
			"token": "%s"
		}
	`, refreshToken)
		resp, err = a.Client.Patch(url, nil, bytes.NewBuffer(payload))
		if err != nil {
			a.Logger.Errorf("error when sending PATCH request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if resp.StatusCode != http.StatusOK {
			a.Logger.Errorf("error PATCH request status: %d", resp.StatusCode)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// security
		for i := range user.Devices {
			user.Devices[i].RefreshToken = ""
		}
		user.SessionVersion = types.NewJsonInt64(-1)

		c.SetCookie("refresh_token", refreshToken, int(a.JwtConfig.RTDurationMs/1000), "/", a.DomainName, false, true)
		c.JSON(http.StatusOK, gin.H{
			"access_token": accessToken,
			"user":         user,
		})
		c.Abort()
	}
}
