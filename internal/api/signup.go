package api

import (
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/services"
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
	AccessToken string `json:"access_token"`
	// This is for native clients that cannot store HttpOnly cookies
	RefreshToken string          `json:"refresh_token,omitempty"`
	SessionId    types.JsonInt64 `json:"session_id"`
	User         model.ChatUser  `json:"user"`
}

func SignUp(appCtx *context.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := types.Response[SignUpResponseBody]{}

		// validate request body
		var signUpRequestBody SignUpRequestBody
		if err := c.ShouldBindJSON(&signUpRequestBody); err != nil {
			appCtx.Logger.Errorfln("ShouldBindJSON(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusBadRequest,
				Message: "Invalid payload",
				Errors: []types.ErrorItem{
					{Message: "Invalid payload"},
				},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		appCtx.Logger.Debugfln("Sign up request body: %#v", signUpRequestBody)

		if err := appCtx.Validator.Validate(signUpRequestBody); err != nil {
			appCtx.Logger.Errorfln("Error while validating request body: %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusBadRequest,
				Message: "Invalid request body",
				Errors:  []types.ErrorItem{{Message: "Invalid request body"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		// create account
		userResponse, err := appCtx.UserService.CreateUser(c, u.UserPostRequest{
			Email:    signUpRequestBody.Email,
			Username: signUpRequestBody.Username,
			Password: signUpRequestBody.Password,
			Role:     model.UserRole(signUpRequestBody.Role),
		})
		if err != nil {
			appCtx.Logger.Errorfln("CreateUser(): %v", err)
			switch err := err.(type) {
			case services.BadResponseError:
				response.Error = &err.ErrBlock
			default:
				response.Error = &types.ErrorBlock{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Errors:  []types.ErrorItem{{Message: "Internal server error"}},
				}
			}

			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		user := userResponse.User

		// create session ID
		sessId, err := appCtx.IdGeneratorService.GenerateId(c)
		if err != nil {
			appCtx.Logger.Errorfln("GenerateId(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("Generated session ID: %d", sessId)

		// create tokens
		accessTokenStr, refreshTokenStr, err := helper.CreateNewTokenPair(appCtx.JwtConfig.JwtSecret, user, appCtx.JwtConfig.ATDurationMs, appCtx.JwtConfig.RTDurationMs, sessId, appCtx.IdGenConfig.Epoch)
		if err != nil {
			appCtx.Logger.Errorfln("CreateTokens(): error creating tokens: %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("AT: %s", accessTokenStr)
		appCtx.Logger.Debugfln("RT: %s", refreshTokenStr)

		// create session
		expiresAt, err := helper.GetTokenExpirationStr(refreshTokenStr, appCtx.JwtConfig.JwtSecret)
		if err != nil {
			appCtx.Logger.Errorfln("GetTokenExpirationStr(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}
		refreshTokenHash := helper.HashWithSHA256(refreshTokenStr)
		appCtx.Logger.Debugfln("Refresh token hash: %s", refreshTokenHash)
		sessionResponse, err := appCtx.UserService.CreateSession(c, user.Id.Int64(), u.SessionPostRequest{
			Id:               sessId,
			Version:          user.SessionVersion.Int64(),
			DeviceInfo:       signUpRequestBody.DeviceInfo,
			RefreshTokenHash: refreshTokenHash,
			ExpiresAt:        expiresAt,
		})
		if err != nil {
			appCtx.Logger.Errorfln("CreateSession(): %v", err)
			switch err := err.(type) {
			case services.BadResponseError:
				response.Error = &err.ErrBlock
			default:
				response.Error = &types.ErrorBlock{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Errors:  []types.ErrorItem{{Message: "Internal server error"}},
				}
			}

			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		session := sessionResponse.Session

		// add new session to user
		user.Sessions = append(user.Sessions, session)

		// security
		user.Password = ""

		native := helper.IsNativeClient(c)
		if !native {
			cfg := helper.GetCookieCfg(appCtx.DomainName, appCtx.Env)
			helper.SetCookieCfg(c, "refresh_token", refreshTokenStr, appCtx.JwtConfig.RTDurationMs, cfg)
			response.Data = &types.DataOrPage[SignUpResponseBody]{
				Item: &SignUpResponseBody{
					AccessToken: accessTokenStr,
					SessionId:   types.NewJsonInt64(sessId),
					User:        user,
				},
			}
			appCtx.Logger.Debugfln("Response body: %#v", response.Data.Item)
			c.JSON(http.StatusOK, response)
			c.Abort()
			return
		}

		response.Data = &types.DataOrPage[SignUpResponseBody]{
			Item: &SignUpResponseBody{
				AccessToken:  accessTokenStr,
				RefreshToken: refreshTokenStr,
				SessionId:    types.NewJsonInt64(sessId),
				User:         user,
			},
		}
		appCtx.Logger.Debugfln("Response body: %#v", response.Data.Item)
		c.JSON(http.StatusOK, response)
		c.Abort()
	}
}
