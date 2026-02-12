package tests

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/api"
	"dungtl2003/chat-app-auth-service/internal/server"
	"dungtl2003/chat-app-auth-service/internal/services/database"
	"dungtl2003/chat-app-auth-service/internal/services/mailer"
	"dungtl2003/chat-app-auth-service/internal/types"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	PASSWORD_RESET_TEST_FILENAME = "password_reset_test.json"
)

func TestPasswordResetFlow(t *testing.T) {
	helper := NewTestHelper()

	var capturedCode string
	var mu sync.Mutex
	mockMailer := &mailer.MockMailer{
		MockSend: func(req *mailer.SendEmailRequest) error {
			mu.Lock()
			defer mu.Unlock()
			// emailBody := fmt.Sprintf("Your password reset code is: %s", code)
			prefix := "Your password reset code is: "
			if after, ok := strings.CutPrefix(req.Body, prefix); ok {
				capturedCode = after
			}
			return nil
		},
	}

	SetUp(helper, &SetUpOptions{
		DataFile: &database.DataFile{
			UserFile: PASSWORD_RESET_TEST_FILENAME,
		},
		ServerOptions: &server.AuthServerOptions{
			MailerService: mockMailer,
		},
	})
	defer TearDown(helper)

	email := "test-reset@example.com"

	// 1. Request Password Reset
	t.Run("RequestPasswordReset_Success", func(t *testing.T) {
		requestPayload := api.RequestPasswordResetRequestBody{
			Email: email,
		}
		payloadJson, err := json.Marshal(requestPayload)
		require.NoError(t, err)

		url := fmt.Sprintf("%s/auth/password-reset/request", helper.AuthURL)
		resp, err := Post(helper.Client, url, nil, bytes.NewBuffer(payloadJson))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		mu.Lock()
		require.NotEmpty(t, capturedCode)
		mu.Unlock()
	})

	// 2. Verify OTP
	var capturedToken string
	t.Run("VerifyOTP_Success", func(t *testing.T) {
		mu.Lock()
		code := capturedCode
		mu.Unlock()

		verifyPayload := api.VerifyOtpRequestBody{
			Email:     email,
			ResetCode: code,
		}
		payloadJson, err := json.Marshal(verifyPayload)
		require.NoError(t, err)

		url := fmt.Sprintf("%s/auth/password-reset/verify-otp", helper.AuthURL)
		resp, err := Post(helper.Client, url, nil, bytes.NewBuffer(payloadJson))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var respBody types.Response[api.VerifyOtpResponseBody]
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err)
		require.NotNil(t, respBody.Data)
		require.NotNil(t, respBody.Data.Item)
		capturedToken = respBody.Data.Item.ResetToken
	})

	// 3. Reset Password
	t.Run("ResetPassword_Success", func(t *testing.T) {
		resetPayload := api.ResetPasswordRequestBody{
			Email:       email,
			ResetToken:  capturedToken,
			NewPassword: "newpassword123",
		}
		payloadJson, err := json.Marshal(resetPayload)
		require.NoError(t, err)

		url := fmt.Sprintf("%s/auth/password-reset/reset", helper.AuthURL)
		resp, err := Post(helper.Client, url, nil, bytes.NewBuffer(payloadJson))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 4. Verify OTP - Invalid Code
	t.Run("VerifyOTP_InvalidCode", func(t *testing.T) {
		verifyPayload := api.VerifyOtpRequestBody{
			Email:     "another-user@example.com",
			ResetCode: "123456",
		}
		payloadJson, err := json.Marshal(verifyPayload)
		require.NoError(t, err)

		url := fmt.Sprintf("%s/auth/password-reset/verify-otp", helper.AuthURL)
		resp, err := Post(helper.Client, url, nil, bytes.NewBuffer(payloadJson))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var respBody types.Response[any]
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err)
		require.NotNil(t, respBody.Error)
		require.Equal(t, "Invalid reset code", respBody.Error.Message)
	})

	// 5. Reset Password - Non-registered Email
	t.Run("ResetPassword_NonRegisteredEmail", func(t *testing.T) {
		nonRegisteredEmail := "non-registered@example.com"

		// 1. Request
		requestPayload := api.RequestPasswordResetRequestBody{
			Email: nonRegisteredEmail,
		}
		payloadJson, _ := json.Marshal(requestPayload)
		url := fmt.Sprintf("%s/auth/password-reset/request", helper.AuthURL)
		resp, err := Post(helper.Client, url, nil, bytes.NewBuffer(payloadJson))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		mu.Lock()
		code := capturedCode
		mu.Unlock()
		require.NotEmpty(t, code)

		// 2. Verify
		verifyPayload := api.VerifyOtpRequestBody{
			Email:     nonRegisteredEmail,
			ResetCode: code,
		}
		payloadJson, _ = json.Marshal(verifyPayload)
		url = fmt.Sprintf("%s/auth/password-reset/verify-otp", helper.AuthURL)
		resp, err = Post(helper.Client, url, nil, bytes.NewBuffer(payloadJson))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var verifyResp types.Response[api.VerifyOtpResponseBody]
		err = json.NewDecoder(resp.Body).Decode(&verifyResp)
		require.NoError(t, err)
		token := verifyResp.Data.Item.ResetToken

		// 3. Reset (Should fail with 404)
		resetPayload := api.ResetPasswordRequestBody{
			Email:       nonRegisteredEmail,
			ResetToken:  token,
			NewPassword: "newpassword123",
		}
		payloadJson, _ = json.Marshal(resetPayload)
		url = fmt.Sprintf("%s/auth/password-reset/reset", helper.AuthURL)
		resp, err = Post(helper.Client, url, nil, bytes.NewBuffer(payloadJson))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestPasswordResetRateLimits(t *testing.T) {

	// 1. Request Rate Limit
	t.Run("RequestRateLimit", func(t *testing.T) {
		t.Setenv("PASSWORD_RESET_RATE_LIMIT_MAX", "2")
		helper := NewTestHelper()
		mockMailer := &mailer.MockMailer{
			MockSend: func(req *mailer.SendEmailRequest) error {
				return nil
			},
		}
		SetUp(helper, &SetUpOptions{
			DataFile: &database.DataFile{UserFile: PASSWORD_RESET_TEST_FILENAME},
			ServerOptions: &server.AuthServerOptions{
				MailerService: mockMailer,
			},
		})
		defer TearDown(helper)

		email := "rate-limit-request@example.com"
		url := fmt.Sprintf("%s/auth/password-reset/request", helper.AuthURL)
		payload, _ := json.Marshal(api.RequestPasswordResetRequestBody{Email: email})

		// 1st request - OK
		resp, err := Post(helper.Client, url, nil, bytes.NewBuffer(payload))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// 2nd request - OK
		resp, err = Post(helper.Client, url, nil, bytes.NewBuffer(payload))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// 3rd request - Too Many Requests
		resp, err = Post(helper.Client, url, nil, bytes.NewBuffer(payload))
		require.NoError(t, err)
		require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	})

	// 2. Request Rate Limit TTL
	t.Run("RequestRateLimitTTL", func(t *testing.T) {
		t.Setenv("PASSWORD_RESET_RATE_LIMIT_MAX", "1")
		t.Setenv("PASSWORD_RESET_RATE_LIMIT_TTL", "1s")
		helper := NewTestHelper()
		mockMailer := &mailer.MockMailer{
			MockSend: func(req *mailer.SendEmailRequest) error {
				return nil
			},
		}
		SetUp(helper, &SetUpOptions{
			DataFile: &database.DataFile{UserFile: PASSWORD_RESET_TEST_FILENAME},
			ServerOptions: &server.AuthServerOptions{
				MailerService: mockMailer,
			},
		})
		defer TearDown(helper)

		email := "rate-limit-ttl-request@example.com"
		url := fmt.Sprintf("%s/auth/password-reset/request", helper.AuthURL)
		payload, _ := json.Marshal(api.RequestPasswordResetRequestBody{Email: email})

		// 1st request - OK
		resp, err := Post(helper.Client, url, nil, bytes.NewBuffer(payload))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// 2nd request - Too Many Requests
		resp, err = Post(helper.Client, url, nil, bytes.NewBuffer(payload))
		require.NoError(t, err)
		require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

		// Wait for TTL to expire
		time.Sleep(1500 * time.Millisecond)

		// 3rd request - OK
		resp, err = Post(helper.Client, url, nil, bytes.NewBuffer(payload))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 3. Attempt Limit
	t.Run("AttemptLimit", func(t *testing.T) {
		t.Setenv("PASSWORD_RESET_ATTEMPT_MAX", "2")
		helper := NewTestHelper()
		mockMailer := &mailer.MockMailer{
			MockSend: func(req *mailer.SendEmailRequest) error {
				return nil
			},
		}
		SetUp(helper, &SetUpOptions{
			DataFile: &database.DataFile{UserFile: PASSWORD_RESET_TEST_FILENAME},
			ServerOptions: &server.AuthServerOptions{
				MailerService: mockMailer,
			},
		})
		defer TearDown(helper)

		email := "test-reset@example.com"
		url := fmt.Sprintf("%s/auth/password-reset/verify-otp", helper.AuthURL)
		payload, _ := json.Marshal(api.VerifyOtpRequestBody{
			Email:     email,
			ResetCode: "WRONG",
		})

		// 1st attempt - Bad Request
		resp, err := Post(helper.Client, url, nil, bytes.NewBuffer(payload))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// 2nd attempt - Bad Request
		resp, err = Post(helper.Client, url, nil, bytes.NewBuffer(payload))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// 3rd attempt - Too Many Requests
		resp, err = Post(helper.Client, url, nil, bytes.NewBuffer(payload))
		require.NoError(t, err)
		require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	})

	// 4. Attempt Limit TTL
	t.Run("AttemptLimitTTL", func(t *testing.T) {
		t.Setenv("PASSWORD_RESET_ATTEMPT_MAX", "1")
		t.Setenv("PASSWORD_RESET_ATTEMPT_TTL", "1s")
		helper := NewTestHelper()
		mockMailer := &mailer.MockMailer{
			MockSend: func(req *mailer.SendEmailRequest) error {
				return nil
			},
		}
		SetUp(helper, &SetUpOptions{
			DataFile: &database.DataFile{UserFile: PASSWORD_RESET_TEST_FILENAME},
			ServerOptions: &server.AuthServerOptions{
				MailerService: mockMailer,
			},
		})
		defer TearDown(helper)

		email := "test-reset@example.com"
		url := fmt.Sprintf("%s/auth/password-reset/verify-otp", helper.AuthURL)
		payload, _ := json.Marshal(api.VerifyOtpRequestBody{
			Email:     email,
			ResetCode: "WRONG",
		})

		// 1st attempt - Bad Request
		resp, err := Post(helper.Client, url, nil, bytes.NewBuffer(payload))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// 2nd attempt - Too Many Requests
		resp, err = Post(helper.Client, url, nil, bytes.NewBuffer(payload))
		require.NoError(t, err)
		require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

		// Wait for TTL to expire
		time.Sleep(1500 * time.Millisecond)

		// 3rd attempt - Bad Request (back to normal)
		resp, err = Post(helper.Client, url, nil, bytes.NewBuffer(payload))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// 5. Code TTL
	t.Run("CodeTTL", func(t *testing.T) {
		t.Setenv("PASSWORD_RESET_CODE_TTL", "1s")

		var capturedCode string
		var mu sync.Mutex
		mockMailer := &mailer.MockMailer{
			MockSend: func(req *mailer.SendEmailRequest) error {
				mu.Lock()
				defer mu.Unlock()
				prefix := "Your password reset code is: "
				if after, ok := strings.CutPrefix(req.Body, prefix); ok {
					capturedCode = after
				}
				return nil
			},
		}

		helper := NewTestHelper()
		SetUp(helper, &SetUpOptions{
			DataFile: &database.DataFile{UserFile: PASSWORD_RESET_TEST_FILENAME},
			ServerOptions: &server.AuthServerOptions{
				MailerService: mockMailer,
			},
		})
		defer TearDown(helper)

		email := "test-reset@example.com"

		// Request
		reqUrl := fmt.Sprintf("%s/auth/password-reset/request", helper.AuthURL)
		reqPayload, _ := json.Marshal(api.RequestPasswordResetRequestBody{Email: email})
		_, err := Post(helper.Client, reqUrl, nil, bytes.NewBuffer(reqPayload))
		require.NoError(t, err)

		mu.Lock()
		code := capturedCode
		mu.Unlock()
		require.NotEmpty(t, code)

		// Wait for TTL to expire
		time.Sleep(1500 * time.Millisecond)

		// Verify OTP
		confUrl := fmt.Sprintf("%s/auth/password-reset/verify-otp", helper.AuthURL)
		confPayload, _ := json.Marshal(api.VerifyOtpRequestBody{
			Email:     email,
			ResetCode: code,
		})
		resp, err := Post(helper.Client, confUrl, nil, bytes.NewBuffer(confPayload))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode) // Should fail because code expired
	})
}
