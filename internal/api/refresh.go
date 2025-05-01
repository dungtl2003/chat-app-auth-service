package api

import (
	"bytes"
	"dungtl2003/chat-app-auth-service/internal/constants"
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/httpclient"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Refresh logic flow:
// 1. If the refresh token is missing, return 401.
// 2. If the refresh token is invalid or expired, delete the cookie and return 401.
// 3. If the session version in the refresh token is less than the session version
// in the database, delete the cookie and return 401 (session version is the amount
// of times server detects that the user uses the same refresh token more than once.
// If the session version is less than the one in the database, it means that the
// user still uses the old valid refresh token).
// 4. If the token and the hash are not matched, delete the cookie and return 401.
// 5. If the corresponding session isn't found, delete the cookie and return 401.
// 6. If the corresponding session is found but already revoked, revoke all
// sessions belonged to that user, increase session version, delete the cookie,
// and return 401.
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

		appCtx.Logger.Debugfln("decoding token %s", refreshTokenStr)
		refreshToken, err := jwthandler.DecodeToken(appCtx.JwtConfig.JwtSecret, refreshTokenStr)
		if err != nil {
			appCtx.Logger.Debugfln("DecodeToken(): %v", err)
			c.SetCookie("refresh_token", "", -1, "/", appCtx.DomainName, false, true)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			c.Abort()
			return
		}

		// decode
		claims := refreshToken.Claims.(*jwthandler.JWTClaim)
		sessionVersion, err := claims.GetSessVersion()
		if err != nil {
			appCtx.Logger.Errorfln("GetSessVersion(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		userIdStr, err := claims.GetSubject()
		if err != nil {
			appCtx.Logger.Errorfln("GetSubject(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		userId, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			appCtx.Logger.Errorfln("strconv.ParseInt(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		sessId, err := claims.GetSessId()
		if err != nil {
			appCtx.Logger.Errorfln("GetSessId(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// 	get user information
		url := helper.EncodeURLPath(fmt.Sprintf("%s/users/%d", appCtx.UserServiceURL, userId))
		appCtx.Logger.Debugfln("sending GET request to %s", url)
		resp, err := appCtx.Client.Get(url, nil)
		if err != nil {
			appCtx.Logger.Errorfln("error when sending GET request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.Status(resp.StatusCode)
			_, err = io.Copy(c.Writer, resp.Body)
			if err != nil {
				appCtx.Logger.Errorfln("Copy(): %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
			}

			return
		}
		body, err := httpclient.ReadResponse(resp)
		if err != nil {
			appCtx.Logger.Errorfln("ReadResponse(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		var user model.ChatUser
		err = helper.ParseAsJson(body, &user)
		if err != nil {
			appCtx.Logger.Errorfln("ParseAsJson(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("response body from GET request: %#v", user)

		// the account was attacked so this refresh token is not valid anymore
		if sessionVersion < user.SessionVersion.Int64() {
			c.SetCookie("refresh_token", "", -1, "/", appCtx.DomainName, false, true)
			appCtx.Logger.Debugfln("invalid session version (expected: %d, got: %d)", user.SessionVersion.Int64(), sessionVersion)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session version", "code": constants.REFRESH_TOKEN_REUSE})
			c.Abort()
			return
		}

		// check if session is revoked
		url = helper.EncodeURLPath(fmt.Sprintf("%s/users/%d/sessions/%d", appCtx.UserServiceURL, user.Id.Int64(), sessId))
		appCtx.Logger.Debugfln("sending GET request to %s", url)
		resp, err = appCtx.Client.Get(url, nil)
		if err != nil {
			appCtx.Logger.Errorfln("error when sending GET request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.Status(resp.StatusCode)
			_, err = io.Copy(c.Writer, resp.Body)
			if err != nil {
				appCtx.Logger.Errorfln("Copy(): %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
			}
			return
		}
		body, err = httpclient.ReadResponse(resp)
		if err != nil {
			appCtx.Logger.Errorfln("ReadResponse(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		var session model.Session
		err = helper.ParseAsJson(body, &session)
		if err != nil {
			appCtx.Logger.Errorfln("ParseAsJson(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("response body from GET request: %#v", session)

		refreshTokenHash := helper.HashWithSHA256(refreshTokenStr)
		// check if refresh token hash is correct
		if refreshTokenHash != session.RefreshTokenHash {
			appCtx.Logger.Debugfln("invalid refresh token hash (expected: %s, got: %s)", session.RefreshTokenHash, refreshTokenHash)
			c.SetCookie("refresh_token", "", -1, "/", appCtx.DomainName, false, true)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			c.Abort()
			return
		}

		// Reuse detected!!!
		if session.RevokedAt.Valid {
			// maybe stolen by someone. Regardless, REVOKE ALL!!!
			url = helper.EncodeURLPath(fmt.Sprintf("%s/users/%d/sessions/revoke", appCtx.UserServiceURL, user.Id.Int64()))
			appCtx.Logger.Debugfln("sending POST request to %s", url)
			resp, err = appCtx.Client.Post(url, nil, nil)
			if err != nil {
				appCtx.Logger.Errorfln("error when sending POST request: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
				return
			}
			if resp.StatusCode != http.StatusOK {
				c.Status(resp.StatusCode)
				_, err = io.Copy(c.Writer, resp.Body)
				if err != nil {
					appCtx.Logger.Errorfln("Copy(): %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				}

				c.Abort()
				return
			}

			// update session version
			url = helper.EncodeURLPath(fmt.Sprintf("%s/users/%d/session-version/increment", appCtx.UserServiceURL, user.Id.Int64()))
			appCtx.Logger.Debugfln("sending POST request to %s", url)
			resp, err = appCtx.Client.Post(url, nil, nil)
			if err != nil {
				appCtx.Logger.Errorfln("error when sending PATCH request: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
				return
			}
			if resp.StatusCode != http.StatusOK {
				c.Status(resp.StatusCode)
				_, err = io.Copy(c.Writer, resp.Body)
				if err != nil {
					appCtx.Logger.Errorfln("Copy(): %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				}

				c.Abort()
				return
			}

			appCtx.Logger.Debugfln("the account might be attacked")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token", "code": constants.REFRESH_TOKEN_REUSE})
			c.SetCookie("refresh_token", "", -1, "/", appCtx.DomainName, false, true)
			c.Abort()
			return
		}

		// revoke session
		url = helper.EncodeURLPath(fmt.Sprintf("%s/users/%d/sessions/%d/revoke", appCtx.UserServiceURL, user.Id.Int64(), sessId))
		appCtx.Logger.Debugfln("sending POST request to %s", url)
		resp, err = appCtx.Client.Post(url, nil, nil)
		if err != nil {
			appCtx.Logger.Errorfln("error when sending POST request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.Status(resp.StatusCode)
			_, err = io.Copy(c.Writer, resp.Body)
			if err != nil {
				appCtx.Logger.Errorfln("Copy(): %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
			c.Abort()
			return
		}

		// create new session ID
		newSessId, err := appCtx.IdGeneratorService.GenerateId()
		if err != nil {
			appCtx.Logger.Errorfln("GenerateId(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// create new tokens
		newAccessTokenStr, err := jwthandler.CreateToken(appCtx.JwtConfig.JwtSecret, user, appCtx.JwtConfig.ATDurationMs, newSessId)
		if err != nil {
			appCtx.Logger.Errorfln("CreateToken(): error creating access token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("AT: %s", newAccessTokenStr)

		newRefreshTokenStr, err := jwthandler.CreateToken(appCtx.JwtConfig.JwtSecret, user, appCtx.JwtConfig.RTDurationMs, newSessId)
		if err != nil {
			appCtx.Logger.Errorfln("CreateToken(): error creating refresh token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("RT: %s", newRefreshTokenStr)
		newRefreshToken, err := jwthandler.DecodeToken(appCtx.JwtConfig.JwtSecret, newRefreshTokenStr)
		if err != nil {
			appCtx.Logger.Errorfln("DecodeToken(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.SetCookie("refresh_token", "", -1, "/", appCtx.DomainName, false, true)
			c.Abort()
			return
		}
		claims = newRefreshToken.Claims.(*jwthandler.JWTClaim)
		expiresAt, err := claims.GetExpirationTime()
		if err != nil {
			appCtx.Logger.Errorfln("GetExpiresAt(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.SetCookie("refresh_token", "", -1, "/", appCtx.DomainName, false, true)
			c.Abort()
			return
		}

		// create new session
		newRefreshTokenHash := helper.HashWithSHA256(newRefreshTokenStr)
		url = helper.EncodeURLPath(fmt.Sprintf("%s/users/%d/sessions", appCtx.UserServiceURL, user.Id.Int64()))
		payload := fmt.Appendf(nil, `
		{
			"id": "%d",
			"version": "%d",
			"device_info": %s,
			"refresh_token_hash": "%s",
			"expires_at": "%s"
		}`, newSessId, user.SessionVersion.Int64(), session.DeviceInfo, newRefreshTokenHash, expiresAt.Format("2006-01-02T15:04:05.999Z"))
		resp, err = appCtx.Client.Post(url, nil, bytes.NewBuffer(payload))
		if err != nil {
			appCtx.Logger.Errorfln("error when sending POST request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if resp.StatusCode != http.StatusCreated {
			c.Status(resp.StatusCode)
			_, err = io.Copy(c.Writer, resp.Body)
			if err != nil {
				appCtx.Logger.Errorfln("Copy(): %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
			}
			return
		}
		body, err = httpclient.ReadResponse(resp)
		if err != nil {
			appCtx.Logger.Errorfln("ReadResponse(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		var newSession model.Session
		err = helper.ParseAsJson(body, &newSession)
		if err != nil {
			appCtx.Logger.Errorfln("ParseAsJson(): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		appCtx.Logger.Debugfln("response body from POST request: %#v", newSession)

		// add new session to user
		user.Sessions = append(user.Sessions, newSession)

		// security
		user.Password = ""

		// set cookie
		c.SetCookie("refresh_token", newRefreshTokenStr, int(appCtx.JwtConfig.RTDurationMs/1000), "/", appCtx.DomainName, false, true)
		c.JSON(http.StatusOK, gin.H{
			"access_token": newAccessTokenStr,
			"user":         user,
			"session_id":   newSessId,
		})
	}
}
