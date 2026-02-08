package tests

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/api"
	"dungtl2003/chat-app-auth-service/internal/constants"
	"dungtl2003/chat-app-auth-service/internal/services/database"
	"dungtl2003/chat-app-auth-service/internal/types"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	USERS__REFRESH_TEST_FILENAME = "users__refresh_test.json"
)

func TestRefreshShouldWorkAsExpected(t *testing.T) {
	helper := NewTestHelper()
	SetUp(helper, &SetUpOptions{
		DataFile: &database.DataFile{
			UserFile: USERS__REFRESH_TEST_FILENAME,
		},
	})
	defer TearDown(helper)

	identifier := "normaluser"
	password := "normalpassword"
	deviceInfo := json.RawMessage(`{"user-agent": "Mozilla/5.0"}`)
	payloadJson := fmt.Appendf(nil, `
			{
				"identifier": "%s",
				"password": "%s",
				"device_info": %s
			}`, identifier, password, deviceInfo)

	URL := fmt.Sprintf("%s/auth/login", helper.AuthURL)
	resp, err := Post(helper.Client, URL, nil, bytes.NewBuffer(payloadJson))
	require.NoError(t, err)
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	refreshToken := GetRTFromResponse(resp)
	require.NotEmpty(t, refreshToken)

	var loginResponseBody types.Response[api.LoginResponseBody]
	err = json.NewDecoder(resp.Body).Decode(&loginResponseBody)
	require.NoError(t, err)
	require.Nil(t, loginResponseBody.Error)
	require.NotNil(t, loginResponseBody.Data)

	accessToken := loginResponseBody.Data.Item.AccessToken
	require.NotEmpty(t, accessToken)

	<-time.After(1 * time.Second) // wait for 1 second to make sure no duplicate token

	refreshBody := api.RefreshTokenRequestBody{
		DeviceInfo: types.NewJson([]byte(`{"user-agent": "Mozilla/5.0"}`)),
	}
	refreshBodyJson, err := json.Marshal(refreshBody)
	require.NoError(t, err)

	URL = fmt.Sprintf("%s/auth/refresh", helper.AuthURL)
	// pass cookie to next request
	header := http.Header{
		"Cookie": {fmt.Sprintf("refresh_token=%s", refreshToken)},
	}
	resp, err = Post(helper.Client, URL, header, bytes.NewBuffer(refreshBodyJson))
	require.NoError(t, err)
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	var refreshResponseBody types.Response[api.RefreshTokenResponseBody]
	err = json.NewDecoder(resp.Body).Decode(&refreshResponseBody)
	require.NoError(t, err)
	require.Nil(t, refreshResponseBody.Error)
	require.NotNil(t, refreshResponseBody.Data)

	newAccessToken := refreshResponseBody.Data.Item.AccessToken
	require.NotEmpty(t, newAccessToken)

	require.NotEqual(t, accessToken, newAccessToken)
	newRefreshToken := GetRTFromResponse(resp)
	require.NotEmpty(t, newRefreshToken)
	require.NotEqual(t, refreshToken, newRefreshToken)
}

func TestRefreshShouldNotWorkWithInvalidToken(t *testing.T) {
	helper := NewTestHelper()
	SetUp(helper, &SetUpOptions{
		DataFile: &database.DataFile{
			UserFile: USERS__REFRESH_TEST_FILENAME,
		},
	})
	defer TearDown(helper)

	URL := fmt.Sprintf("%s/auth/refresh", helper.AuthURL)

	refreshBody := api.RefreshTokenRequestBody{
		DeviceInfo: types.NewJson([]byte(`{"user-agent": "Mozilla/5.0"}`)),
	}
	refreshBodyJson, err := json.Marshal(refreshBody)
	require.NoError(t, err)

	// no cookie
	resp, err := Post(helper.Client, URL, nil, bytes.NewBuffer(refreshBodyJson))
	require.NoError(t, err)
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)
	var errResponseBody types.Response[api.RefreshTokenResponseBody]

	err = json.NewDecoder(resp.Body).Decode(&errResponseBody)
	require.NoError(t, err)
	require.NotNil(t, errResponseBody.Error)
	require.EqualValues(t, constants.REFRESH_TOKEN_NOT_FOUND, errResponseBody.Error.Status)

	// invalid refresh token
	header := http.Header{
		"Cookie": {fmt.Sprintf("refresh_token=%s", "invalidtoken")},
	}
	resp, err = Post(helper.Client, URL, header, bytes.NewBuffer(refreshBodyJson))
	require.NoError(t, err)
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)

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

	URL = fmt.Sprintf("%s/auth/login", helper.AuthURL)
	resp, err = Post(helper.Client, URL, nil, bytes.NewBuffer(payloadJson))
	require.NoError(t, err)
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	refreshToken := GetRTFromResponse(resp)
	require.NotEmpty(t, refreshToken)

	<-time.After(time.Millisecond * time.Duration(helper.RTDurationMs))

	URL = fmt.Sprintf("%s/auth/refresh", helper.AuthURL)
	header = http.Header{
		"Cookie": {fmt.Sprintf("refresh_token=%s", refreshToken)},
	}
	resp, err = Post(helper.Client, URL, header, bytes.NewBuffer(refreshBodyJson))
	require.NoError(t, err)
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRefreshShouldHaveReuseDetection(t *testing.T) {
	helper := NewTestHelper()
	SetUp(helper, &SetUpOptions{
		DataFile: &database.DataFile{
			UserFile: USERS__REFRESH_TEST_FILENAME,
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
	refreshTokens := []string{}

	refreshBody := api.RefreshTokenRequestBody{
		DeviceInfo: types.NewJson([]byte(`{"user-agent": "Mozilla/5.0"}`)),
	}
	refreshBodyJson, err := json.Marshal(refreshBody)
	require.NoError(t, err)

	for i, devInfo := range deviceInfos {
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
		refreshToken := GetRTFromResponse(resp)
		require.NotEmpty(t, refreshToken)
		refreshTokens = append(refreshTokens, refreshToken)

		if i == 0 {
			// consume the first refresh token
			URL = fmt.Sprintf("%s/auth/refresh", helper.AuthURL)
			header := http.Header{
				"Cookie": {fmt.Sprintf("refresh_token=%s", refreshToken)},
			}
			resp, err = Post(helper.Client, URL, header, bytes.NewBuffer(refreshBodyJson))
			require.NoError(t, err)
			refreshToken = GetRTFromResponse(resp)
			require.NotEmpty(t, refreshToken)
		}
	}

	// the first refresh token is used, if we try to use it again, it should fail
	refreshToken := refreshTokens[0]
	URL := fmt.Sprintf("%s/auth/refresh", helper.AuthURL)
	header := http.Header{
		"Cookie": {fmt.Sprintf("refresh_token=%s", refreshToken)},
	}
	resp, err := Post(helper.Client, URL, header, bytes.NewBuffer(refreshBodyJson))
	require.NoError(t, err)
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)

	// all other refresh tokens should NOT work because server invalidated them
	for _, refreshToken := range refreshTokens[1:] {
		URL = fmt.Sprintf("%s/auth/refresh", helper.AuthURL)
		header := http.Header{
			"Cookie": {fmt.Sprintf("refresh_token=%s", refreshToken)},
		}
		resp, err = Post(helper.Client, URL, header, bytes.NewBuffer(refreshBodyJson))
		require.NoError(t, err)
		require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)
	}
}
