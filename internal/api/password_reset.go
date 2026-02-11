package api

import (
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/services/user"
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RequestPasswordResetRequestBody struct {
	Email string `json:"email" validate:"required"`
}

func RequestPasswordReset(handlerDeps *HandlerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := types.Response[any]{}

		var reqBody RequestPasswordResetRequestBody
		if err := c.ShouldBindJSON(&reqBody); err != nil {
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

		handlerDeps.Logger.Debugfln("Request password reset request body: %#v", reqBody)

		if err := handlerDeps.Validator.Validate(reqBody); err != nil {
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

		attempts, err := handlerDeps.RedisClient.IncrPasswordResetRateLimit(c, reqBody.Email)
		if err != nil {
			handlerDeps.Logger.Errorfln("IncrPwdResetRateLimit(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		// This is the first attempt, set expiration
		if attempts == 1 {
			if err := handlerDeps.RedisClient.ExpirePasswordResetRateLimit(
				c,
				reqBody.Email,
				handlerDeps.Config.PasswordResetConfig.RateLimitTtl,
			); err != nil {
				handlerDeps.Logger.Errorfln("SetPwdResetRateLimitExpiration(): %v", err)
				response.Error = &types.ErrorBlock{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Errors:  []types.ErrorItem{{Message: "Internal server error"}},
				}
				c.JSON(response.Error.Code, response)
				c.Abort()
				return
			}
		}

		if attempts > handlerDeps.Config.PasswordResetConfig.RateLimitMax {
			handlerDeps.Logger.Errorfln("Password reset rate limit exceeded for email: %s", reqBody.Email)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusTooManyRequests,
				Message: "Too many password reset requests. Please try again later.",
				Errors:  []types.ErrorItem{{Message: "Too many password reset requests. Please try again later."}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		code, err := helper.GenerateSecureOTP(6)
		if err != nil {
			handlerDeps.Logger.Errorfln("GenerateSecureOTP(): %v", err)
			handlerDeps.RedisClient.DecrPasswordResetRateLimit(c, reqBody.Email)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		if err := handlerDeps.RedisClient.SetPasswordResetCode(
			c,
			reqBody.Email,
			code,
			handlerDeps.Config.PasswordResetConfig.ResetCodeTtl,
		); err != nil {
			handlerDeps.Logger.Errorfln("SetPasswordResetCode(): %v", err)
			handlerDeps.RedisClient.DecrPasswordResetRateLimit(c, reqBody.Email)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		emailBody := fmt.Sprintf("Your password reset code is: %s", code)
		if err := handlerDeps.MailerService.Send(
			reqBody.Email,
			"Password Reset Request",
			emailBody,
		); err != nil {
			handlerDeps.Logger.Errorfln("MailerService.Send(): %v", err)
			handlerDeps.RedisClient.DecrPasswordResetRateLimit(c, reqBody.Email)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, response)
		c.Abort()
	}
}

type ResetPasswordRequestBody struct {
	Email       string `json:"email" validate:"required"`
	ResetCode   string `json:"reset_code" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

func ResetPassword(handlerDeps *HandlerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := types.Response[any]{}

		var reqBody ResetPasswordRequestBody
		if err := c.ShouldBindJSON(&reqBody); err != nil {
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

		att, err := handlerDeps.RedisClient.IncrPasswordResetAttempt(c, reqBody.Email)
		if err != nil {
			handlerDeps.Logger.Errorfln("IncrPasswordResetAttempt(): %v", err)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}
		if att == 1 {
			if err := handlerDeps.RedisClient.ExpirePasswordResetAttempt(c, reqBody.Email, handlerDeps.Config.PasswordResetConfig.AttemptTtl); err != nil {
				handlerDeps.Logger.Errorfln("ExpirePasswordResetAttempt(): %v", err)
				response.Error = &types.ErrorBlock{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Errors:  []types.ErrorItem{{Message: "Internal server error"}},
				}
				c.JSON(response.Error.Code, response)
				c.Abort()
				return
			}
		}
		if att > handlerDeps.Config.PasswordResetConfig.AttemptMax {
			handlerDeps.Logger.Errorfln("Password reset attempt limit exceeded for email: %s", reqBody.Email)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusTooManyRequests,
				Message: "Too many password reset attempts. Please try again later.",
				Errors:  []types.ErrorItem{{Message: "Too many password reset attempts. Please try again later."}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		storedCode, err := handlerDeps.RedisClient.GetPasswordResetCode(c, reqBody.Email)
		if err != nil {
			handlerDeps.Logger.Errorfln("GetPasswordResetCode(): %v", err)
			handlerDeps.RedisClient.DecrPasswordResetAttempt(c, reqBody.Email) // decrement attempt count on error
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		if storedCode == "" || storedCode != reqBody.ResetCode {
			handlerDeps.Logger.Errorfln("Invalid reset code for email: %s", reqBody.Email)
			response.Error = &types.ErrorBlock{
				Code:    http.StatusBadRequest,
				Message: "Invalid reset code",
				Errors:  []types.ErrorItem{{Message: "Invalid reset code"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		if err := handlerDeps.UserService.ResetPassword(c, &user.ResetPasswordRequest{
			Email:       reqBody.Email,
			NewPassword: reqBody.NewPassword,
		}); err != nil {
			handlerDeps.Logger.Errorfln("UpdateUserPassword(): %v", err)
			handlerDeps.RedisClient.DecrPasswordResetAttempt(c, reqBody.Email) // decrement attempt count on error
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		// delete reset code from redis
		if err := handlerDeps.RedisClient.DeletePasswordResetCode(c, reqBody.Email); err != nil {
			handlerDeps.Logger.Errorfln("DeletePasswordResetCode(): %v", err)
			handlerDeps.RedisClient.DecrPasswordResetAttempt(c, reqBody.Email) // decrement attempt count on error
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, response)
		c.Abort()
	}
}
