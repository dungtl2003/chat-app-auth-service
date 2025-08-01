package helper

import (
	"crypto/sha256"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"dungtl2003/chat-app-auth-service/internal/model"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

// CreateNewTokenPair creates a new access token and refresh token pair for the given user.
func CreateNewTokenPair(secret string, user model.ChatUser, atDurationMs int64, rtDurationMs int64, sessionId int64) (string, string, error) {
	accessTokenStr, err := jwthandler.CreateUserToken(secret, user, atDurationMs, sessionId)
	if err != nil {
		return "", "", fmt.Errorf("error creating access token: CreateToken(): %v", err)
	}
	refreshTokenStr, err := jwthandler.CreateUserToken(secret, user, rtDurationMs, sessionId)
	if err != nil {
		return "", "", fmt.Errorf("error creating refresh token: CreateToken(): %v", err)
	}

	return accessTokenStr, refreshTokenStr, nil
}

// GetTokExpStr returns the expiration time of the token as a string in the format "2006-01-02T15:04:05.999Z".
func GetTokExpStr(tokenStr string, secret string) (string, error) {
	token, err := jwthandler.DecodeToken(secret, tokenStr)
	if err != nil {
		return "", fmt.Errorf("DecodeToken(): %v", err)
	}
	claims := token.Claims.(*jwthandler.UserJWTClaim)
	expiresAt, err := claims.GetExpirationTime()
	if err != nil {
		return "", fmt.Errorf("GetExpiresAt(): %v", err)
	}

	return expiresAt.Format("2006-01-02T15:04:05.999Z"), nil
}

func ParseAsJson(data []byte, target any) error {
	return json.Unmarshal(data, target)
}

func StripWS(s string) string {
	s = strings.TrimSpace(s)

	var b strings.Builder
	b.Grow(len(s))

	isSpace := false
	for _, ch := range s {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
			isSpace = false
		} else if !isSpace {
			b.WriteRune(' ')
			isSpace = true
		}
	}

	return b.String()
}

func EncodeURLPath(path string) string {
	// Encode the path to make it safe for use in a URL
	return strings.ReplaceAll(path, " ", "%20")
}

func HashWithSHA256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func Filter[T any](arr []T, f func(T) bool) []T {
	var res []T
	for _, v := range arr {
		if f(v) {
			res = append(res, v)
		}
	}
	return res
}
