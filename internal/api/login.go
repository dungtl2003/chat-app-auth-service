package api

import (
	"dungtl2003/chat-app-auth-service/internal/constants"
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/services"
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
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token,omitempty"`
	SessionId    types.JsonInt64 `json:"session_id"`
	User         model.ChatUser  `json:"user"`
}

func Login(appCtx *context.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := types.Response[LoginResponseBody]{}

		var loginRequestBody LoginRequestBody
		if err := c.ShouldBindJSON(&loginRequestBody); err != nil {
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

		appCtx.Logger.Debugfln("Login request body: %#v", loginRequestBody)

		if err := appCtx.Validator.Validate(loginRequestBody); err != nil {
			appCtx.Logger.Errorfln("error while validating request body: %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusBadRequest,
				Message: "Invalid request body",
				Errors:  []types.ErrorItem{{Message: "Invalid request body"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		// get user information
		userGetAuthResponse, err := appCtx.UserService.GetUserAuth(c, loginRequestBody.Identifier)
		if err != nil {
			appCtx.Logger.Errorfln("GetUserAuth(): %v", err)
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

		user := &userGetAuthResponse.User

		// check password
		if !appCtx.PasswordManager.IsCorrectPassword(loginRequestBody.Password, user.Password) {
			appCtx.Logger.Errorfln("Invalid password")
			response.Error = &types.ErrorBlock{
				Code:    http.StatusForbidden,
				Status:  constants.INVALID_PASSWORD,
				Message: "Invalid password",
				Errors:  []types.ErrorItem{{Message: "Invalid password", Reason: constants.INVALID_PASSWORD}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

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
		accessTokenStr, refreshTokenStr, err := helper.CreateNewTokenPair(appCtx.JwtConfig.JwtSecret, *user, appCtx.JwtConfig.ATDurationMs, appCtx.JwtConfig.RTDurationMs, sessId)
		if err != nil {
			appCtx.Logger.Errorfln("CreateNewTokPair(): %v", err)
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
			appCtx.Logger.Errorfln("GetTokExpStr(): %v", err)
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
			DeviceInfo:       loginRequestBody.DeviceInfo,
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

		// add new session to user
		session := sessionResponse.Session
		user.Sessions = append(user.Sessions, session)

		// security
		user.Password = ""

		native := helper.IsNativeClient(c)

		if !native {
			cfg := helper.GetCookieCfg(appCtx.DomainName, appCtx.Env)
			helper.SetCookieCfg(c, "refresh_token", refreshTokenStr, appCtx.JwtConfig.RTDurationMs, cfg)

			response.Data = &types.DataOrPage[LoginResponseBody]{
				Item: &LoginResponseBody{
					AccessToken: accessTokenStr,
					SessionId:   types.NewJsonInt64(sessId),
					User:        *user,
				},
			}

			appCtx.Logger.Debugfln("Detected non-native client, setting refresh_token cookie. Headers: %+v", c.Request.Header)
			appCtx.Logger.Debugfln("Login response: %+v", response.Data.Item)
			c.JSON(http.StatusOK, response)
			c.Abort()
			return
		}

		appCtx.Logger.Debugfln("Detected native client, returning refresh_token in response body. Headers: %+v", c.Request.Header)
		response.Data = &types.DataOrPage[LoginResponseBody]{
			Item: &LoginResponseBody{
				AccessToken:  accessTokenStr,
				RefreshToken: refreshTokenStr,
				SessionId:    types.NewJsonInt64(sessId),
				User:         *user,
			},
		}

		appCtx.Logger.Debugfln("Login response: %+v", response.Data.Item)
		c.JSON(http.StatusOK, response)
		c.Abort()
	}
}
