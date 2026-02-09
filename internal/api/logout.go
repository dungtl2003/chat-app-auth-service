package api

import (
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/types"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Logout(handlerDeps *HandlerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := types.Response[any]{}

		// Bearer <token>
		authHeader := c.GetHeader("Authorization")
		handlerDeps.Logger.Debugfln("Authorization header: %s", authHeader)
		if authHeader == "" {
			handlerDeps.Logger.Errorfln("Missing authorization header")
			response.Error = &types.ErrorBlock{
				Code:    http.StatusUnauthorized,
				Message: "Missing authorization header",
				Errors:  []types.ErrorItem{{Message: "Missing authorization header"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			handlerDeps.Logger.Errorfln("Authorization header should have 2 parts")
			response.Error = &types.ErrorBlock{
				Code:    http.StatusUnauthorized,
				Message: "Authorization header should have 2 parts",
				Errors:  []types.ErrorItem{{Message: "Authorization header should have 2 parts"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		if parts[0] != "Bearer" {
			handlerDeps.Logger.Errorfln("Authorization header should start with `Bearer`")
			response.Error = &types.ErrorBlock{
				Code:    http.StatusUnauthorized,
				Message: "Authorization header should start with `Bearer`",
				Errors:  []types.ErrorItem{{Message: "Authorization header should start with `Bearer`"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		accessTokenString := parts[1]
		accessToken, err := jwthandler.DecodeToken(handlerDeps.Config.JwtTokenConfig.JwtSecret, accessTokenString)
		if err != nil {
			handlerDeps.Logger.Errorfln("DecodeToken(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusUnauthorized,
				Message: "Invalid access token",
				Errors:  []types.ErrorItem{{Message: "Invalid access token"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		parsedToken, err := helper.ParseToken(accessToken)
		if err != nil {
			handlerDeps.Logger.Errorfln("ParseToken(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}
		userId := parsedToken.UserId
		sessId := parsedToken.SessionId

		// revoke session
		err = handlerDeps.UserService.RevokeSession(c, userId, sessId)
		if err != nil {
			handlerDeps.Logger.Errorfln("RevokeSession(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		handlerDeps.Logger.Debugfln("Revoked session %d for user %d", sessId, userId)

		native := helper.IsNativeClient(c)
		if !native {
			helper.ClearCookie(c, "refresh_token", handlerDeps.Config.DomainName)
		}

		handlerDeps.Logger.Debugfln("Logout successful for user %d", userId)
		c.JSON(http.StatusOK, response)
		c.Abort()
	}
}
