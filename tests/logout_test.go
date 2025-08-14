package tests

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/api"
	"dungtl2003/chat-app-auth-service/internal/services/database"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	USERS__LOGOUT_TEST_FILENAME = "users__logout_test.json"
)

func TestLogoutShouldLogoutOneDevice(t *testing.T) {
	helper := NewTestHelper()
	SetUp(helper, &SetUpOptions{
		DataFile: &database.DataFile{
			UserFile: USERS__LOGOUT_TEST_FILENAME,
		},
	})
	defer TearDown(helper)

	identifier := "normaluser"
	password := "normalpassword"
	deviceInfos := []json.RawMessage{
		json.RawMessage(`{"user-agent": "Mozilla/5.0"}`),
		json.RawMessage(`{"user-agent": "Chrome/5.0"}`),
		json.RawMessage(`{"user-agent": "Firefox/5.0"}`),
	}
	accessTokens := []string{}
	refreshTokens := []string{}
	sessionIds := []int64{}
	for _, devInfo := range deviceInfos {
		payloadJson := fmt.Appendf(nil, `
			{
				"identifier": "%s",
				"password": "%s",
				"device_info": %s
			}`, identifier, password, devInfo)

		URL := fmt.Sprintf("%s/auth/login", helper.AuthURL)
		resp, err := Post(helper.Client, URL, nil, bytes.NewBuffer(payloadJson))
		require.NoError(t, err)
		require.EqualValues(t, http.StatusOK, resp.StatusCode)

		var responseBody api.LoginResponseBody
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		require.NoError(t, err)

		sessionId := responseBody.SessionId
		sessionIds = append(sessionIds, sessionId)

		accessToken := responseBody.AccessToken
		require.NotEmpty(t, accessToken)
		accessTokens = append(accessTokens, accessToken)

		refreshToken := GetRTFromResponse(resp)
		require.NotEmpty(t, refreshToken)
		refreshTokens = append(refreshTokens, refreshToken)
	}

	// we will logout the first session
	URL := fmt.Sprintf("%s/auth/logout", helper.AuthURL)
	header := http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", accessTokens[0])},
	}
	resp, err := Post(helper.Client, URL, header, nil)
	require.NoError(t, err)
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	// other refresh tokens should work just fine
	for i, refreshToken := range refreshTokens {
		if i != 0 {
			URL = fmt.Sprintf("%s/auth/refresh", helper.AuthURL)
			header = http.Header{
				"Cookie": {fmt.Sprintf("refresh_token=%s", refreshToken)},
			}
			resp, err = Post(helper.Client, URL, header, nil)
			require.NoError(t, err)
			require.EqualValues(t, http.StatusOK, resp.StatusCode)
		}
	}

	// the first one cannot work
	// we have to test this one last because the server can misunderstood that
	// the token is reused or something
	URL = fmt.Sprintf("%s/auth/refresh", helper.AuthURL)
	header = http.Header{
		"Cookie": {fmt.Sprintf("refresh_token=%s", refreshTokens[0])},
	}
	resp, err = Post(helper.Client, URL, header, nil)
	require.NoError(t, err)
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)
}
