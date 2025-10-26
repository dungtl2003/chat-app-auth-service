package api

import (
	"dungtl2003/chat-app-auth-service/internal/constants"
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/services"
	u "dungtl2003/chat-app-auth-service/internal/services/user"
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type RefreshTokenResponseBody struct {
	AccessToken string `json:"access_token"`
	// This is for native clients that cannot store HttpOnly cookies
	RefreshToken string          `json:"refresh_token,omitempty"`
	SessionId    types.JsonInt64 `json:"session_id"`
	User         model.ChatUser  `json:"user"`
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
		response := types.Response[RefreshTokenResponseBody]{}
		native := helper.IsNativeClient(c)

		refreshTokenStr, err := extractRTAbort(c, appCtx, native)
		if err != nil {
			return
		}

		parsedToken, err := validateRTAbort(c, appCtx, refreshTokenStr, native)
		if err != nil {
			return
		}

		userId := parsedToken.UserId
		sessionVersion := parsedToken.SessionVersion
		sessId := parsedToken.SessionId

		user, err := getUserAbort(appCtx, c, userId)
		if err != nil {
			return
		}
		// the account was attacked so this refresh token is not valid anymore
		if sessionVersion < user.SessionVersion.Int64() {
			handleLowerSessionVersionWithAbort(appCtx, c, userId, sessionVersion, sessId, native)
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
			handleRTReuseWithAbort(appCtx, c, *user)
			return
		}

		refreshTokenHash := helper.HashWithSHA256(refreshTokenStr)
		// check if refresh token hash is correct
		if refreshTokenHash != session.RefreshTokenHash {
			handleDifferentRTHashWithAbort(appCtx, c, userId, sessId, session.RefreshTokenHash, refreshTokenHash, native)
			return
		}

		// revoke session
		err = revokeSessionAbort(appCtx, c, userId, sessId)
		if err != nil {
			return
		}

		// remove revoked session from user
		user.Sessions = helper.Filter(user.Sessions, func(s model.Session) bool {
			return s.Id.Int64() != sessId
		})

		newSessId, err := generateIdAbort(appCtx, c)
		if err != nil {
			return
		}

		newAccessTokenStr, newRefreshTokenStr, err := createNewTokenPairAbort(appCtx, c, *user, newSessId)

		newSession, err := createNewSessionAbort(appCtx, c, *user, newSessId, newRefreshTokenStr, session.DeviceInfo)
		if err != nil {
			return
		}
		user.Sessions = append(user.Sessions, *newSession)

		if !native {
			// set cookie
			refreshResponseBody := RefreshTokenResponseBody{
				AccessToken: newAccessTokenStr,
				SessionId:   types.NewJsonInt64(newSessId),
				User:        *user,
			}
			response.Data = &types.DataOrPage[RefreshTokenResponseBody]{
				Item: &refreshResponseBody,
			}
			appCtx.Logger.Debugfln("New session created: %d", newSessId)
			helper.SetCookieOverride(c, "refresh_token", newRefreshTokenStr, appCtx.DomainName, appCtx.JwtConfig.RTDurationMs)
			c.JSON(http.StatusOK, response)
			c.Abort()
			return
		}

		// for native clients, return refresh token in response body
		refreshResponseBody := RefreshTokenResponseBody{
			AccessToken:  newAccessTokenStr,
			RefreshToken: newRefreshTokenStr,
			SessionId:    types.NewJsonInt64(newSessId),
			User:         *user,
		}
		response.Data = &types.DataOrPage[RefreshTokenResponseBody]{
			Item: &refreshResponseBody,
		}
		appCtx.Logger.Debugfln("New session created: %d", newSessId)
		c.JSON(http.StatusOK, response)
		c.Abort()
	}
}

// getUserAbort is a helper function that gets user information.
// If there is an error, it returns an error response and aborts the request.
func getUserAbort(
	appCtx *context.AppContext,
	c *gin.Context,
	userId int64,
) (*model.ChatUser, error) {
	response := types.Response[RefreshTokenResponseBody]{}

	userResponse, err := appCtx.UserService.GetUserById(c, userId)
	if err != nil {
		appCtx.Logger.Errorfln("GetUserById(): %v", err)
		switch err := err.(type) {
		case services.BadResponseError:
			response.Error = &err.ErrBlock
		default:
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
		}

		c.JSON(response.Error.Code, response)
		c.Abort()
		return nil, err

	}

	// the account was attacked so this refresh token is not valid anymore
	user := userResponse.User
	return &user, nil
}

// revokeSessionAbort is a helper function that revokes a session.
// If there is an error, it returns an error response and aborts the request.
func revokeSessionAbort(
	appCtx *context.AppContext,
	c *gin.Context,
	userId int64,
	sessId int64,
) error {
	response := types.Response[RefreshTokenResponseBody]{}

	err := appCtx.UserService.RevokeSession(c, userId, sessId)
	if err != nil {
		appCtx.Logger.Errorfln("RevokeSession(): %v", err)
		response.Error = &types.ErrorBlock{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Errors:  []types.ErrorItem{{Message: "Internal server error"}},
		}
		c.JSON(response.Error.Code, response)
		c.Abort()
		return err
	}
	return nil
}

// generateIdAbort is a helper function that generates a new ID.
// If there is an error, it returns an error response and aborts the request.
func generateIdAbort(appCtx *context.AppContext, c *gin.Context) (int64, error) {
	response := types.Response[RefreshTokenResponseBody]{}

	id, err := appCtx.IdGeneratorService.GenerateId(c)
	if err != nil {
		appCtx.Logger.Errorfln("GenerateId(): %v", err)
		response.Error = &types.ErrorBlock{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Errors:  []types.ErrorItem{{Message: "Internal server error"}},
		}
		c.JSON(response.Error.Code, response)
		c.Abort()
		return 0, err
	}

	return id, nil
}

// createNewSessionAbort is a helper function that creates a new session.
// If there is an error, it returns an error response and aborts the request.
func createNewSessionAbort(
	appCtx *context.AppContext,
	c *gin.Context,
	user model.ChatUser,
	newSessId int64,
	newRefreshTokenStr string,
	deviceInfo types.Json,
) (*model.Session, error) {
	response := types.Response[RefreshTokenResponseBody]{}

	expiresAt, err := helper.GetTokenExpirationStr(newRefreshTokenStr, appCtx.JwtConfig.JwtSecret)
	if err != nil {
		appCtx.Logger.Errorfln("GetTokenExpirationStr(): %v", err)
		response.Error = &types.ErrorBlock{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Errors:  []types.ErrorItem{{Message: "Internal server error"}},
		}
		c.JSON(response.Error.Code, response)
		c.Abort()
		return nil, err
	}

	newRefreshTokenHash := helper.HashWithSHA256(newRefreshTokenStr)
	appCtx.Logger.Debugfln("New refresh token hash: %s", newRefreshTokenHash)
	newSessionResponse, err := appCtx.UserService.CreateSession(c, user.Id.Int64(), u.SessionPostRequest{
		Id:               newSessId,
		Version:          user.SessionVersion.Int64(),
		DeviceInfo:       deviceInfo,
		RefreshTokenHash: newRefreshTokenHash,
		ExpiresAt:        expiresAt,
	})
	if err != nil {
		appCtx.Logger.Errorfln("CreateSession(): %v", err)
		switch err := err.(type) {
		case services.BadResponseError:
			response.Error = &err.ErrBlock
		default:
			response.Error = &types.ErrorBlock{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors:  []types.ErrorItem{{Message: "Internal server error"}},
			}
		}

		c.JSON(response.Error.Code, response)
		c.Abort()
		return nil, err
	}

	appCtx.Logger.Debugfln("New session created: %d", newSessId)
	return &newSessionResponse.Session, nil
}

// createNewTokenPairAbort is a helper function that creates a new token pair.
// If there is an error, it returns an error response and aborts the request.
func createNewTokenPairAbort(
	appCtx *context.AppContext,
	c *gin.Context,
	user model.ChatUser,
	newSessId int64,
) (string, string, error) {
	response := types.Response[RefreshTokenResponseBody]{}

	newAccessTokenStr, newRefreshTokenStr, err := helper.CreateNewTokenPair(appCtx.JwtConfig.JwtSecret, user, appCtx.JwtConfig.ATDurationMs, appCtx.JwtConfig.RTDurationMs, newSessId)
	if err != nil {
		appCtx.Logger.Errorfln("CreateNewTokenPair(): %v", err)
		response.Error = &types.ErrorBlock{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Errors:  []types.ErrorItem{{Message: "Internal server error"}},
		}
		c.JSON(response.Error.Code, response)
		c.Abort()
		return "", "", err
	}

	appCtx.Logger.Debugfln("New access token: %s", newAccessTokenStr)
	appCtx.Logger.Debugfln("New refresh token: %s", newRefreshTokenStr)
	return newAccessTokenStr, newRefreshTokenStr, nil
}

func handleDifferentRTHashWithAbort(
	appCtx *context.AppContext,
	c *gin.Context,
	userId int64,
	sessId int64,
	expectedHash string, refreshTokenHash string,
	native bool,
) {
	response := types.Response[RefreshTokenResponseBody]{}

	err := appCtx.UserService.RevokeSession(c, userId, sessId)
	if err != nil {
		appCtx.Logger.Errorfln("RevokeSession(): %v", err)
		response.Error = &types.ErrorBlock{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Errors:  []types.ErrorItem{{Message: "Internal server error"}},
		}
		c.JSON(response.Error.Code, response)
		c.Abort()
		return
	}

	appCtx.Logger.Errorfln("invalid refresh token hash (expected: %s, got: %s)", expectedHash, refreshTokenHash)
	if !native {
		helper.ClearCookie(c, "refresh_token", appCtx.DomainName)
	}
	response.Error = &types.ErrorBlock{
		Code:    http.StatusUnauthorized,
		Message: "Invalid refresh token",
		Errors:  []types.ErrorItem{{Message: "Invalid refresh token"}},
	}
	c.JSON(response.Error.Code, response)
	c.Abort()
}

// handleLowerSessionVersionWithAbort is a helper function that handles the case
// when the session version in the refresh token is lower than the one in the
// database. It revokes the session, clears the cookie (if not native client),
// and returns a 401 response.
// This indicates that the refresh token is no longer valid (the user has already
// refreshed the token before with a higher session version).
func handleLowerSessionVersionWithAbort(
	appCtx *context.AppContext,
	c *gin.Context,
	userId int64,
	sessionVersion int64,
	sessId int64,
	native bool,
) {
	response := types.Response[RefreshTokenResponseBody]{}

	appCtx.Logger.Errorfln("Invalid session version (expected: %d, got: %d)", userId, sessionVersion)
	err := appCtx.UserService.RevokeSession(c, userId, sessId)
	if err != nil {
		appCtx.Logger.Errorfln("RevokeSession(): %v", err)
		response.Error = &types.ErrorBlock{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Errors:  []types.ErrorItem{{Message: "Internal server error"}},
		}
		c.JSON(response.Error.Code, response)
		c.Abort()
		return
	}

	if !native {
		helper.ClearCookie(c, "refresh_token", appCtx.DomainName)
	}

	response.Error = &types.ErrorBlock{
		Code:    http.StatusUnauthorized,
		Message: "Invalid session version",
		Status:  constants.REFRESH_TOKEN_REUSE,
		Errors:  []types.ErrorItem{{Message: "Invalid session version", Reason: constants.REFRESH_TOKEN_REUSE}},
	}
	c.JSON(response.Error.Code, response)
	c.Abort()
}

// extractRTAbort is a helper function that extracts the refresh token from
// either the Authorization header (for native clients) or the refresh_token
// cookie (for web clients). If the refresh token is missing, it returns an
// error response and aborts the request.
func extractRTAbort(c *gin.Context, appCtx *context.AppContext, native bool) (string, error) {
	response := types.Response[RefreshTokenResponseBody]{}
	var refreshTokenStr string
	var err error

	if native {
		authHeader := c.GetHeader("Authorization")
		appCtx.Logger.Debugfln("Authorization header: %s", authHeader)
		if authHeader == "" {
			appCtx.Logger.Errorfln("Missing authorization header")
			response.Error = &types.ErrorBlock{
				Code:    http.StatusUnauthorized,
				Message: "Missing authorization header",
				Errors:  []types.ErrorItem{{Message: "Missing authorization header"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return "", fmt.Errorf("missing authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			appCtx.Logger.Errorfln("Authorization header should have 2 parts")
			response.Error = &types.ErrorBlock{
				Code:    http.StatusUnauthorized,
				Message: "Authorization header should have 2 parts",
				Errors:  []types.ErrorItem{{Message: "Authorization header should have 2 parts"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return "", fmt.Errorf("authorization header should have 2 parts")
		}

		if parts[0] != "Bearer" {
			appCtx.Logger.Errorfln("Authorization header should start with `Bearer`")
			response.Error = &types.ErrorBlock{
				Code:    http.StatusUnauthorized,
				Message: "Authorization header should start with `Bearer`",
				Errors:  []types.ErrorItem{{Message: "Authorization header should start with `Bearer`"}},
			}
			c.JSON(response.Error.Code, response)
			c.Abort()
			return "", fmt.Errorf("authorization header should start with `Bearer`")
		}

		refreshTokenStr = parts[1]
	} else {
		refreshTokenStr, err = c.Cookie("refresh_token")
		if err != nil {
			if err == http.ErrNoCookie {
				appCtx.Logger.Errorfln("Missing refresh token")
				response.Error = &types.ErrorBlock{
					Code:    http.StatusUnauthorized,
					Message: "Missing refresh token",
					Status:  constants.REFRESH_TOKEN_NOT_FOUND,
					Errors:  []types.ErrorItem{{Message: "Missing refresh token", Reason: constants.REFRESH_TOKEN_NOT_FOUND}},
				}
				c.JSON(response.Error.Code, response)
			} else {
				appCtx.Logger.Errorfln("Cookie(): %v", err)
				response.Error = &types.ErrorBlock{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Errors:  []types.ErrorItem{{Message: "Internal server error"}},
				}
				c.JSON(response.Error.Code, response)
			}
			c.Abort()
			return "", fmt.Errorf("cannot get refresh token cookie: %w", err)
		}
	}

	appCtx.Logger.Debugfln("Refresh token: %s", refreshTokenStr)
	return refreshTokenStr, nil
}

// validateRTAbort is a helper function that validates the refresh token. If the
// token is invalid or expired, it will clear the cookie (if not native client)
// and return an error response and abort the request.
func validateRTAbort(
	c *gin.Context,
	appCtx *context.AppContext,
	refreshTokenStr string,
	native bool,
) (*helper.ParsedToken, error) {
	response := types.Response[RefreshTokenResponseBody]{}

	// Validate and parse the refresh token. If the token is invalid or
	// expired, it will clear the cookie. For expired token, it will also
	// revoke the session. In any case, it will return an error response.
	parsedToken, err := validateToken(refreshTokenStr, appCtx.JwtConfig.JwtSecret)
	if err == nil {
		return parsedToken, nil
	}

	switch e := err.(type) {
	case InvalidTokenError:
		appCtx.Logger.Errorfln("Invalid refresh token")
		if !native {
			helper.ClearCookie(c, "refresh_token", appCtx.DomainName)
		}
		response.Error = &types.ErrorBlock{
			Code:    http.StatusUnauthorized,
			Message: "Invalid refresh token",
			Errors:  []types.ErrorItem{{Message: "Invalid refresh token"}},
		}
	case ExpiredTokenError:
		appCtx.Logger.Errorfln("Expired refresh token")
		// TODO: make this kafka event
		err = appCtx.UserService.RevokeSession(c, e.UserId, e.SessionId)
		if err != nil {
			appCtx.Logger.Warnfln("RevokeSession(): %v", e)
		}
		if !native {
			helper.ClearCookie(c, "refresh_token", appCtx.DomainName)
		}
		response.Error = &types.ErrorBlock{
			Code:    http.StatusUnauthorized,
			Message: "Expired refresh token",
			Errors:  []types.ErrorItem{{Message: "Expired refresh token"}},
		}
	default:
		appCtx.Logger.Errorfln("ParseToken(): %v", e)
		response.Error = &types.ErrorBlock{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Errors:  []types.ErrorItem{{Message: "Internal server error"}},
		}
	}

	c.JSON(response.Error.Code, response)
	c.Abort()
	return nil, err
}

// handleRTReuseWithAbort is a function that handles the case when the refresh token
// is reused. It revokes all sessions of the user and increments the session version.
// It also clears the refresh token cookie and returns a 401 response.
func handleRTReuseWithAbort(appCtx *context.AppContext, c *gin.Context, user model.ChatUser) {
	response := types.Response[RefreshTokenResponseBody]{}
	// maybe stolen by someone. Regardless, REVOKE ALL!!!
	err := appCtx.UserService.RevokeAllSessions(c, user.Id.Int64())
	if err != nil {
		appCtx.Logger.Errorfln("RevokeAllSessions(): %v", err)
		response.Error = &types.ErrorBlock{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Errors:  []types.ErrorItem{{Message: "Internal server error"}},
		}
		c.JSON(response.Error.Code, response)
		c.Abort()
		return
	}

	// update session version
	err = appCtx.UserService.IncrementSessionVersion(c, user.Id.Int64())
	if err != nil {
		appCtx.Logger.Errorfln("IncrementSessionVersion(): %v", err)
		response.Error = &types.ErrorBlock{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Errors:  []types.ErrorItem{{Message: "Internal server error"}},
		}
		c.JSON(response.Error.Code, response)
		c.Abort()
		return
	}

	appCtx.Logger.Errorfln("The account might be attacked")
	helper.ClearCookie(c, "refresh_token", appCtx.DomainName)
	response.Error = &types.ErrorBlock{
		Code:    http.StatusUnauthorized,
		Message: "Refresh token reuse detected",
		Status:  constants.REFRESH_TOKEN_REUSE,
		Errors:  []types.ErrorItem{{Message: "Refresh token reuse detected", Reason: constants.REFRESH_TOKEN_REUSE}},
	}
	c.JSON(response.Error.Code, response)
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
