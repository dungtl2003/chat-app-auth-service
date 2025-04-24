package api

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/httpclient"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/services"
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LoginDeviceRequestBody struct {
	Id         string `json:"id"`
	DeviceName string `json:"device_name"`
	DeviceType string `json:"device_type"`
	Os         string `json:"os"`
}

type LoginRequestBody struct {
	Identifier string                 `json:"identifier" validate:"required"`
	Password   string                 `json:"password" validate:"required"`
	Device     LoginDeviceRequestBody `json:"device" validate:"required"`
}

const (
	INVALID_DEVICE_ID = "Invalid device ID"
)

func Login(a *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequestBody LoginRequestBody

		if err := c.ShouldBindJSON(&loginRequestBody); err != nil {
			a.Logger.Debugf("ShouldBindJSON(): %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			c.Abort()
			return
		}

		a.Logger.Debugf("login request body: %#v", loginRequestBody)

		var deviceId int64
		deviceId = -1
		if loginRequestBody.Device.Id != "" {
			var err error
			deviceId, err = strconv.ParseInt(loginRequestBody.Device.Id, 10, 64)
			if err != nil {
				a.Logger.Debugf("invalid device ID: %s", loginRequestBody.Device.Id)
				c.JSON(http.StatusBadRequest, gin.H{"error": INVALID_DEVICE_ID})
				c.Abort()
				return
			}
		}

		// get user information
		url := fmt.Sprintf("%s/users/auth-info?identifier=%s", a.UserServiceURL, loginRequestBody.Identifier)
		a.Logger.Debugf("sending GET request to %s", url)
		resp, err := a.Client.Get(url, nil)
		if err != nil {
			a.Logger.Errorf("error when sending GET request: %v", err)
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

		// check password
		if !a.PasswordManager.IsCorrectPassword(loginRequestBody.Password, user.Password) {
			a.Logger.Debug("invalid password")
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid password"})
			c.Abort()
			return
		}

		// we need to add a new device
		if deviceId == -1 {
			url = fmt.Sprintf("%s/users/%d/devices", a.UserServiceURL, user.Id.Int64())
			payload := fmt.Appendf(nil, `
				{
					"device_name": "%s",
					"device_type": "%s",
					"os": "%s"
				}`, loginRequestBody.Device.DeviceName, loginRequestBody.Device.DeviceType, loginRequestBody.Device.Os)
			a.Logger.Debugf("sending POST request to %s", url)
			resp, err = a.Client.Post(url, nil, bytes.NewBuffer(payload))
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

			var device model.Device
			err = helper.ParseAsJson(body, &device)
			if err != nil {
				a.Logger.Errorf("ParseAsJson(): %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
				return
			}

			a.Logger.Debugf("response body from POST request: %#v", device)

			deviceId = device.Id.Int64()
			user.Devices = append(user.Devices, device)
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": INVALID_DEVICE_ID})
			c.Abort()
			return
		}

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
			"access_token":      accessToken,
			"user":              user,
			"current_device_id": deviceId,
		})
	}
}
