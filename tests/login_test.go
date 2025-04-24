package tests

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"fmt"
	"net/http"
	"strconv"
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
		deviceName string
		deviceType string
		os         string
		status     int
	}{
		{
			identifier: "normaluser@gmail.com",
			password:   "normalpassword",
			deviceId:   "1",
			status:     http.StatusOK,
		},
		{
			identifier: "normaluser",
			password:   "normalpassword",
			deviceId:   "1",
			status:     http.StatusOK,
		},
		{
			identifier: "normaluser",
			password:   "normalpassword",
			deviceName: "device_name",
			deviceType: "device_type",
			os:         "os",
			status:     http.StatusOK,
		},
		{
			identifier: "normaluser",
			password:   "normalpassword",
			deviceId:   "notanumber",
			status:     http.StatusBadRequest,
		},
		{
			identifier: "normaluser",
			password:   "normalpassword2",
			deviceId:   "1",
			status:     http.StatusForbidden,
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("identifier: %s, password: %s, deviceId: %s, deviceName: %s, deviceType: %s, os: %s", tc.identifier, tc.password, tc.deviceId, tc.deviceName, tc.deviceType, tc.os), func(t *testing.T) {
			payloadJson := fmt.Appendf(nil, `
			{
				"identifier": "%s",
				"password": "%s",
				"device": {
					"id": "%s",
					"device_name": "%s",
					"device_type": "%s",
					"os": "%s"
				}
			}`, tc.identifier, tc.password, tc.deviceId, tc.deviceName, tc.deviceType, tc.os)

			URL := fmt.Sprintf("%s/login", helper.AuthURL)

			resp, err := helper.Client.Post(URL, nil, bytes.NewBuffer(payloadJson))
			require.NoError(t, err)
			require.EqualValues(t, tc.status, resp.StatusCode)
		})
	}
}

func TestLoginCorrectDeviceIdSuccessfully(t *testing.T) {
	helper := NewHelper()
	err := helper.Snapshot()
	require.NoError(t, err)
	defer func() {
		err := helper.Rollback()
		require.NoError(t, err)
	}()

	uid := 2
	identifier := "normaluser"
	password := "normalpassword"
	deviceId := 1
	role := "USER"

	payloadJson := fmt.Appendf(nil, `
			{
				"identifier": "%s",
				"password": "%s",
				"device": {
					"id": "%d"
				}
			}`, identifier, password, deviceId)

	URL := fmt.Sprintf("%s/login", helper.AuthURL)

	resp, err := helper.Client.Post(URL, nil, bytes.NewBuffer(payloadJson))
	require.NoError(t, err)
	require.EqualValues(t, 200, resp.StatusCode)

	// cookie: "refresh_token": "%s"
	refreshToken := GetRTFromResponse(resp)
	require.NotEmpty(t, refreshToken)

	// validate RT
	tok, err := jwthandler.DecodeToken(helper.JwtSecret, refreshToken)
	claims := tok.Claims.(*jwthandler.JWTClaim)
	require.NoError(t, err)
	subStr, err := claims.GetSubject()
	require.NoError(t, err)
	sub, err := strconv.Atoi(subStr)
	require.NoError(t, err)
	require.EqualValues(t, uid, sub)
	aud, err := claims.GetAudience()
	require.NoError(t, err)
	require.EqualValues(t, role, aud[0])

	// body: "access_token: %s"
	respJson, err := GetRespJson(resp)
	require.NoError(t, err)
	accessToken := respJson["access_token"].(string)
	require.NotEmpty(t, accessToken)
	currDevId := int64(respJson["current_device_id"].(float64))
	require.EqualValues(t, deviceId, currDevId)

	// validate AT
	tok, err = jwthandler.DecodeToken(helper.JwtSecret, accessToken)
	require.NoError(t, err)
	subStr, err = tok.Claims.GetSubject()
	require.NoError(t, err)
	sub, err = strconv.Atoi(subStr)
	require.NoError(t, err)
	require.EqualValues(t, uid, sub)
	aud, err = tok.Claims.GetAudience()
	require.NoError(t, err)
	require.EqualValues(t, role, aud[0])
}

func TestLoginWithDeviceInfoSuccessfully(t *testing.T) {
	helper := NewHelper()
	err := helper.Snapshot()
	require.NoError(t, err)
	defer func() {
		err := helper.Rollback()
		require.NoError(t, err)
	}()

	uid := 2
	identifier := "normaluser"
	password := "normalpassword"
	role := "USER"
	deviceName := "device_name"
	deviceType := "device_type"
	os := "os"

	payloadJson := fmt.Appendf(nil, `
			{
				"identifier": "%s",
				"password": "%s",
				"device": {
					"device_name": "%s",
					"device_type": "%s",
					"os": "%s"
				}
			}`, identifier, password, deviceName, deviceType, os)

	URL := fmt.Sprintf("%s/login", helper.AuthURL)

	resp, err := helper.Client.Post(URL, nil, bytes.NewBuffer(payloadJson))
	require.NoError(t, err)
	require.EqualValues(t, 200, resp.StatusCode)

	// cookie: "refresh_token": "%s"
	refreshToken := GetRTFromResponse(resp)
	require.NotEmpty(t, refreshToken)

	// validate RT
	tok, err := jwthandler.DecodeToken(helper.JwtSecret, refreshToken)
	claims := tok.Claims.(*jwthandler.JWTClaim)
	require.NoError(t, err)
	subStr, err := claims.GetSubject()
	require.NoError(t, err)
	sub, err := strconv.Atoi(subStr)
	require.NoError(t, err)
	require.EqualValues(t, uid, sub)
	aud, err := claims.GetAudience()
	require.NoError(t, err)
	require.EqualValues(t, role, aud[0])

	// body: "access_token: %s"
	respJson, err := GetRespJson(resp)
	require.NoError(t, err)
	accessToken := respJson["access_token"].(string)
	require.NotEmpty(t, accessToken)
	currDevId := int64(respJson["current_device_id"].(float64))

	// validate AT
	tok, err = jwthandler.DecodeToken(helper.JwtSecret, accessToken)
	require.NoError(t, err)
	subStr, err = tok.Claims.GetSubject()
	require.NoError(t, err)
	sub, err = strconv.Atoi(subStr)
	require.NoError(t, err)
	require.EqualValues(t, uid, sub)
	aud, err = tok.Claims.GetAudience()
	require.NoError(t, err)
	require.EqualValues(t, role, aud[0])

	err = helper.Db.client.QueryRow("SELECT id FROM chat_user.device WHERE id = $1", currDevId).Scan(&currDevId)
	require.NoError(t, err)
}
