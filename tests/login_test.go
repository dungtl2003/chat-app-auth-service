package tests

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"encoding/json"
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
		identifier     string
		password       string
		deviceInfo     json.RawMessage
		expectedStatus int
	}{
		{
			identifier:     "normaluser@gmail.com",
			password:       "normalpassword",
			deviceInfo:     json.RawMessage(`{"user-agent": "Mozilla/5.0"}`),
			expectedStatus: http.StatusOK,
		},
		{
			identifier:     "normaluser",
			password:       "normalpassword",
			deviceInfo:     json.RawMessage(`{"user-agent": "Mozilla/5.0"}`),
			expectedStatus: http.StatusOK,
		},
		{
			// 2 letters
			identifier:     "john doe",
			password:       "normalpassword4",
			deviceInfo:     json.RawMessage(`{"user-agent": "Mozilla/5.0"}`),
			expectedStatus: http.StatusOK,
		},
		{
			password:       "normalpassword",
			deviceInfo:     json.RawMessage(`{"user-agent": "Mozilla/5.0"}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			identifier:     "normaluser",
			deviceInfo:     json.RawMessage(`{"user-agent": "Mozilla/5.0"}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			identifier:     "normaluser",
			password:       "normalpassword2",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("identifier: %s, password: %s, deviceInfo: %s", tc.identifier, tc.password, string(tc.deviceInfo)), func(t *testing.T) {
			payloadJson := fmt.Appendf(nil, `
			{
				"identifier": "%s",
				"password": "%s",
				"device_info": %s
			}`, tc.identifier, tc.password, tc.deviceInfo)

			URL := fmt.Sprintf("%s/login", helper.AuthURL)

			resp, err := helper.Client.Post(URL, nil, bytes.NewBuffer(payloadJson))
			require.NoError(t, err)
			require.EqualValues(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestLoginSuccessfully(t *testing.T) {
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
	role := model.USER
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
	require.EqualValues(t, 200, resp.StatusCode)

	// cookie: "refresh_token": "%s"
	refreshToken := GetRTFromResponse(resp)
	require.NotEmpty(t, refreshToken)

	// validate RT
	tok, err := jwthandler.DecodeToken(helper.JwtSecret, refreshToken)
	require.NoError(t, err)
	claims := tok.Claims.(*jwthandler.UserJWTClaim)
	subStr, err := claims.GetSubject()
	require.NoError(t, err)
	sub, err := strconv.Atoi(subStr)
	require.NoError(t, err)
	require.EqualValues(t, uid, sub)
	aud, err := claims.GetAudience()
	require.NoError(t, err)
	require.EqualValues(t, jwthandler.FRONTEND_AUDIENCE, aud[0])
	r, err := claims.GetRole()
	require.NoError(t, err)
	require.EqualValues(t, role, r)

	respJson, err := GetRespJson(resp)
	require.NoError(t, err)
	require.NotEmpty(t, respJson["user"])
	require.NotEmpty(t, respJson["session_id"])
	accessToken := respJson["access_token"].(string)
	require.NotEmpty(t, accessToken)

	// validate AT
	tok, err = jwthandler.DecodeToken(helper.JwtSecret, accessToken)
	require.NoError(t, err)
	claims = tok.Claims.(*jwthandler.UserJWTClaim)
	subStr, err = claims.GetSubject()
	require.NoError(t, err)
	sub, err = strconv.Atoi(subStr)
	require.NoError(t, err)
	require.EqualValues(t, uid, sub)
	aud, err = claims.GetAudience()
	require.NoError(t, err)
	require.EqualValues(t, jwthandler.FRONTEND_AUDIENCE, aud[0])
	r, err = claims.GetRole()
	require.NoError(t, err)
	require.EqualValues(t, role, r)
}
