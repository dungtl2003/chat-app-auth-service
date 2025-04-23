package api

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/httpclient"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/services"
	"fmt"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LoginRequestBody struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
	DeviceId   string `json:"device_id" validate:"required"`
}

func Login(a *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequestBody LoginRequestBody

		if err := c.ShouldBindJSON(&loginRequestBody); err != nil {
			a.Logger.Errorf("ShouldBindJSON(): %v", err)
			c.JSON(400, "invalid body")
			return
		}

		deviceId, err := strconv.ParseInt(loginRequestBody.DeviceId, 10, 64)
		if err != nil {
			a.Logger.Errorf("invalid device ID: %s", loginRequestBody.DeviceId)
			c.JSON(400, "invalid device ID")
			return
		}

		a.Logger.Debugf("login request body: %#v", loginRequestBody)

		// get user information
		url := fmt.Sprintf("%s/users/auth-info?identifier=%s", a.UserServiceURL, loginRequestBody.Identifier)
		a.Logger.Debugf("sending GET request to %s", url)
		resp, err := a.Client.Get(url, nil)
		if err != nil {
			a.Logger.Errorf("error when sending GET request: %v", err)
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

		body, err := httpclient.ReadResponse(resp)
		if err != nil {
			a.Logger.Errorf("ReadResponse(): %v", err)
			c.AbortWithStatus(500)
			return
		}

		var user model.ChatUser
		err = helper.ParseAsJson(body, &user)
		if err != nil {
			a.Logger.Errorf("ParseAsJson(): %v", err)
			c.AbortWithStatus(500)
			return
		}
		a.Logger.Debugf("response body from GET request: %#v", user)

		// check password
		if !a.PasswordManager.IsCorrectPassword(loginRequestBody.Password, user.Password) {
			a.Logger.Debug("invalid password")
			c.JSON(403, "invalid payload")
			return
		}

		has := false
		for _, device := range user.Devices {
			if device.Id.Int64() == deviceId {
				has = true
				break
			}
		}

		if !has {
			a.Logger.Debugf("user with ID %d does not have device with ID %d", user.Id.Int64(), deviceId)
			c.JSON(403, "invalid payload")
			return
		}

		// create tokens
		accessToken, err := jwthandler.CreateToken(a.JwtConfig.JwtSecret, user, a.JwtConfig.ATDurationMs, deviceId)
		if err != nil {
			a.Logger.Errorf("CreateToken(): error creating access token: %v", err)
			c.AbortWithStatus(500)
			return
		}
		a.Logger.Debugf("AT: %s", accessToken)
		refreshToken, err := jwthandler.CreateToken(a.JwtConfig.JwtSecret, user, a.JwtConfig.RTDurationMs, deviceId)
		if err != nil {
			a.Logger.Errorf("CreateToken(): error creating refresh token: %v", err)
			c.AbortWithStatus(500)
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
			c.AbortWithStatus(500)
			return
		}
		if resp.StatusCode != 200 {
			a.Logger.Errorf("error PATCH request status: %d", resp.StatusCode)
			c.AbortWithStatus(500)
			return
		}

		c.SetCookie("refresh_token", refreshToken, int(a.JwtConfig.RTDurationMs/1000), "/", a.DomainName, false, true)
		c.JSON(200, fmt.Sprintf("access_token: %s", accessToken))
	}
}
