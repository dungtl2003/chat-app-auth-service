package api

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/constants"
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/httpclient"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginRequestBody struct {
	Identifier string     `json:"identifier" validate:"required"`
	Password   string     `json:"password" validate:"required"`
	DeviceInfo types.JSON `json:"device_info" validate:"required"`
}

func Login(appCtx *context.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequestBody LoginRequestBody
		if err := c.ShouldBindJSON(&loginRequestBody); err != nil {
			appCtx.Logger.Debugfln("ShouldBindJSON(): %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			c.Abort()
			return
		}
		if err := appCtx.Validator.Validate(loginRequestBody); err != nil {
			appCtx.Logger.Debugfln("error while validating request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			c.Abort()
			return
		}

		appCtx.Logger.Debugfln("login request body: %#v", loginRequestBody)

		// get user information
		url := helper.EncodeURLPath(fmt.Sprintf(`%s/users/auth-info?identifier=%s`, appCtx.UserServiceURL, loginRequestBody.Identifier))
		appCtx.Logger.Debugfln("sending GET request to %s", url)
		resp, err := appCtx.Client.Get(url, nil)
		if err != nil {
			appCtx.Logger.Errorfln("error when sending GET request: %v", err)
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

		// check password
		if !appCtx.PasswordManager.IsCorrectPassword(loginRequestBody.Password, user.Password) {
			appCtx.Logger.Debug("invalid password")
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid password", "code": constants.INVALID_PASSWORD})
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

		// create session ID
		sessId, err := appCtx.IdGeneratorService.GenerateId()
		if err != nil {
			appCtx.Logger.Errorfln("GenerateId(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("generated session ID: %d", sessId)

		// create tokens
		accessTokenStr, err := jwthandler.CreateToken(appCtx.JwtConfig.JwtSecret, user, appCtx.JwtConfig.ATDurationMs, sessId)
		if err != nil {
			appCtx.Logger.Errorfln("CreateToken(): error creating access token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("AT: %s", accessTokenStr)
		refreshTokenStr, err := jwthandler.CreateToken(appCtx.JwtConfig.JwtSecret, user, appCtx.JwtConfig.RTDurationMs, sessId)
		if err != nil {
			appCtx.Logger.Errorfln("CreateToken(): error creating refresh token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("RT: %s", refreshTokenStr)
		refreshToken, err := jwthandler.DecodeToken(appCtx.JwtConfig.JwtSecret, refreshTokenStr)
		if err != nil {
			appCtx.Logger.Errorfln("DecodeToken(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		claims := refreshToken.Claims.(*jwthandler.JWTClaim)
		expiresAt, err := claims.GetExpirationTime()
		if err != nil {
			appCtx.Logger.Errorfln("GetExpiresAt(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// create session
		refreshTokenHash := helper.HashWithSHA256(refreshTokenStr)
		appCtx.Logger.Debugfln("refresh token hash: %s", refreshTokenHash)
		url = helper.EncodeURLPath(fmt.Sprintf("%s/users/%d/sessions", appCtx.UserServiceURL, user.Id.Int64()))
		payload := fmt.Appendf(nil, `
		{
			"id": "%d",
			"version": "%d",
			"device_info": %s,
			"refresh_token_hash": "%s",
			"expires_at": "%s"
		}`, sessId, user.SessionVersion.Int64(), loginRequestBody.DeviceInfo, refreshTokenHash, expiresAt.Format("2006-01-02T15:04:05.999Z"))
		resp, err = appCtx.Client.Post(url, nil, bytes.NewBuffer(payload))
		if err != nil {
			appCtx.Logger.Errorfln("error when sending POST request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if resp.StatusCode != http.StatusCreated {
			c.Status(resp.StatusCode)
			_, err = io.Copy(c.Writer, resp.Body)
			if err != nil {
				appCtx.Logger.Errorfln("Copy(): %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
			}
			return
		}
		body, err = httpclient.ReadResponse(resp)
		if err != nil {
			appCtx.Logger.Errorfln("ReadResponse(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		var session model.Session
		err = helper.ParseAsJson(body, &session)
		if err != nil {
			appCtx.Logger.Errorfln("ParseAsJson(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("response body from POST request: %#v", session)

		// add new session to user
		user.Sessions = append(user.Sessions, session)

		c.SetCookie("refresh_token", refreshTokenStr, int(appCtx.JwtConfig.RTDurationMs/1000), "/", appCtx.DomainName, false, true)
		c.JSON(http.StatusOK, gin.H{
			"access_token": accessTokenStr,
			"user":         user,
			"session_id":   sessId,
		})
	}
}
