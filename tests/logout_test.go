package tests

import (
	"bytes"
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
	deviceIds := []int{1, 2, 4}
	accessTokens := []string{}
	refreshTokens := []string{}
	for _, deviceId := range deviceIds {
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
		accessTokens = append(accessTokens, accessToken)

		refreshToken := GetRTFromResponse(resp)
		require.NotEmpty(t, refreshToken)
		refreshTokens = append(refreshTokens, refreshToken)
	}

	// we will logout the first device
	URL := fmt.Sprintf("%s/logout", helper.AuthURL)
	header := http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", accessTokens[0])},
	}
	resp, err := helper.Client.Get(URL, header)
	require.NoError(t, err)
	require.EqualValues(t, 200, resp.StatusCode)

	// other refresh tokens should work just fine
	for i, refreshToken := range refreshTokens {
		if i != 0 {
			URL = fmt.Sprintf("%s/refresh", helper.AuthURL)
			header = http.Header{
				"Cookie": {fmt.Sprintf("refresh_token=%s", refreshToken)},
			}
			resp, err = helper.Client.Get(URL, header)
			require.NoError(t, err)
			require.EqualValues(t, 200, resp.StatusCode)
		}
	}

	// the first one cannot work
	// we have to test this one last because the server can misunderstood that
	// the token is reused or something
	URL = fmt.Sprintf("%s/refresh", helper.AuthURL)
	header = http.Header{
		"Cookie": {fmt.Sprintf("refresh_token=%s", refreshTokens[0])},
	}
	resp, err = helper.Client.Get(URL, header)
	require.NoError(t, err)
	require.EqualValues(t, 401, resp.StatusCode)
}
