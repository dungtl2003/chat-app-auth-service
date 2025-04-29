package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRefreshShouldWorkAsExpected(t *testing.T) {
	helper := NewHelper()
	err := helper.Snapshot()
	require.NoError(t, err)
	defer func() {
		err := helper.Rollback()
		require.NoError(t, err)
		err = helper.Db.Close()
		require.NoError(t, err)
	}()

	identifier := "normaluser"
	password := "normalpassword"
	deviceInfo := json.RawMessage(`{"user-agent": "Mozilla/5.0"}`)
	payloadJson := fmt.Appendf(nil, `
			{
				"identifier": "%s",
				"password": "%s",
				"device_info": %s
			}`, identifier, password, deviceInfo)

	URL := fmt.Sprintf("%s/login", helper.AuthURL)
	resp, err := helper.Client.Post(URL, nil, bytes.NewBuffer(payloadJson))
	require.NoError(t, err)
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	refreshToken := GetRTFromResponse(resp)
	require.NotEmpty(t, refreshToken)
	respJson, err := GetRespJson(resp)
	require.NoError(t, err)
	accessToken := respJson["access_token"].(string)
	require.NotEmpty(t, accessToken)

	SuckDelay(1000) // wait for 1 second to make sure no duplicate token

	URL = fmt.Sprintf("%s/refresh", helper.AuthURL)
	// pass cookie to next request
	header := http.Header{
		"Cookie": {fmt.Sprintf("refresh_token=%s", refreshToken)},
	}
	resp, err = helper.Client.Post(URL, header, nil)
	require.NoError(t, err)
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	respJson, err = GetRespJson(resp)
	require.NoError(t, err)
	newAccessToken := respJson["access_token"].(string)
	require.NotEmpty(t, newAccessToken)
	require.NotEqual(t, accessToken, newAccessToken)
	newRefreshToken := GetRTFromResponse(resp)
	require.NotEmpty(t, newRefreshToken)
	require.NotEqual(t, refreshToken, newRefreshToken)
}

func TestRefreshShouldNotWorkWithInvalidToken(t *testing.T) {
	helper := NewHelper()
	err := helper.Snapshot()
	require.NoError(t, err)
	defer func() {
		err := helper.Rollback()
		require.NoError(t, err)
		err = helper.Db.Close()
		require.NoError(t, err)
	}()

	URL := fmt.Sprintf("%s/refresh", helper.AuthURL)

	// no cookie
	resp, err := helper.Client.Post(URL, nil, nil)
	require.NoError(t, err)
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)

	// invalid refresh token
	header := http.Header{
		"Cookie": {fmt.Sprintf("refresh_token=%s", "invalidtoken")},
	}
	resp, err = helper.Client.Post(URL, header, nil)
	require.NoError(t, err)
	require.EqualValues(t, 401, resp.StatusCode)

	// expired refresh token
	identifier := "normaluser"
	password := "normalpassword"
	deviceInfo := json.RawMessage(`{"user-agent": "Mozilla/5.0"}`)
	payloadJson := fmt.Appendf(nil, `
			{
				"identifier": "%s",
				"password": "%s",
				"device_info": %s
			}`, identifier, password, deviceInfo)

	URL = fmt.Sprintf("%s/login", helper.AuthURL)
	resp, err = helper.Client.Post(URL, nil, bytes.NewBuffer(payloadJson))
	require.NoError(t, err)
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	refreshToken := GetRTFromResponse(resp)
	require.NotEmpty(t, refreshToken)
	SuckDelay(helper.RTDurationMs) // wait for refresh token to expire

	URL = fmt.Sprintf("%s/refresh", helper.AuthURL)
	header = http.Header{
		"Cookie": {fmt.Sprintf("refresh_token=%s", refreshToken)},
	}
	resp, err = helper.Client.Post(URL, header, nil)
	require.NoError(t, err)
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRefreshShouldHaveReuseDetection(t *testing.T) {
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
	refreshTokens := []string{}
	for i, devInfo := range deviceInfos {
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
		refreshToken := GetRTFromResponse(resp)
		require.NotEmpty(t, refreshToken)
		refreshTokens = append(refreshTokens, refreshToken)

		if i == 0 {
			// consume the first refresh token
			// SuckDelay(1000)
			URL = fmt.Sprintf("%s/refresh", helper.AuthURL)
			header := http.Header{
				"Cookie": {fmt.Sprintf("refresh_token=%s", refreshToken)},
			}
			resp, err = helper.Client.Post(URL, header, nil)
			require.NoError(t, err)
			refreshToken = GetRTFromResponse(resp)
			require.NotEmpty(t, refreshToken)
		}
	}

	// the first refresh token is used, if we try to use it again, it should fail
	refreshToken := refreshTokens[0]
	URL := fmt.Sprintf("%s/refresh", helper.AuthURL)
	header := http.Header{
		"Cookie": {fmt.Sprintf("refresh_token=%s", refreshToken)},
	}
	resp, err := helper.Client.Post(URL, header, nil)
	require.NoError(t, err)
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)

	// all other refresh tokens should NOT work because server invalidated them
	for _, refreshToken := range refreshTokens[1:] {
		URL = fmt.Sprintf("%s/refresh", helper.AuthURL)
		header := http.Header{
			"Cookie": {fmt.Sprintf("refresh_token=%s", refreshToken)},
		}
		resp, err = helper.Client.Post(URL, header, nil)
		require.NoError(t, err)
		require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)
	}
}
