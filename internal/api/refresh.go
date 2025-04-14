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
	"net/http"

	"github.com/gin-gonic/gin"
)

// Refresh logic flow:
// 1. If the refresh token is missing, return 401.
// 2. If the refresh token is invalid or expired, delete the cookie and return 401.
// 3. If the session version in the refresh token is less than the session version
// in the database, delete the cookie and return 401 (session version is the amount
// of times server detects that the user uses the same refresh token more than once.
// If the session version is less than the one in the database, it means that the
// user still uses the old valid refresh token, which is a sign of an attack).
// 4. If the refresh token is in the database, create a new access token and refresh
// token, update the refresh token in the database, and set the new refresh token
// in the cookie.
// 5. Else, delete all refresh tokens in the database and update the session version
// in the database. This is a sign of an attack, so return 401.
func Refresh(a *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		refreshTokenStr, err := c.Cookie("refresh_token")
		if err != nil {
			if err == http.ErrNoCookie {
				a.Logger.Debug("missing refresh token")
				c.JSON(401, "missing refresh token")
				return
			} else {
				a.Logger.Errorf("Cookie(): %v", err)
				c.AbortWithStatus(500)
				return
			}
		}

		a.Logger.Debugf("decoding token %s", refreshTokenStr)
		refreshToken, err := jwthandler.DecodeToken(a.JwtConfig.JwtSecret, refreshTokenStr)
		if err != nil {
			a.Logger.Debugf("DecodeToken(): %v", err)
			c.SetCookie("refresh_token", "", -1, "/", a.DomainName, false, true)
			c.JSON(401, "invalid token")
			return
		}

		claims := refreshToken.Claims.(*jwthandler.JWTClaim)
		sessionVersion, err := claims.GetSessVersion()
		if err != nil {
			a.Logger.Errorf("GetSessVersion(): %v", err)
			c.AbortWithStatus(500)
			return
		}

		username, err := claims.GetUsername()
		if err != nil {
			a.Logger.Errorf("GetSessVersion(): %v", err)
			c.AbortWithStatus(500)
			return
		}

		deviceId, err := claims.GetDeviceId()
		if err != nil {
			a.Logger.Errorf("GetDeviceId(): %v", err)
			c.AbortWithStatus(500)
			return
		}

		url := fmt.Sprintf("%s/users/auth-info?identifier=%s", a.UserServiceURL, username)
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

		if sessionVersion < user.SessionVersion.Int64() {
			c.SetCookie("refresh_token", "", -1, "/", a.DomainName, false, true)
			a.Logger.Debugf("invalid session version (expected: %d, got: %d)", user.SessionVersion.Int64(), sessionVersion)
			c.JSON(401, "invalid session version")
			return
		}

		has := false
		for _, device := range user.Devices {
			if device.RefreshToken == refreshTokenStr {
				if deviceId != device.Id.Int64() {
					a.Logger.Errorf("the token must be saved in the correct device (expected: %d, actual: %d). It seems like the create token logic is wrong!!!", device.Id.Int64(), deviceId)
					c.JSON(500, "it seems like the create token logic is wrong!!!")
					return
				}
				has = true
				break
			}
		}

		// HOW CAN YOU USE THIS TOKEN ??!!!!!!
		if !has {
			// maybe stolen by someone. Regardless, DELETE ALL!!!!!!!!!
			url = fmt.Sprintf("%s/devices/tokens?user_id=%d", a.UserServiceURL, user.Id.Int64())
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

			// update session version
			url = fmt.Sprintf("%s/users/%d/session", a.UserServiceURL, user.Id.Int64())
			payload := fmt.Appendf(nil, `
				{
					"session_version": "increment"
				}
			`)
			a.Logger.Debugf("sending PATCH request to %s with payload: %s", url, helper.StripWS(string(payload)))
			resp, err = a.Client.Patch(url, nil, bytes.NewBuffer(payload))
			if err != nil {
				a.Logger.Errorf("error when sending PATCH request: %v", err)
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

			a.Logger.Debugf("the account might be attacked")
			c.JSON(401, "the account might be attacked")
			return
		}

		// create new tokens
		newAccessToken, err := jwthandler.CreateToken(a.JwtConfig.JwtSecret, user, a.JwtConfig.ATDurationMs, deviceId)
		if err != nil {
			a.Logger.Errorf("CreateToken(): error creating access token: %v", err)
			c.AbortWithStatus(500)
			return
		}
		a.Logger.Debugf("AT: %s", newAccessToken)

		newRefreshToken, err := jwthandler.CreateToken(a.JwtConfig.JwtSecret, user, a.JwtConfig.RTDurationMs, deviceId)
		if err != nil {
			a.Logger.Errorf("CreateToken(): error creating refresh token: %v", err)
			c.AbortWithStatus(500)
			return
		}
		a.Logger.Debugf("RT: %s", newRefreshToken)

		// update device's token
		url = fmt.Sprintf("%s/devices/%d/token", a.UserServiceURL, deviceId)
		payload := fmt.Appendf(nil, `
			{
				"token": "%s",
				"user_id": "%d"
			}
		`, newRefreshToken, user.Id.Int64())
		a.Logger.Debugf("sending PATCH request to %s with payload: %s", url, helper.StripWS(string(payload)))
		resp, err = a.Client.Patch(url, nil, bytes.NewBuffer(payload))
		if err != nil {
			a.Logger.Errorf("error when sending PATCH request: %v", err)
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

		// set cookie
		c.SetCookie("refresh_token", newRefreshToken, int(a.JwtConfig.RTDurationMs/1000), "/", a.DomainName, false, true)
		c.JSON(200, fmt.Sprintf("access_token: %s", newAccessToken))
	}
}
