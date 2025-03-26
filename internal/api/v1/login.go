package v1

import (
	"dungtl2003/chat-app-auth-service/internal/config"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/httpclient"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
)

type LoginRequestBody struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

func Login(c *gin.Context, logger *slog.Logger, validator *helper.Validator, client *httpclient.HttpClient, authEndpoint string, jwtConfig config.JwtTokenConfig, domainName string) {
	l := helper.NewLoggerWrapper(logger)
	var loginRequestBody LoginRequestBody

	if err := c.ShouldBindJSON(&loginRequestBody); err != nil {
		l.Errorf("error while binding request body: %v", err)
		c.AbortWithStatus(400)
		return
	}

	l.Debugf("login request body: %#v", loginRequestBody)
	url := fmt.Sprintf("%s?identifier=%s", authEndpoint, loginRequestBody.Identifier)
	l.Debugf("sending GET request to %s", url)
	resp, err := client.Get(url, c.Request.Header)
	if err != nil {
		l.Errorf("error when sending GET request: %v", err)
		c.AbortWithStatus(500)
		return
	}
	if resp.StatusCode != 200 {
		l.Errorf("error GET request status: %d", resp.StatusCode)
		c.AbortWithStatus(500)
		return
	}

	body, err := httpclient.ReadResponse(resp)
	if err != nil {
		l.Errorf("error when reading response: %v", err)
		c.AbortWithStatus(500)
		return
	}

	var user model.ChatUser
	err = helper.ParseAsJson(body, &user)
	if err != nil {
		l.Errorf("error when parsing body: %v", err)
		c.AbortWithStatus(500)
		return
	}
	l.Debugf("response body from GET request: %#v", user)

	accessToken, err := jwthandler.CreateToken(jwtConfig.JwtSecret, user, jwtConfig.ATDurationMs)
	if err != nil {
		l.Errorf("error when creating access token: %v", err)
		c.AbortWithStatus(500)
		return
	}
	l.Debugf("AT: %s", accessToken)

	refreshToken, err := jwthandler.CreateToken(jwtConfig.JwtSecret, user, jwtConfig.RTDurationMs)
	if err != nil {
		l.Errorf("error when creating refresh token: %v", err)
		c.AbortWithStatus(500)
		return
	}
	l.Debugf("RT: %s", refreshToken)

	c.SetCookie("refresh_token", refreshToken, int(jwtConfig.RTDurationMs/1000), "/api/v1/login", domainName, false, true)
	c.JSON(200, fmt.Sprintf("access_token: %s", accessToken))
}
