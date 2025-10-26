package api

import (
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/types"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Logout(appCtx *context.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := types.Response[any]{}

		// Bearer <token>
		authHeader := c.GetHeader("Authorization")
		appCtx.Logger.Debugfln("Authorization header: %s", authHeader)
		if authHeader == "" {
			appCtx.Logger.Errorfln("Missing authorization header")
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
			appCtx.Logger.Errorfln("Authorization header should have 2 parts")
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
			appCtx.Logger.Errorfln("Authorization header should start with `Bearer`")
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
		accessToken, err := jwthandler.DecodeToken(appCtx.JwtConfig.JwtSecret, accessTokenString)
		if err != nil {
			appCtx.Logger.Errorfln("DecodeToken(): %v", err)
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
			appCtx.Logger.Errorfln("ParseToken(): %v", err)
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
		err = appCtx.UserService.RevokeSession(c, userId, sessId)
		if err != nil {
			appCtx.Logger.Errorfln("RevokeSession(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		appCtx.Logger.Debugfln("Revoked session %d for user %d", sessId, userId)

		native := helper.IsNativeClient(c)
		if !native {
			helper.ClearCookie(c, "refresh_token", appCtx.DomainName)
		}

		appCtx.Logger.Debugfln("Logout successful for user %d", userId)
		c.JSON(http.StatusOK, response)
		c.Abort()
	}
}
