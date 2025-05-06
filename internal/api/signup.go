package api

import (
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/model"
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

func SignUp(appCtx *context.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// validate request body
		var signUpRequestBody SignUpRequestBody
		if err := c.ShouldBindJSON(&signUpRequestBody); err != nil {
			appCtx.Logger.Debugfln("ShouldBindJSON(): %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			c.Abort()
			return
		}
		if err := appCtx.Validator.Validate(signUpRequestBody); err != nil {
			appCtx.Logger.Debugfln("error while validating request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("sign up request body: %#v", signUpRequestBody)

		// create account
		user := helper.HandleCreateUser(appCtx, c, helper.UserPostPayload{
			Email:    signUpRequestBody.Email,
			Username: signUpRequestBody.Username,
			Password: signUpRequestBody.Password,
			Role:     model.UserRole(signUpRequestBody.Role),
		})
		if user == nil {
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
		accessTokenStr, refreshTokenStr, err := helper.CreateNewTokenPair(appCtx.JwtConfig.JwtSecret, *user, appCtx.JwtConfig.ATDurationMs, appCtx.JwtConfig.RTDurationMs, sessId)
		if err != nil {
			appCtx.Logger.Errorfln("CreateTokens(): error creating tokens: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("AT: %s", accessTokenStr)
		appCtx.Logger.Debugfln("RT: %s", refreshTokenStr)

		// create session
		expiresAt, err := helper.GetTokExpStr(refreshTokenStr, appCtx.JwtConfig.JwtSecret)
		if err != nil {
			appCtx.Logger.Errorfln("GetTokExpStr(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		refreshTokenHash := helper.HashWithSHA256(refreshTokenStr)
		appCtx.Logger.Debugfln("refresh token hash: %s", refreshTokenHash)
		session := helper.HandleCreateSession(appCtx, c, helper.SessionPostPayload{
			Id:               sessId,
			Version:          user.SessionVersion.Int64(),
			DeviceInfo:       signUpRequestBody.DeviceInfo,
			RefreshTokenHash: refreshTokenHash,
			ExpiresAt:        expiresAt,
		}, user.Id.Int64())
		if session == nil {
			return
		}

		// add new session to user
		user.Sessions = append(user.Sessions, *session)

		// security
		user.Password = ""

		helper.SetCookie(c, "refresh_token", refreshTokenStr, appCtx.DomainName, appCtx.JwtConfig.RTDurationMs)
		c.JSON(http.StatusOK, gin.H{
			"access_token": accessTokenStr,
			"user":         *user,
			"session_id":   sessId,
		})
		c.Abort()
	}
}
