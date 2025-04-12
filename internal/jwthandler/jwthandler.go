package jwthandler

import (
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaim struct {
	jwt.RegisteredClaims
	SessVersion types.JsonInt64 `json:"sess_version"`
}

func (c JWTClaim) GetSessVersion() (int64, error) {
	return c.SessVersion.Int64(), nil
}

// CreateToken will create jwt token with claims of user data, and the token is
// valid for `duration` milliseconds.
func CreateToken(key string, user model.ChatUser, duration int64) (string, error) {
	iat := time.Now().UTC()
	exp := iat.Add(time.Duration(duration) * time.Millisecond)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		&JWTClaim{
			jwt.RegisteredClaims{
				IssuedAt:  &jwt.NumericDate{Time: iat},
				ExpiresAt: &jwt.NumericDate{Time: exp},
				Subject:   user.Username,
				Issuer:    "chat-app",
				Audience:  []string{string(user.Role)},
			},
			user.SessionVersion,
		},
	)
	// token := jwt.NewWithClaims(jwt.SigningMethodHS256,
	// 	jwt.MapClaims{
	// 		"sub":          user.Username,
	// 		"iss":          "chat-app",
	// 		"aud":          user.Role,
	// 		"iat":          iat.Unix(),
	// 		"exp":          exp.Unix(),
	// 		"sess_version": user.SessionVersion.Int64(),

	signedToken, err := token.SignedString([]byte(key))
	return signedToken, err
}

func DecodeToken(key string, tokStr string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokStr, &JWTClaim{}, func(t *jwt.Token) (any, error) {
		return []byte(key), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	// token, err := jwt.Parse(tokStr, func(t *jwt.Token) (any, error) {
	// 	return []byte(key), nil
	// }, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}
