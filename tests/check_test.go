package tests

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckReturnCorrectStatus(t *testing.T) {
	helper := NewHelper()
	err := helper.Snapshot()
	require.NoError(t, err)
	defer func() {
		err := helper.Rollback()
		require.NoError(t, err)
	}()

	// 200 status code will be tested seperately because it need real token (login first)
	testcases := []struct {
		authHeader     string
		expectedStatus int
	}{
		{
			authHeader:     "Bearer sometoken",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			authHeader:     "sometoken",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			authHeader:     "Bearer   sometoken",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("authHeader: %s, status: %d", tc.authHeader, tc.expectedStatus), func(t *testing.T) {

			URL := fmt.Sprintf("%s/check", helper.AuthURL)
			header := http.Header{
				"Authorization": {tc.authHeader},
			}

			resp, err := helper.Client.Get(URL, header)
			require.NoError(t, err)
			require.EqualValues(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestCheckWithRealToken(t *testing.T) {
	helper := NewHelper()
	err := helper.Snapshot()
	require.NoError(t, err)
	defer func() {
		err := helper.Rollback()
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

	accessToken := GetATFromResponse(resp)
	require.NotEmpty(t, accessToken)

	URL = fmt.Sprintf("%s/check", helper.AuthURL)
	header := http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", accessToken)},
	}
	resp, err = helper.Client.Get(URL, header)
	require.NoError(t, err)
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	SuckDelay(helper.ATDurationMs)
	header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", accessToken)},
	}
	resp, err = helper.Client.Get(URL, header)
	require.NoError(t, err)
	// token should be expired
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCheckWithInternalTokenShouldFail(t *testing.T) {
	helper := NewHelper()
	err := helper.Snapshot()
	require.NoError(t, err)
	defer func() {
		err := helper.Rollback()
		require.NoError(t, err)
	}()

	internalToken, err := jwthandler.CreateInternalToken(helper.JwtSecret, 5_000_000, 2)
	require.NoError(t, err)

	URL := fmt.Sprintf("%s/check", helper.AuthURL)
	header := http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", internalToken)},
	}

	resp, err := helper.Client.Get(URL, header)
	require.NoError(t, err)
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)
}
