package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogoutShouldLogoutOneDevice(t *testing.T) {
	helper := NewHelper()
	err := helper.Snapshot()
	require.NoError(t, err)
	defer func() {
		err := helper.Rollback()
		require.NoError(t, err)
	}()

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

		URL := fmt.Sprintf("%s/login", helper.AuthURL)
		resp, err := helper.Client.Post(URL, nil, bytes.NewBuffer(payloadJson))
		require.NoError(t, err)
		require.EqualValues(t, http.StatusOK, resp.StatusCode)

		respJson, err := GetRespJson(resp)
		require.NoError(t, err)

		sessionId := int64(respJson["session_id"].(float64))
		sessionIds = append(sessionIds, sessionId)

		accessToken := respJson["access_token"].(string)
		require.NotEmpty(t, accessToken)
		accessTokens = append(accessTokens, accessToken)

		refreshToken := GetRTFromResponse(resp)
		require.NotEmpty(t, refreshToken)
		refreshTokens = append(refreshTokens, refreshToken)
	}

	// we will logout the first session
	URL := fmt.Sprintf("%s/logout", helper.AuthURL)
	header := http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", accessTokens[0])},
	}
	resp, err := helper.Client.Post(URL, header, nil)
	require.NoError(t, err)
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	// other refresh tokens should work just fine
	for i, refreshToken := range refreshTokens {
		if i != 0 {
			URL = fmt.Sprintf("%s/refresh", helper.AuthURL)
			header = http.Header{
				"Cookie": {fmt.Sprintf("refresh_token=%s", refreshToken)},
			}
			resp, err = helper.Client.Post(URL, header, nil)
			require.NoError(t, err)
			require.EqualValues(t, http.StatusOK, resp.StatusCode)
		}
	}

	// the first one cannot work
	// we have to test this one last because the server can misunderstood that
	// the token is reused or something
	URL = fmt.Sprintf("%s/refresh", helper.AuthURL)
	header = http.Header{
		"Cookie": {fmt.Sprintf("refresh_token=%s", refreshTokens[0])},
	}
	resp, err = helper.Client.Post(URL, header, nil)
	require.NoError(t, err)
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)
}
