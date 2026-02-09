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
		MockSend: func(to, subject, body string) error {
			mu.Lock()
			defer mu.Unlock()
			// emailBody := fmt.Sprintf("Your password reset code is: %s", code)
			prefix := "Your password reset code is: "
			if after, ok := strings.CutPrefix(body, prefix); ok {
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

	// 2. Confirm Password Reset
	t.Run("ConfirmPasswordReset_Success", func(t *testing.T) {
		mu.Lock()
		code := capturedCode
		mu.Unlock()

		confirmPayload := api.ResetPasswordRequestBody{
			Email:       email,
			ResetCode:   code,
			NewPassword: "newpassword123",
		}
		payloadJson, err := json.Marshal(confirmPayload)
		require.NoError(t, err)

		url := fmt.Sprintf("%s/auth/password-reset/confirm", helper.AuthURL)
		resp, err := Post(helper.Client, url, nil, bytes.NewBuffer(payloadJson))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 3. Confirm Password Reset - Invalid Code
	t.Run("ConfirmPasswordReset_InvalidCode", func(t *testing.T) {
		confirmPayload := api.ResetPasswordRequestBody{
			Email:       "another-user@example.com",
			ResetCode:   "123456",
			NewPassword: "newpassword123",
		}
		payloadJson, err := json.Marshal(confirmPayload)
		require.NoError(t, err)

		url := fmt.Sprintf("%s/auth/password-reset/confirm", helper.AuthURL)
		resp, err := Post(helper.Client, url, nil, bytes.NewBuffer(payloadJson))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var respBody types.Response[any]
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err)
		require.NotNil(t, respBody.Error)
		require.Equal(t, "Invalid reset code", respBody.Error.Message)
	})
}
