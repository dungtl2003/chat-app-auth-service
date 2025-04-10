package jwthandler

import (
	"dungtl2003/chat-app-auth-service/internal/model"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CreateToken will create jwt token with claims of user data, and the token is
// valid for `duration` milliseconds.
func CreateToken(key string, user model.ChatUser, duration int64) (string, error) {
	iat := time.Now().UnixMilli()
	exp := iat + duration
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": user.Username,
			"iss": "chat-app",
			"aud": user.Role,
			"iat": iat,
			"exp": exp,
		})

	signedToken, err := token.SignedString([]byte(key))
	return signedToken, err
}

func DecodeToken(key string, tokStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokStr, func(t *jwt.Token) (any, error) {
		return []byte(key), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}
