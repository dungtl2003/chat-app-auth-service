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
		Role:       model.UserRoleUser,
		DeviceInfo: types.NewJson([]byte(`{"user-agent": "Mozilla/5.0"}`)),
	}
	payloadJson, err := json.Marshal(signUpPayload)
	require.NoError(t, err)
	URL := fmt.Sprintf("%s/auth/signup", helper.AuthURL)
	header := http.Header{
		"Content-Type": []string{"application/json"},
	}

	resp, err := Post(helper.Client, URL, header, bytes.NewBuffer(payloadJson))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// cookie: "refresh_token": "%s"
	refreshToken := GetRTFromResponse(resp)
	require.NotEmpty(t, refreshToken)

	var respBody types.Response[api.SignUpResponseBody]
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Nil(t, respBody.Error)
	require.NotNil(t, respBody.Data)

	respBodyData := respBody.Data.Item
	require.NotEmpty(t, respBodyData.AccessToken)
	require.NotEmpty(t, respBodyData.SessionId)
	require.EqualValues(t, signUpPayload.Email, respBodyData.User.Email)
	require.EqualValues(t, signUpPayload.Username, respBodyData.User.Username)
	require.EqualValues(t, signUpPayload.Role, respBodyData.User.Role)
	require.Empty(t, respBodyData.User.Password)
}
