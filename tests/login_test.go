package tests

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/httpclient"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoginReturnCorrectStatus(t *testing.T) {
	helper := NewHelper()
	err := helper.Snapshot()
	require.NoError(t, err)
	defer func() {
		err := helper.Rollback()
		require.NoError(t, err)
	}()

	// we will use user's information in fixed json file
	testcases := []struct {
		identifier string
		password   string
		deviceId   string
		status     int
	}{
		{
			identifier: "normaluser@gmail.com",
			password:   "normalpassword",
			deviceId:   "1",
			status:     200,
		},
		{
			identifier: "normaluser",
			password:   "normalpassword",
			deviceId:   "1",
			status:     200,
		},
		{
			identifier: "normaluser",
			password:   "normalpassword",
			deviceId:   "notanumber",
			status:     400,
		},
		{
			identifier: "normaluser",
			password:   "normalpassword2",
			deviceId:   "1",
			status:     403,
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("identifier: %s, password: %s, deviceId: %s, status: %d", tc.identifier, tc.password, tc.deviceId, tc.status), func(t *testing.T) {
			payloadJson := fmt.Appendf(nil, `
			{
				"identifier": "%s",
				"password": "%s",
				"device_id": "%s"
			}`, tc.identifier, tc.password, tc.deviceId)

			URL := fmt.Sprintf("%s/login", helper.AuthURL)

			resp, err := helper.Client.Post(URL, nil, bytes.NewBuffer(payloadJson))
			require.NoError(t, err)
			require.EqualValues(t, tc.status, resp.StatusCode)
		})
	}
}

func TestLoginCorrectDataSuccessfully(t *testing.T) {
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
	role := "USER"

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

	// cookie: "refresh_token": "%s"
	cookies := resp.Cookies()
	require.NotEmpty(t, cookies)
	// we will find the refresh token in the cookie
	var refreshToken string
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshToken = cookie.Value
			break
		}
	}

	// validate RT
	tok, err := jwthandler.DecodeToken(helper.JwtSecret, refreshToken)
	require.NoError(t, err)
	sub, err := tok.Claims.GetSubject()
	require.NoError(t, err)
	require.EqualValues(t, identifier, sub)
	aud, err := tok.Claims.GetAudience()
	require.NoError(t, err)
	require.EqualValues(t, role, aud[0])

	// body: "access_token: %s"
	respBody, err := httpclient.ReadResponse(resp)
	require.NoError(t, err)
	accessTokenPrefix := `"access_token: `
	accessTokenSuffix := `"`
	accessToken := string(respBody)
	accessToken = accessToken[len(accessTokenPrefix):]
	accessToken = accessToken[:len(accessToken)-len(accessTokenSuffix)]

	// validate AT
	tok, err = jwthandler.DecodeToken(helper.JwtSecret, accessToken)
	require.NoError(t, err)
	sub, err = tok.Claims.GetSubject()
	require.NoError(t, err)
	require.EqualValues(t, identifier, sub)
	aud, err = tok.Claims.GetAudience()
	require.NoError(t, err)
	require.EqualValues(t, role, aud[0])
}
