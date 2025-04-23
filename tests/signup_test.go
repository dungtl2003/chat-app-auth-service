package tests

import (
	"bytes"
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
	deviceName := "IPhone 14 Pro Max"
	deviceType := "IPhone"
	deviceOs := "IOS 16.5"

	payloadJson := fmt.Appendf(nil, `
		{
			"email": "%s",
			"username": "%s",
			"password": "%s",
			"role": "%s",
			"device": {
				"device_name": "%s",
				"device_type": "%s",
				"os": "%s"
			}
		}`, email, username, password, role, deviceName, deviceType, deviceOs)

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

	// body: "access_token: %s"
	accessToken := GetATFromResponse(resp)
	require.NotEmpty(t, accessToken)
}
