package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSignUpSuccessShouldAutoLogin(t *testing.T) {
	helper := NewHelper()
	err := helper.Snapshot()
	require.NoError(t, err)
	defer func() {
		err := helper.Rollback()
		require.NoError(t, err)
	}()

	email := "you@example.com"
	username := "you"
	password := "password"
	role := "USER"
	deviceInfo := json.RawMessage(`{"user-agent": "Mozilla/5.0"}`)

	payloadJson := fmt.Appendf(nil, `
		{
			"email": "%s",
			"username": "%s",
			"password": "%s",
			"role": "%s",
			"device_info": %s
		}`, email, username, password, role, deviceInfo)

	URL := fmt.Sprintf("%s/signup", helper.AuthURL)
	header := http.Header{
		"Content-Type": []string{"application/json"},
	}

	resp, err := helper.Client.Post(URL, header, bytes.NewBuffer(payloadJson))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// cookie: "refresh_token": "%s"
	refreshToken := GetRTFromResponse(resp)
	require.NotEmpty(t, refreshToken)

	respJson, err := GetRespJson(resp)
	require.NoError(t, err)
	require.NotEmpty(t, respJson["user"])
	require.NotEmpty(t, respJson["session_id"])
	require.NotEmpty(t, respJson["access_token"])
}
