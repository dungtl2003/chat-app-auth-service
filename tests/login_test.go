package tests

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/api"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/services/database"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	USERS__LOGIN_TEST_FILENAME = "users__login_test.json"
)

func TestLoginReturnCorrectStatus(t *testing.T) {
	helper := NewTestHelper()
	SetUp(helper, &SetUpOptions{
		DataFile: &database.DataFile{
			UserFile: USERS__LOGIN_TEST_FILENAME,
		},
	})
	defer TearDown(helper)

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

			URL := fmt.Sprintf("%s/auth/login", helper.AuthURL)

			resp, err := Post(helper.Client, URL, nil, bytes.NewBuffer(payloadJson))
			require.NoError(t, err)
			require.EqualValues(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestLoginSuccessfully(t *testing.T) {
	helper := NewTestHelper()
	SetUp(helper, &SetUpOptions{
		DataFile: &database.DataFile{
			UserFile: USERS__LOGIN_TEST_FILENAME,
		},
	})
	defer TearDown(helper)

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

	URL := fmt.Sprintf("%s/auth/login", helper.AuthURL)

	resp, err := Post(helper.Client, URL, nil, bytes.NewBuffer(payloadJson))
	require.NoError(t, err)
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	// cookie: "refresh_token": "%s"
	refreshToken := GetRTFromResponse(resp)
	require.NotEmpty(t, refreshToken)

	// validate RT
	decodedRT, err := jwthandler.DecodeToken(helper.JwtSecret, refreshToken)
	require.NoError(t, err)
	claims := decodedRT.Claims.(*jwthandler.UserJWTClaim)
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

	var responseBody api.LoginResponseBody
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	require.NoError(t, err)
	accessToken := responseBody.AccessToken
	require.NotEmpty(t, accessToken)

	// validate AT
	decodedRT, err = jwthandler.DecodeToken(helper.JwtSecret, accessToken)
	require.NoError(t, err)
	claims = decodedRT.Claims.(*jwthandler.UserJWTClaim)
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
