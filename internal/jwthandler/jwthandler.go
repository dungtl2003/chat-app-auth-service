package jwthandler

import (
	"dungtl2003/chat-app-auth-service/internal/model"
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
			"username": user.Username,
			"role":     user.Role,
			"iat":      iat,
			"exp":      exp,
		})

	signedToken, err := token.SignedString([]byte(key))
	return signedToken, err
}
