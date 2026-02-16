package api

import (
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

func SignUp(handlerDeps *HandlerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := types.Response[SignUpResponseBody]{}

		// validate request body
		var signUpRequestBody SignUpRequestBody
		if err := c.ShouldBindJSON(&signUpRequestBody); err != nil {
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

		handlerDeps.Logger.Debugfln("Sign up request body: %#v", signUpRequestBody)

		if err := handlerDeps.Validator.Validate(signUpRequestBody); err != nil {
			handlerDeps.Logger.Errorfln("Error while validating request body: %v", err)
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
		userResponse, err := handlerDeps.UserService.CreateUser(c, u.UserPostRequest{
			Email:    signUpRequestBody.Email,
			Username: signUpRequestBody.Username,
			Password: signUpRequestBody.Password,
			Role:     model.UserRole(signUpRequestBody.Role),
		})
		if err != nil {
			handlerDeps.Logger.Errorfln("CreateUser(): %v", err)
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
			user,
			handlerDeps.Config.JwtTokenConfig.ATDurationMs,
			handlerDeps.Config.JwtTokenConfig.RTDurationMs,
			sessId,
			handlerDeps.Config.IdGeneratorConfig.Epoch,
			handlerDeps.Config.JwtTokenConfig.Issuer,
			handlerDeps.Config.JwtTokenConfig.FrontendAudience,
		)
		if err != nil {
			handlerDeps.Logger.Errorfln("CreateTokens(): error creating tokens: %v", err)
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
			handlerDeps.Logger.Errorfln("GetTokenExpirationStr(): %v", err)
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
			DeviceInfo:       signUpRequestBody.DeviceInfo,
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

		session := sessionResponse.Session

		// add new session to user
		user.Sessions = append(user.Sessions, session)

		// security
		user.Password = ""

		native := helper.IsNativeClient(c)
		if !native {
			cfg := helper.GetCookieCfg(handlerDeps.Config.DomainName, handlerDeps.Config.Env)
			helper.SetCookieCfg(c, "refresh_token", refreshTokenStr, handlerDeps.Config.JwtTokenConfig.RTDurationMs, cfg)
			response.Data = &types.DataOrPage[SignUpResponseBody]{
				Item: &SignUpResponseBody{
					AccessToken: accessTokenStr,
					SessionId:   types.NewJsonInt64(sessId),
					User:        user,
				},
			}
			handlerDeps.Logger.Debugfln("Response body: %#v", response.Data.Item)
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
		handlerDeps.Logger.Debugfln("Response body: %#v", response.Data.Item)
		c.JSON(http.StatusOK, response)
		c.Abort()
	}
}
