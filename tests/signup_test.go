package tests

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/api"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/services/database"
	"dungtl2003/chat-app-auth-service/internal/types"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	USERS__SIGNUP_TEST_FILENAME = "users__signup_test.json"
)

func TestSignUpSuccessShouldAutoLogin(t *testing.T) {
	helper := NewTestHelper()
	SetUp(helper, &SetUpOptions{
		DataFile: &database.DataFile{
			UserFile: USERS__SIGNUP_TEST_FILENAME,
		},
	})
	defer TearDown(helper)

	signUpPayload := api.SignUpRequestBody{
		Email:      "you@example.com",
		Username:   "you",
		Password:   "password",
		Role:       model.USER,
		DeviceInfo: types.NewJson([]byte(`{"user-agent": "Mozilla/5.0"}`)),
	}
	payloadJson, err := json.Marshal(signUpPayload)
	require.NoError(t, err)
	URL := fmt.Sprintf("%s/signup", helper.AuthURL)
	header := http.Header{
		"Content-Type": []string{"application/json"},
	}

	resp, err := Post(helper.Client, URL, header, bytes.NewBuffer(payloadJson))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// cookie: "refresh_token": "%s"
	refreshToken := GetRTFromResponse(resp)
	require.NotEmpty(t, refreshToken)

	var respBody api.SignUpResponseBody
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.NotEmpty(t, respBody.AccessToken)
	require.NotEmpty(t, respBody.SessionId)
	require.EqualValues(t, signUpPayload.Email, respBody.User.Email)
	require.EqualValues(t, signUpPayload.Username, respBody.User.Username)
	require.EqualValues(t, signUpPayload.Role, respBody.User.Role)
	require.Empty(t, respBody.User.Password)
}
