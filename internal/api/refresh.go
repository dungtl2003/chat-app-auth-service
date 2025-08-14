package api

import (
	"dungtl2003/chat-app-auth-service/internal/constants"
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	u "dungtl2003/chat-app-auth-service/internal/services/user"
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RefreshTokenResponseBody struct {
	AccessToken string          `json:"access_token"`
	SessionId   types.JsonInt64 `json:"session_id"`
	User        model.ChatUser  `json:"user"`
}

type ExpiredTokenError struct {
	UserId    int64
	SessionId int64
}

type InvalidTokenError struct {
}

func (e ExpiredTokenError) Error() string {
	return "expired token"
}

func (e InvalidTokenError) Error() string {
	return "invalid token"
}

// Refresh logic flow:
// 1. If the refresh token is missing, return 401.
// 2. If the refresh token is invalid, delete the cookie and return 401.
// 3. If the refresh token is expired, revoke the session, delete the cookie, and
// return 401.
// 4. If the session version in the refresh token is less than the session version
// in the database, revoke the session, delete the cookie and return 401 (session
// version is the amount of times server detects that the user uses the same
// refresh token more than once. If the session version is less than the one
// in the database, it means that the user still uses the old valid refresh token).
// 5. If the corresponding session is found but already revoked (or not found),
// revoke all sessions belonged to that user, increase session version, delete
// the cookie, and return 401 (in current logic, if the session is not revoked,
// user service MUST return that session. If not, it means that the session is revoked).
// 6. If the token and the hash are not matched, revoke the session, delete the
// cookie and return 401.
func Refresh(appCtx *context.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		refreshTokenStr, err := c.Cookie("refresh_token")
		if err != nil {
			if err == http.ErrNoCookie {
				appCtx.Logger.Errorfln("Missing refresh token")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token"})
			} else {
				appCtx.Logger.Errorfln("Cookie(): %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("Refresh token: %s", refreshTokenStr)

		// Validate and parse the refresh token. If the token is invalid or
		// expired, it will clear the cookie. For expired token, it will also
		// revoke the session. In any case, it will return an error response.
		parsedToken, err := validateToken(refreshTokenStr, appCtx.JwtConfig.JwtSecret)
		if err != nil {
			switch e := err.(type) {
			case InvalidTokenError:
				appCtx.Logger.Errorfln("Invalid refresh token")
				helper.ClearCookie(c, "refresh_token", appCtx.DomainName)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			case ExpiredTokenError:
				appCtx.Logger.Errorfln("Expired refresh token")
				err = appCtx.UserService.RevokeSession(c, e.UserId, e.SessionId)
				if err != nil {
					appCtx.Logger.Warnfln("RevokeSession(): %v", e)
				}
				helper.ClearCookie(c, "refresh_token", appCtx.DomainName)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Expired refresh token"})
			default:
				appCtx.Logger.Errorfln("ParseToken(): %v", e)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}

			c.Abort()
			return
		}

		userId := parsedToken.UserId
		sessionVersion := parsedToken.SessionVersion
		sessId := parsedToken.SessionId

		// Get user information
		userResponse, err := appCtx.UserService.GetUserById(c, userId)
		if err != nil {
			appCtx.Logger.Errorfln("GetUserById(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if userResponse == nil {
			panic("GetUserById() should not return nil user response if there is no error")
		}
		if userResponse.Body.Error != "" {
			appCtx.Logger.Errorfln("GetUserById() returned error: %s", userResponse.Body.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": userResponse.Body.Error})
			c.Abort()
			return
		}

		// the account was attacked so this refresh token is not valid anymore
		user := userResponse.Body.User
		if sessionVersion < user.SessionVersion.Int64() {
			appCtx.Logger.Errorfln("Invalid session version (expected: %d, got: %d)", user.SessionVersion.Int64(), sessionVersion)
			err = appCtx.UserService.RevokeSession(c, userId, sessId)
			if err != nil {
				appCtx.Logger.Errorfln("RevokeSession(): %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
				return
			}

			helper.ClearCookie(c, "refresh_token", appCtx.DomainName)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session version", "code": constants.REFRESH_TOKEN_REUSE})
			c.Abort()
			return
		}

		// check if session is revoked
		session := model.Session{
			// we need to set this field default to valid time
			RevokedAt: types.NewJsonNullTime(time.Now()),
		}
		for _, s := range user.Sessions {
			if s.Id.Int64() == sessId {
				session = s
				break
			}
		}
		appCtx.Logger.Debugfln("Session: %#v", session)

		// Reuse detected!!!
		if session.RevokedAt.Valid {
			handleRefreshTokenReuse(appCtx, c, user)
			return
		}

		refreshTokenHash := helper.HashWithSHA256(refreshTokenStr)
		// check if refresh token hash is correct
		if refreshTokenHash != session.RefreshTokenHash {
			err = appCtx.UserService.RevokeSession(c, userId, sessId)
			if err != nil {
				appCtx.Logger.Errorfln("RevokeSession(): %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
				return
			}

			appCtx.Logger.Errorfln("invalid refresh token hash (expected: %s, got: %s)", session.RefreshTokenHash, refreshTokenHash)
			helper.ClearCookie(c, "refresh_token", appCtx.DomainName)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			c.Abort()
			return
		}

		// revoke session
		err = appCtx.UserService.RevokeSession(c, userId, sessId)
		if err != nil {
			appCtx.Logger.Errorfln("RevokeSession(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// remove revoked session from user
		user.Sessions = helper.Filter(user.Sessions, func(s model.Session) bool {
			return s.Id.Int64() != sessId
		})

		// create new session ID
		newSessId, err := appCtx.IdGeneratorService.GenerateId(c)
		if err != nil {
			appCtx.Logger.Errorfln("GenerateId(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// create new tokens
		newAccessTokenStr, newRefreshTokenStr, err := helper.CreateNewTokenPair(appCtx.JwtConfig.JwtSecret, user, appCtx.JwtConfig.ATDurationMs, appCtx.JwtConfig.RTDurationMs, newSessId)
		if err != nil {
			appCtx.Logger.Errorfln("CreateNewTokenPair(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("AT: %s", newAccessTokenStr)
		appCtx.Logger.Debugfln("RT: %s", newRefreshTokenStr)

		// create new session
		expiresAt, err := helper.GetTokenExpirationStr(newRefreshTokenStr, appCtx.JwtConfig.JwtSecret)
		if err != nil {
			appCtx.Logger.Errorfln("GetTokenExpirationStr(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		newRefreshTokenHash := helper.HashWithSHA256(newRefreshTokenStr)
		appCtx.Logger.Debugfln("New refresh token hash: %s", newRefreshTokenHash)
		newSessionResponse, err := appCtx.UserService.CreateSession(c, user.Id.Int64(), u.SessionPostRequestBody{
			Id:               newSessId,
			Version:          user.SessionVersion.Int64(),
			DeviceInfo:       session.DeviceInfo,
			RefreshTokenHash: newRefreshTokenHash,
			ExpiresAt:        expiresAt,
		})
		if err != nil {
			appCtx.Logger.Errorfln("CreateSession(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if newSessionResponse == nil {
			panic("CreateSession() should not return nil response if there is no error")
		}
		if newSessionResponse.Body.Error != "" {
			appCtx.Logger.Errorfln("CreateSession() returned error: %s", newSessionResponse.Body.Error)
			c.JSON(newSessionResponse.StatusCode, gin.H{"error": newSessionResponse.Body.Error})
			c.Abort()
			return
		}

		// add new session to user
		newSession := newSessionResponse.Body.Session
		user.Sessions = append(user.Sessions, newSession)

		// set cookie
		refreshResponseBody := RefreshTokenResponseBody{
			AccessToken: newAccessTokenStr,
			SessionId:   types.NewJsonInt64(newSessId),
			User:        user,
		}
		appCtx.Logger.Debugfln("New session created: %d", newSessId)
		helper.SetCookieOverride(c, "refresh_token", newRefreshTokenStr, appCtx.DomainName, appCtx.JwtConfig.RTDurationMs)
		c.JSON(http.StatusOK, refreshResponseBody)
	}
}

// handleRefreshTokenReuse is a function that handles the case when the refresh token
// is reused. It revokes all sessions of the user and increments the session version.
// It also clears the refresh token cookie and returns a 401 response.
func handleRefreshTokenReuse(appCtx *context.AppContext, c *gin.Context, user model.ChatUser) {
	// maybe stolen by someone. Regardless, REVOKE ALL!!!
	err := appCtx.UserService.RevokeAllSessions(c, user.Id.Int64())
	if err != nil {
		appCtx.Logger.Errorfln("RevokeAllSessions(): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return
	}

	// update session version
	err = appCtx.UserService.IncrementSessionVersion(c, user.Id.Int64())
	if err != nil {
		appCtx.Logger.Errorfln("IncrementSessionVersion(): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return
	}

	appCtx.Logger.Errorfln("The account might be attacked")
	helper.ClearCookie(c, "refresh_token", appCtx.DomainName)
	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token", "code": constants.REFRESH_TOKEN_REUSE})
	c.Abort()
}

// validateToken is a function that handles both parsing and validating refresh
// token. It decodes the token twice: first without claims validation and then
// with claims validation. If the first decoding fails, it means that the token
// is invalid. If the second decoding fails, it means that the token is expired.
// In both cases, it returns error. Else, it returns the parsed token.
// In case of an expired token, it also revokes the session. Note that we assume
// all token's fields are valid, and only expires field can be invalid.
func validateToken(tokenStr string, secret string) (*helper.ParsedToken, error) {
	// decode without claims validation first
	token, err := jwthandler.DecodeTokenWithoutClaimsValidation(secret, tokenStr)
	if err != nil {
		return nil, InvalidTokenError{}
	}

	// parse token
	parsedToken, err := helper.ParseToken(token)
	if err != nil {
		return nil, fmt.Errorf("ParseToken(): %w", err)
	}

	// decode with claims validation
	_, err = jwthandler.DecodeToken(secret, tokenStr)
	if err != nil {
		return nil, ExpiredTokenError{}
	}

	return parsedToken, nil
}
