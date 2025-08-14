package api

import (
	"dungtl2003/chat-app-auth-service/internal/constants"
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/model"
	u "dungtl2003/chat-app-auth-service/internal/services/user"
	"dungtl2003/chat-app-auth-service/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginRequestBody struct {
	Identifier string     `json:"identifier" validate:"required"`
	Password   string     `json:"password" validate:"required"`
	DeviceInfo types.Json `json:"device_info" validate:"required"`
}

type LoginResponseBody struct {
	AccessToken string          `json:"access_token"`
	SessionId   types.JsonInt64 `json:"session_id"`
	User        model.ChatUser  `json:"user"`
}

func Login(appCtx *context.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequestBody LoginRequestBody
		if err := c.ShouldBindJSON(&loginRequestBody); err != nil {
			appCtx.Logger.Errorfln("ShouldBindJSON(): %v", err)
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
		appCtx.Logger.Debugfln("Login request body: %#v", loginRequestBody)

		// get user information
		userGetAuthResponse, err := appCtx.UserService.GetUserAuth(c, loginRequestBody.Identifier)
		if err != nil {
			appCtx.Logger.Errorfln("GetUserAuth(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if userGetAuthResponse.Body.Error != "" {
			appCtx.Logger.Errorfln("GetUserAuth() returned error: %v", userGetAuthResponse.Body.Error)
			c.JSON(userGetAuthResponse.StatusCode, gin.H{"error": userGetAuthResponse.Body.Error})
			c.Abort()
			return
		}
		user := &userGetAuthResponse.Body.User

		// check password
		if !appCtx.PasswordManager.IsCorrectPassword(loginRequestBody.Password, user.Password) {
			appCtx.Logger.Errorfln("Invalid password")
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid password", "code": constants.INVALID_PASSWORD})
			c.Abort()
			return
		}

		// create session ID
		sessId, err := appCtx.IdGeneratorService.GenerateId(c)
		if err != nil {
			appCtx.Logger.Errorfln("GenerateId(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("Generated session ID: %d", sessId)

		// create tokens
		accessTokenStr, refreshTokenStr, err := helper.CreateNewTokenPair(appCtx.JwtConfig.JwtSecret, *user, appCtx.JwtConfig.ATDurationMs, appCtx.JwtConfig.RTDurationMs, sessId)
		if err != nil {
			appCtx.Logger.Errorfln("CreateNewTokPair(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("AT: %s", accessTokenStr)
		appCtx.Logger.Debugfln("RT: %s", refreshTokenStr)

		// create session
		expiresAt, err := helper.GetTokenExpirationStr(refreshTokenStr, appCtx.JwtConfig.JwtSecret)
		if err != nil {
			appCtx.Logger.Errorfln("GetTokExpStr(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		refreshTokenHash := helper.HashWithSHA256(refreshTokenStr)
		appCtx.Logger.Debugfln("Refresh token hash: %s", refreshTokenHash)
		sessionResponse, err := appCtx.UserService.CreateSession(c, user.Id.Int64(), u.SessionPostRequestBody{
			Id:               sessId,
			Version:          user.SessionVersion.Int64(),
			DeviceInfo:       loginRequestBody.DeviceInfo,
			RefreshTokenHash: refreshTokenHash,
			ExpiresAt:        expiresAt,
		})
		if err != nil {
			appCtx.Logger.Errorfln("CreateSession(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if sessionResponse.Body.Error != "" {
			appCtx.Logger.Errorfln("CreateSession() returned error: %v", sessionResponse.Body.Error)
			c.JSON(sessionResponse.StatusCode, gin.H{"error": sessionResponse.Body.Error})
			c.Abort()
			return
		}

		// add new session to user
		session := sessionResponse.Body.Session
		user.Sessions = append(user.Sessions, session)

		// security
		user.Password = ""

		helper.SetCookie(c, "refresh_token", refreshTokenStr, appCtx.DomainName, appCtx.JwtConfig.RTDurationMs)
		c.JSON(http.StatusOK, LoginResponseBody{
			AccessToken: accessTokenStr,
			SessionId:   types.NewJsonInt64(sessId),
			User:        *user,
		})
	}
}
