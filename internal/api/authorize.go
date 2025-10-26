package api

import (
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/services"
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authorize(appCtx *context.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := types.Response[model.ChatUser]{}

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
			appCtx.Logger.Errorfln("GetAudience(): %v", err)
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
			appCtx.Logger.Errorfln("token has no audience")
			response.Error = &types.ErrorBlock{
				Code:    http.StatusUnauthorized,
				Message: "Invalid token",
				Errors:  []types.ErrorItem{{Message: "Invalid token"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}
		if audiences[0] == jwthandler.INTERNAL_AUDIENCE {
			appCtx.Logger.Errorfln("token is internal token")
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

		// create internal token
		internalToken, err := jwthandler.CreateInternalToken(appCtx.JwtConfig.JwtSecret, appCtx.JwtConfig.ATDurationMs, userId)
		if err != nil {
			appCtx.Logger.Errorfln("CreateInternalToken(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		// get user information
		userResponse, err := appCtx.UserService.GetUserById(c, userId)
		if err != nil {
			appCtx.Logger.Errorfln("GetUserById(): %v", err)
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
		// sanitize user data
		if user.Password != "" {
			appCtx.Logger.Warnfln("User password should not be returned in response")
			user.Password = ""
		}
		// replace user token with internal token for internal service
		c.Header("Authorization", fmt.Sprintf("Bearer %s", internalToken))

		response.Data = &types.DataOrPage[model.ChatUser]{
			Item: &user,
		}

		appCtx.Logger.Debugfln("Response body: %+v", response.Data.Item)
		c.JSON(http.StatusOK, response)
	}
}
