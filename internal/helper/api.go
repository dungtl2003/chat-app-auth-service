package helper

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/httpclient"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserPostPayload struct {
	Email    string
	Username string
	Password string
	Role     model.UserRole
}

type SessionPostPayload struct {
	Id               int64
	Version          int64
	DeviceInfo       types.Json
	RefreshTokenHash string
	ExpiresAt        string
}

type ParsedToken struct {
	UserId         int64
	SessionVersion int64
	SessionId      int64
}

// HandleGetUserAuth is a function that handles the retrieval of user information.
// It sends a GET request to the user service with the identifier (email or username).
// If the request fails or the response status code is not 200, it returns nil.
// If the request is successful, it parses the response and returns the user information.
// This function is meant to be used for authentication purposes (checking password).
func HandleGetUserAuth(appCtx *context.AppContext, c *gin.Context, identifier string) *model.ChatUser {
	url := EncodeURLPath(fmt.Sprintf(`%s/users/auth-info?identifier=%s`, appCtx.UserServiceURL, identifier))
	appCtx.Logger.Debugfln("sending GET request to %s", url)
	resp, err := appCtx.Client.Get(url, nil)
	if err != nil {
		appCtx.Logger.Errorfln("error when sending GET request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		c.Status(resp.StatusCode)
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			appCtx.Logger.Errorfln("Copy(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
		}

		return nil
	}
	var user model.ChatUser
	err = httpclient.ParseResponse(resp, &user)
	if err != nil {
		appCtx.Logger.Errorfln("ParseResponse(): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return nil
	}
	appCtx.Logger.Debugfln("user from GET request: %#v", user)

	return &user
}

// HandleGetUserSecure is a function that handles the retrieval of user information.
// It sends a GET request to the user service with the user ID.
// If the request fails or the response status code is not 200, it returns nil.
// If the request is successful, it parses the response and returns the user information.
// It also clears the password field for security reasons.
func HandleGetUserSecure(appCtx *context.AppContext, c *gin.Context, userId int64) *model.ChatUser {
	url := EncodeURLPath(fmt.Sprintf("%s/users/%d", appCtx.UserServiceURL, userId))
	resp, err := appCtx.Client.Get(url, nil)
	if err != nil {
		appCtx.Logger.Errorfln("error when sending GET request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		appCtx.Logger.Errorfln("error GET request status: %d", resp.StatusCode)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return nil
	}
	var user model.ChatUser
	err = httpclient.ParseResponse(resp, &user)
	if err != nil {
		appCtx.Logger.Errorfln("ParseResponse(): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return nil
	}
	// security
	user.Password = ""
	appCtx.Logger.Debugfln("user from GET request: %#v", user)

	return &user
}

func HandleIncrementSessionVersion(appCtx *context.AppContext, c *gin.Context, userId int64) bool {
	url := EncodeURLPath(fmt.Sprintf("%s/users/%d/session-version/increment", appCtx.UserServiceURL, userId))
	appCtx.Logger.Debugfln("sending POST request to %s", url)
	resp, err := appCtx.Client.Post(url, nil, nil)
	if err != nil {
		appCtx.Logger.Errorfln("error when sending PATCH request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return false
	}
	if resp.StatusCode != http.StatusOK {
		c.Status(resp.StatusCode)
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			appCtx.Logger.Errorfln("Copy(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}

		c.Abort()
		return false
	}

	return true
}

// HandleRevokeAllSessions is a function that handles the revocation of all sessions for a user.
// It sends a POST request to the user service to revoke all sessions.
// If the request fails or the response status code is not 200, it returns false (not ok).
// If the request is successful, it clears the refresh token cookie and returns true (ok).
// This function is meant to be used when the refresh token is reused or compromised.
func HandleRevokeAllSessions(appCtx *context.AppContext, c *gin.Context, userId int64) bool {
	url := EncodeURLPath(fmt.Sprintf("%s/users/%d/sessions/revoke", appCtx.UserServiceURL, userId))
	appCtx.Logger.Debugfln("sending POST request to %s", url)
	resp, err := appCtx.Client.Post(url, nil, nil)
	if err != nil {
		appCtx.Logger.Errorfln("error when sending POST request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return false
	}
	if resp.StatusCode != http.StatusOK {
		c.Status(resp.StatusCode)
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			appCtx.Logger.Errorfln("Copy(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}

		c.Abort()
		return false
	}

	ClearCookie(c, "refresh_token", appCtx.DomainName)
	return true
}

// handleRevokeSession is a function that handles the revocation of a session.
// It sends a POST request to the user service to revoke the session.
// If the request fails or the response status code is not 200, it returns false (not ok).
// If the request is successful, it clears the refresh token cookie and returns true (ok).
func HandleRevokeSession(appCtx *context.AppContext, c *gin.Context, userId int64, sessId int64) bool {
	// revoke session
	url := EncodeURLPath(fmt.Sprintf("%s/users/%d/sessions/%d/revoke", appCtx.UserServiceURL, userId, sessId))
	appCtx.Logger.Debugfln("sending POST request to %s", url)
	resp, err := appCtx.Client.Post(url, nil, nil)
	if err != nil {
		appCtx.Logger.Errorfln("error when sending POST request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return false
	}
	if resp.StatusCode != http.StatusOK {
		c.Status(resp.StatusCode)
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			appCtx.Logger.Errorfln("Copy(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		c.Abort()
		return false
	}

	ClearCookie(c, "refresh_token", appCtx.DomainName)
	return true
}

func HandleCreateSession(appCtx *context.AppContext, c *gin.Context, data SessionPostPayload, userId int64) *model.Session {
	url := EncodeURLPath(fmt.Sprintf("%s/users/%d/sessions", appCtx.UserServiceURL, userId))
	payload := fmt.Appendf(nil, `
		{
			"id": "%d",
			"version": "%d",
			"device_info": %s,
			"refresh_token_hash": "%s",
			"expires_at": "%s"
		}`, data.Id, data.Version, data.DeviceInfo, data.RefreshTokenHash, data.ExpiresAt)
	resp, err := appCtx.Client.Post(url, nil, bytes.NewBuffer(payload))
	if err != nil {
		appCtx.Logger.Errorfln("error when sending POST request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return nil
	}
	if resp.StatusCode != http.StatusCreated {
		c.Status(resp.StatusCode)
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			appCtx.Logger.Errorfln("Copy(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
		}
		return nil
	}
	var session model.Session
	err = httpclient.ParseResponse(resp, &session)
	if err != nil {
		appCtx.Logger.Errorfln("ParseResponse(): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return nil
	}
	appCtx.Logger.Debugfln("session from POST request: %#v", session)
	return &session
}

// HandleCreateUser is a function that handles the creation of a user.
// It sends a POST request to the user service with the user information.
// If the request fails or the response status code is not 201, it returns nil.
func HandleCreateUser(appCtx *context.AppContext, c *gin.Context, data UserPostPayload) *model.ChatUser {
	payload := fmt.Appendf(nil, `
		{
			"email": "%s",
			"username": "%s",
			"password": "%s",
			"role": "%s"
		}`, data.Email, data.Username, data.Password, data.Role)
	url := EncodeURLPath(fmt.Sprintf("%s/users", appCtx.UserServiceURL))
	appCtx.Logger.Debugfln("sending POST request to %s with payload: %s", url, StripWS(string(payload)))
	resp, err := appCtx.Client.Post(url, nil, bytes.NewBuffer(payload))
	if err != nil {
		appCtx.Logger.Errorfln("error when sending POST request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return nil
	}
	if resp.StatusCode != http.StatusCreated {
		c.Status(resp.StatusCode)
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			appCtx.Logger.Errorfln("Copy(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
		}

		return nil
	}
	var user model.ChatUser
	err = httpclient.ParseResponse(resp, &user)
	if err != nil {
		appCtx.Logger.Errorfln("ParseResponse(): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return nil
	}
	appCtx.Logger.Debugfln("user from POST request: %#v", user)
	return &user
}

// ParseToken is a function that parses the JWT token and returns the parsed token.
// It extracts the user ID, session version, and session ID from the token claims.
// If any of these values cannot be extracted, it returns an error.
func ParseToken(tok *jwt.Token) (*ParsedToken, error) {
	claims := tok.Claims.(*jwthandler.UserJWTClaim)
	sessionVersion, err := claims.GetSessVersion()
	if err != nil {
		return nil, fmt.Errorf("GetSessVersion(): %v", err)
	}
	userIdStr, err := claims.GetSubject()
	if err != nil {
		return nil, fmt.Errorf("GetSubject(): %v", err)
	}
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("strconv.ParseInt(): %v", err)
	}
	sessId, err := claims.GetSessId()
	if err != nil {
		return nil, fmt.Errorf("GetSessId(): %v", err)
	}

	return &ParsedToken{
		UserId:         userId,
		SessionVersion: sessionVersion,
		SessionId:      sessId,
	}, nil
}

func ClearCookie(c *gin.Context, cookieName string, domainName string) {
	c.SetCookie(cookieName, "", -1, "/", domainName, false, false)
}

func SetCookie(c *gin.Context, cookieName string, cookieValue string, domainName string, durationMs int64) {
	c.SetCookie(cookieName, cookieValue, int(durationMs/1000), "/", domainName, false, false)
}

func SetCookieOverride(c *gin.Context, name, value string, domain string, durationMs int64) {
	header := c.Writer.Header()
	prev := header["Set-Cookie"]
	prefix := name + "=" // cookie lines all start "name="
	filtered := Filter(prev, func(h string) bool {
		return !strings.HasPrefix(h, prefix)
	})
	header["Set-Cookie"] = filtered

	// now add the new one
	SetCookie(c, name, value, domain, durationMs)
}
