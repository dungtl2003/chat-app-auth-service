package tests

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/api"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/services/database"
	"dungtl2003/chat-app-auth-service/internal/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	USERS__CHECK_TEST_FILENAME = "users__check_test.json"
)

func TestCheckReturnCorrectStatus(t *testing.T) {
	helper := NewTestHelper()
	SetUp(helper, &SetUpOptions{
		DataFile: &database.DataFile{
			UserFile: USERS__CHECK_TEST_FILENAME,
		},
	})
	defer TearDown(helper)

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

			URL := fmt.Sprintf("%s/auth/check", helper.AuthURL)
			header := http.Header{
				"Authorization": {tc.authHeader},
			}

			resp, err := Get(helper.Client, URL, header)
			require.NoError(t, err)
			require.EqualValues(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestCheckWithRealToken(t *testing.T) {
	helper := NewTestHelper()
	SetUp(helper, &SetUpOptions{
		DataFile: &database.DataFile{
			UserFile: USERS__CHECK_TEST_FILENAME,
		},
	})
	defer TearDown(helper)

	requestPayload := api.LoginRequestBody{
		Identifier: "normaluser",
		Password:   "normalpassword",
		DeviceInfo: types.NewJson([]byte(`{"user-agent": "Mozilla/5.0"}`)),
	}
	bodyBytes, err := json.Marshal(requestPayload)
	require.NoError(t, err)

	URL := fmt.Sprintf("%s/auth/login", helper.AuthURL)
	resp, err := Post(helper.Client, URL, nil, io.Reader(bytes.NewBuffer(bodyBytes)))
	require.NoError(t, err)
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	var responseBody types.Response[api.LoginResponseBody]
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	require.NoError(t, err)
	require.Nil(t, responseBody.Error)
	require.NotNil(t, responseBody.Data)

	accessToken := responseBody.Data.Item.AccessToken
	require.NotEmpty(t, accessToken)

	URL = fmt.Sprintf("%s/auth/check", helper.AuthURL)
	header := http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", accessToken)},
	}
	resp, err = Get(helper.Client, URL, header)
	require.NoError(t, err)
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	helper.Logger.Debugfln("Current time: %v", time.Now())
	<-time.After(time.Duration(helper.ATDurationMs) * time.Millisecond)
	helper.Logger.Debugfln("After wait time: %v", time.Now())
	header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", accessToken)},
	}
	resp, err = Get(helper.Client, URL, header)
	require.NoError(t, err)
	// token should be expired
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCheckWithInternalTokenShouldFail(t *testing.T) {
	issuer := os.Getenv("TOKEN_ISSUER")
	require.NotEmpty(t, issuer)
	internalAud := os.Getenv("TOKEN_INTERNAL_AUDIENCE")
	require.NotEmpty(t, internalAud)
	helper := NewTestHelper()
	SetUp(helper, &SetUpOptions{
		DataFile: &database.DataFile{
			UserFile: USERS__CHECK_TEST_FILENAME,
		},
	})
	defer TearDown(helper)

	internalToken, err := jwthandler.CreateInternalToken(helper.JwtSecret, 5_000_000, 2, issuer, internalAud)
	require.NoError(t, err)

	URL := fmt.Sprintf("%s/auth/check", helper.AuthURL)
	header := http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", internalToken)},
	}

	resp, err := Get(helper.Client, URL, header)
	require.NoError(t, err)
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)
}
