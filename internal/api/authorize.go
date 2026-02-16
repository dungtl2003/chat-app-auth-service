package api

import (
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authorize(handlerDeps *HandlerDeps) gin.HandlerFunc {
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
		accessToken, err := jwthandler.DecodeToken(
			handlerDeps.Config.JwtTokenConfig.JwtSecret,
			accessTokenString,
		)
		if err != nil {
			handlerDeps.Logger.Errorfln("DecodeToken(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusUnauthorized,
				Message: "Invalid token",
				Errors:  []types.ErrorItem{{Message: "Invalid token"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		// make sure token is not internal token
		audiences, err := accessToken.Claims.GetAudience()
		if err != nil {
			handlerDeps.Logger.Errorfln("GetAudience(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}
		if len(audiences) == 0 {
			handlerDeps.Logger.Errorfln("token has no audience")
			response.Error = &types.ErrorBlock{
				Code:    http.StatusUnauthorized,
				Message: "Invalid token",
				Errors:  []types.ErrorItem{{Message: "Invalid token"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}
		if audiences[0] == handlerDeps.Config.JwtTokenConfig.InternalAudience {
			handlerDeps.Logger.Errorfln("token is internal token")
			response.Error = &types.ErrorBlock{
				Code:    http.StatusUnauthorized,
				Message: "Invalid token",
				Errors:  []types.ErrorItem{{Message: "Invalid token"}},
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

		// create internal token
		internalToken, err := jwthandler.CreateInternalToken(
			handlerDeps.Config.JwtTokenConfig.JwtSecret,
			handlerDeps.Config.JwtTokenConfig.ATDurationMs,
			userId,
			handlerDeps.Config.JwtTokenConfig.Issuer,
			handlerDeps.Config.JwtTokenConfig.InternalAudience,
		)
		if err != nil {
			handlerDeps.Logger.Errorfln("CreateInternalToken(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		// replace user token with internal token for internal service
		c.Header("Authorization", fmt.Sprintf("Bearer %s", internalToken))

		c.JSON(http.StatusOK, response)
	}
}
