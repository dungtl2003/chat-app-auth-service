package api

import (
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/model"
	u "dungtl2003/chat-app-auth-service/internal/services/user"
	"dungtl2003/chat-app-auth-service/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SignUpRequestBody struct {
	DeviceInfo types.Json `json:"device_info" validate:"required"`
	Email      string     `json:"email" validate:"required"`
	Username   string     `json:"username" validate:"required"`
	Password   string     `json:"password" validate:"required"`
	Role       string     `json:"role" validate:"required"`
}

type SignUpResponseBody struct {
	AccessToken string          `json:"access_token"`
	SessionId   types.JsonInt64 `json:"session_id"`
	User        model.ChatUser  `json:"user"`
}

func SignUp(appCtx *context.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// validate request body
		var signUpRequestBody SignUpRequestBody
		if err := c.ShouldBindJSON(&signUpRequestBody); err != nil {
			appCtx.Logger.Errorfln("ShouldBindJSON(): %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			c.Abort()
			return
		}
		if err := appCtx.Validator.Validate(signUpRequestBody); err != nil {
			appCtx.Logger.Errorfln("Error while validating request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("Sign up request body: %#v", signUpRequestBody)

		// create account
		userResponse, err := appCtx.UserService.CreateUser(c, u.UserPostRequestBody{
			Email:    signUpRequestBody.Email,
			Username: signUpRequestBody.Username,
			Password: signUpRequestBody.Password,
			Role:     model.UserRole(signUpRequestBody.Role),
		})
		if err != nil {
			appCtx.Logger.Errorfln("CreateUser(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if userResponse.Body.Error != "" {
			appCtx.Logger.Errorfln("CreateUser() returned error: %v", userResponse.Body.Error)
			c.JSON(userResponse.StatusCode, gin.H{"error": userResponse.Body.Error, "code": userResponse.Body.Code})
			c.Abort()
			return
		}
		user := userResponse.Body.User

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
		accessTokenStr, refreshTokenStr, err := helper.CreateNewTokenPair(appCtx.JwtConfig.JwtSecret, user, appCtx.JwtConfig.ATDurationMs, appCtx.JwtConfig.RTDurationMs, sessId)
		if err != nil {
			appCtx.Logger.Errorfln("CreateTokens(): error creating tokens: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("AT: %s", accessTokenStr)
		appCtx.Logger.Debugfln("RT: %s", refreshTokenStr)

		// create session
		expiresAt, err := helper.GetTokenExpirationStr(refreshTokenStr, appCtx.JwtConfig.JwtSecret)
		if err != nil {
			appCtx.Logger.Errorfln("GetTokenExpirationStr(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		refreshTokenHash := helper.HashWithSHA256(refreshTokenStr)
		appCtx.Logger.Debugfln("Refresh token hash: %s", refreshTokenHash)
		sessionResponse, err := appCtx.UserService.CreateSession(c, user.Id.Int64(), u.SessionPostRequestBody{
			Id:               sessId,
			Version:          user.SessionVersion.Int64(),
			DeviceInfo:       signUpRequestBody.DeviceInfo,
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
		session := sessionResponse.Body.Session

		// add new session to user
		user.Sessions = append(user.Sessions, session)

		// security
		user.Password = ""

		helper.SetCookie(c, "refresh_token", refreshTokenStr, appCtx.DomainName, appCtx.JwtConfig.RTDurationMs)
		signUpResponseBody := SignUpResponseBody{
			AccessToken: accessTokenStr,
			SessionId:   types.NewJsonInt64(sessId),
			User:        user,
		}
		appCtx.Logger.Debugfln("Response body: %#v", signUpResponseBody)
		c.JSON(http.StatusOK, signUpResponseBody)
		c.Abort()
	}
}
