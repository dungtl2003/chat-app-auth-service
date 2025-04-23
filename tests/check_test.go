package tests

import (
	"bytes"
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

			URL := fmt.Sprintf("%s/check", helper.AuthURL)
			header := http.Header{
				"Authorization": {tc.authHeader},
			}

			resp, err := helper.Client.Get(URL, header)
			require.NoError(t, err)
			require.EqualValues(t, tc.status, resp.StatusCode)
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

	accessToken := GetATFromResponse(resp)
	require.NotEmpty(t, accessToken)

	URL = fmt.Sprintf("%s/check", helper.AuthURL)
	header := http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", accessToken)},
	}
	resp, err = helper.Client.Get(URL, header)
	require.NoError(t, err)
	require.EqualValues(t, 200, resp.StatusCode)

	SuckDelay(helper.ATDurationMs)
	header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", accessToken)},
	}
	resp, err = helper.Client.Get(URL, header)
	require.NoError(t, err)
	// token should be expired
	require.EqualValues(t, 401, resp.StatusCode)
}
