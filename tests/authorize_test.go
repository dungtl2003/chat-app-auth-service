package tests

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/httpclient"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAuthorizeReturnCorrectStatus(t *testing.T) {
	helper := NewHelper()
	err := helper.Snapshot()
	require.NoError(t, err)
	defer func() {
		err := helper.Rollback()
		require.NoError(t, err)
	}()

	// 200 status code will be tested seperately because it need real token (login first)
	testcases := []struct {
		authHeader string
		status     int
	}{
		{
			authHeader: "Bearer sometoken",
			status:     401,
		},
		{
			authHeader: "sometoken",
			status:     401,
		},
		{
			authHeader: "",
			status:     401,
		},
		{
			authHeader: "Bearer   sometoken",
			status:     401,
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("authHeader: %s, status: %d", tc.authHeader, tc.status), func(t *testing.T) {

			URL := fmt.Sprintf("%s/authorize", helper.AuthURL)
			header := http.Header{
				"Authorization": {tc.authHeader},
			}

			resp, err := helper.Client.Get(URL, header)
			require.NoError(t, err)
			require.EqualValues(t, tc.status, resp.StatusCode)
		})
	}
}

func TestAuthorizeWithRealToken(t *testing.T) {
	helper := NewHelper()
	err := helper.Snapshot()
	require.NoError(t, err)
	defer func() {
		err := helper.Rollback()
		require.NoError(t, err)
	}()

	identifier := "normaluser"
	password := "normalpassword"
	deviceId := 1
	payloadJson := fmt.Appendf(nil, `
			{
				"identifier": "%s",
				"password": "%s",
				"device_id": "%d"
			}`, identifier, password, deviceId)

	URL := fmt.Sprintf("%s/login", helper.AuthURL)
	resp, err := helper.Client.Post(URL, nil, bytes.NewBuffer(payloadJson))
	require.NoError(t, err)
	require.EqualValues(t, 200, resp.StatusCode)

	respBody, err := httpclient.ReadResponse(resp)
	require.NoError(t, err)
	accessTokenPrefix := `"access_token: `
	accessTokenSuffix := `"`
	accessToken := string(respBody)
	accessToken = accessToken[len(accessTokenPrefix):]
	accessToken = accessToken[:len(accessToken)-len(accessTokenSuffix)]

	URL = fmt.Sprintf("%s/authorize", helper.AuthURL)
	header := http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", accessToken)},
	}
	resp, err = helper.Client.Get(URL, header)
	require.NoError(t, err)
	require.EqualValues(t, 200, resp.StatusCode)

	// fmt.Printf("%d\n", duration)
	// time.Sleep(duration)
	// idk why we can't use time.Sleep(duration) here
	start := time.Now()
	duration := time.Duration(helper.ATDurationMs) * time.Millisecond
	for start.Add(duration).After(time.Now()) {
	}
	header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", accessToken)},
	}
	resp, err = helper.Client.Get(URL, header)
	require.NoError(t, err)
	// token should be expired
	require.EqualValues(t, 401, resp.StatusCode)
}
