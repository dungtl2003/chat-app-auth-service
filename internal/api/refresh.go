package api

import (
	"dungtl2003/chat-app-auth-service/internal/constants"
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

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
				appCtx.Logger.Debug("missing refresh token")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token"})
				c.Abort()
				return
			} else {
				appCtx.Logger.Errorfln("Cookie(): %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
				return
			}
		}

		// validate and parse the refresh token
		parsedToken := handleRefreshTokenValidation(appCtx, c, refreshTokenStr)
		if parsedToken == nil {
			return
		}
		userId := parsedToken.UserId
		sessionVersion := parsedToken.SessionVersion
		sessId := parsedToken.SessionId

		// 	get user information
		user := helper.HandleGetUserSecure(appCtx, c, userId)
		if user == nil {
			return
		}

		// the account was attacked so this refresh token is not valid anymore
		if sessionVersion < user.SessionVersion.Int64() {
			ok := helper.HandleRevokeSession(appCtx, c, userId, sessId)
			if ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Expired refresh token"})
				c.Abort()
			}
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
		appCtx.Logger.Debugfln("session: %#v", session)

		// Reuse detected!!!
		if session.RevokedAt.Valid {
			handleRefreshTokenReuse(appCtx, c, *user)
			return
		}

		refreshTokenHash := helper.HashWithSHA256(refreshTokenStr)
		// check if refresh token hash is correct
		if refreshTokenHash != session.RefreshTokenHash {
			ok := helper.HandleRevokeSession(appCtx, c, userId, sessId)
			if ok {
				appCtx.Logger.Debugfln("invalid refresh token hash (expected: %s, got: %s)", session.RefreshTokenHash, refreshTokenHash)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
				c.Abort()
			}
			return
		}

		// revoke session
		ok := helper.HandleRevokeSession(appCtx, c, userId, sessId)
		if !ok {
			return
		}

		// remove revoked session from user
		user.Sessions = helper.Filter(user.Sessions, func(s model.Session) bool {
			return s.Id.Int64() != sessId
		})

		// create new session ID
		newSessId, err := appCtx.IdGeneratorService.GenerateId()
		if err != nil {
			appCtx.Logger.Errorfln("GenerateId(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// create new tokens
		newAccessTokenStr, newRefreshTokenStr, err := helper.CreateNewTokenPair(appCtx.JwtConfig.JwtSecret, *user, appCtx.JwtConfig.ATDurationMs, appCtx.JwtConfig.RTDurationMs, newSessId)
		if err != nil {
			appCtx.Logger.Errorfln("CreateNewTokPair(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("AT: %s", newAccessTokenStr)
		appCtx.Logger.Debugfln("RT: %s", newRefreshTokenStr)

		// create new session
		expiresAt, err := helper.GetTokExpStr(newRefreshTokenStr, appCtx.JwtConfig.JwtSecret)
		if err != nil {
			appCtx.Logger.Errorfln("GetTokExpStr(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		newRefreshTokenHash := helper.HashWithSHA256(newRefreshTokenStr)
		appCtx.Logger.Debugfln("new refresh token hash: %s", newRefreshTokenHash)
		newSession := helper.HandleCreateSession(appCtx, c, helper.SessionPostPayload{
			Id:               newSessId,
			Version:          user.SessionVersion.Int64(),
			DeviceInfo:       session.DeviceInfo,
			RefreshTokenHash: newRefreshTokenHash,
			ExpiresAt:        expiresAt,
		}, user.Id.Int64())
		if newSession == nil {
			return
		}

		// add new session to user
		user.Sessions = append(user.Sessions, *newSession)

		// set cookie
		helper.SetCookieOverride(c, "refresh_token", newRefreshTokenStr, appCtx.DomainName, appCtx.JwtConfig.RTDurationMs)
		c.JSON(http.StatusOK, gin.H{
			"access_token": newAccessTokenStr,
			"user":         user,
			"session_id":   newSessId,
		})
	}
}

// handleRefreshTokenReuse is a function that handles the case when the refresh token
// is reused. It revokes all sessions of the user and increments the session version.
// It also clears the refresh token cookie and returns a 401 response.
func handleRefreshTokenReuse(appCtx *context.AppContext, c *gin.Context, user model.ChatUser) {
	// maybe stolen by someone. Regardless, REVOKE ALL!!!
	ok := helper.HandleRevokeAllSessions(appCtx, c, user.Id.Int64())
	if !ok {
		return
	}

	// update session version
	ok = helper.HandleIncrementSessionVersion(appCtx, c, user.Id.Int64())
	if !ok {
		return
	}

	appCtx.Logger.Debugfln("the account might be attacked")
	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token", "code": constants.REFRESH_TOKEN_REUSE})
	c.Abort()
	return
}

// handleRefreshTokenValidation is a function that handles the validation of the refresh token.
// It decodes the token twice: first without claims validation and then with claims validation.
// If the first decoding fails, it means that the token is invalid. If the second decoding
// fails, it means that the token is expired. In both cases, it clears the refresh token cookie
// and returns a 401 response. If the token is valid, it returns the parsed token.
// In case of an expired token, it also revokes the session. Note that we assume
// all token's fields are valid, and only expires field can be invalid.
func handleRefreshTokenValidation(appCtx *context.AppContext, c *gin.Context, tokenStr string) *helper.ParsedToken {
	// decode without claims validation first
	appCtx.Logger.Debugfln("decoding token 1st time: %s", tokenStr)
	token, err := jwthandler.DecodeTokenWithoutClaimsValidation(appCtx.JwtConfig.JwtSecret, tokenStr)
	if err != nil {
		appCtx.Logger.Debugfln("DecodeTokenWithoutClaimsValidation(): %v", err)
		helper.ClearCookie(c, "refresh_token", appCtx.DomainName)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		c.Abort()
		return nil
	}
	parsedToken, err := helper.ParseToken(token)
	if err != nil {
		appCtx.Logger.Errorfln("decodeTok(): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
		return nil
	}
	userId := parsedToken.UserId
	sessId := parsedToken.SessionId

	// decode with claims validation
	appCtx.Logger.Debugfln("decoding token 2nd time: %s", tokenStr)
	_, err = jwthandler.DecodeToken(appCtx.JwtConfig.JwtSecret, tokenStr)
	if err != nil {
		// probably an expired token, so revoke session
		appCtx.Logger.Debugfln("DecodeToken(): %v", err)
		ok := helper.HandleRevokeSession(appCtx, c, userId, sessId)
		if ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Expired refresh token"})
			c.Abort()
		}
		return nil
	}

	return parsedToken
}
