package api

import (
	"dungtl2003/chat-app-auth-service/internal/constants"
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

func Login(handlerDeps *HandlerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := types.Response[LoginResponseBody]{}

		var loginRequestBody LoginRequestBody
		if err := c.ShouldBindJSON(&loginRequestBody); err != nil {
			handlerDeps.Logger.Errorfln("ShouldBindJSON(): %v", err)
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

		handlerDeps.Logger.Debugfln("Login request body: %#v", loginRequestBody)

		if err := handlerDeps.Validator.Validate(loginRequestBody); err != nil {
			handlerDeps.Logger.Errorfln("error while validating request body: %v", err)
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
		userGetAuthResponse, err := handlerDeps.UserService.GetUserAuth(c, loginRequestBody.Identifier)
		if err != nil {
			handlerDeps.Logger.Errorfln("GetUserAuth(): %v", err)
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
		if !handlerDeps.PasswordManager.IsCorrectPassword(loginRequestBody.Password, user.Password) {
			handlerDeps.Logger.Errorfln("Invalid password")
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
		sessId, err := handlerDeps.IdGeneratorService.GenerateId(c)
		if err != nil {
			handlerDeps.Logger.Errorfln("GenerateId(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}
		handlerDeps.Logger.Debugfln("Generated session ID: %d", sessId)

		// create tokens
		accessTokenStr, refreshTokenStr, err := helper.CreateNewTokenPair(
			handlerDeps.Config.JwtTokenConfig.JwtSecret,
			*user,
			handlerDeps.Config.JwtTokenConfig.ATDurationMs,
			handlerDeps.Config.JwtTokenConfig.RTDurationMs,
			sessId,
			handlerDeps.Config.IdGeneratorConfig.Epoch,
		)
		if err != nil {
			handlerDeps.Logger.Errorfln("CreateNewTokPair(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}
		handlerDeps.Logger.Debugfln("AT: %s", accessTokenStr)
		handlerDeps.Logger.Debugfln("RT: %s", refreshTokenStr)

		// create session
		expiresAt, err := helper.GetTokenExpirationStr(refreshTokenStr, handlerDeps.Config.JwtTokenConfig.JwtSecret)
		if err != nil {
			handlerDeps.Logger.Errorfln("GetTokExpStr(): %v", err)
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
		handlerDeps.Logger.Debugfln("Refresh token hash: %s", refreshTokenHash)
		sessionResponse, err := handlerDeps.UserService.CreateSession(c, user.Id.Int64(), u.SessionPostRequest{
			Id:               sessId,
			Version:          user.SessionVersion.Int64(),
			DeviceInfo:       loginRequestBody.DeviceInfo,
			RefreshTokenHash: refreshTokenHash,
			ExpiresAt:        expiresAt,
		})
		if err != nil {
			handlerDeps.Logger.Errorfln("CreateSession(): %v", err)
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
			cfg := helper.GetCookieCfg(handlerDeps.Config.DomainName, handlerDeps.Config.Env)
			helper.SetCookieCfg(c, "refresh_token", refreshTokenStr, handlerDeps.Config.JwtTokenConfig.RTDurationMs, cfg)

			response.Data = &types.DataOrPage[LoginResponseBody]{
				Item: &LoginResponseBody{
					AccessToken: accessTokenStr,
					SessionId:   types.NewJsonInt64(sessId),
					User:        *user,
				},
			}

			handlerDeps.Logger.Debugfln("Detected non-native client, setting refresh_token cookie. Headers: %+v", c.Request.Header)
			handlerDeps.Logger.Debugfln("Login response: %+v", response.Data.Item)
			c.JSON(http.StatusOK, response)
			c.Abort()
			return
		}

		handlerDeps.Logger.Debugfln("Detected native client, returning refresh_token in response body. Headers: %+v", c.Request.Header)
		response.Data = &types.DataOrPage[LoginResponseBody]{
			Item: &LoginResponseBody{
				AccessToken:  accessTokenStr,
				RefreshToken: refreshTokenStr,
				SessionId:    types.NewJsonInt64(sessId),
				User:         *user,
			},
		}

		handlerDeps.Logger.Debugfln("Login response: %+v", response.Data.Item)
		c.JSON(http.StatusOK, response)
		c.Abort()
	}
}
