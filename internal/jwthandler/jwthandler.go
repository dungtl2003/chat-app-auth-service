package jwthandler

import (
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	ISSUER            = "auth.zolo.me"
	INTERNAL_AUDIENCE = "internal-service"
	FRONTEND_AUDIENCE = "zolo-frontend"
)

type InternalJWTClaim struct {
	jwt.RegisteredClaims
}

type UserJWTClaim struct {
	jwt.RegisteredClaims
	SessId      types.JsonInt64 `json:"sess_id"`
	SessVersion types.JsonInt64 `json:"sess_version"`
	Username    string          `json:"name"`
	Role        model.UserRole  `json:"role"`
}

func (c UserJWTClaim) GetSessVersion() (int64, error) {
	return c.SessVersion.Int64(), nil
}

func (c UserJWTClaim) GetSessId() (int64, error) {
	return c.SessId.Int64(), nil
}

func (c UserJWTClaim) GetUsername() (string, error) {
	return c.Username, nil
}

func (c UserJWTClaim) GetRole() (model.UserRole, error) {
	return c.Role, nil
}

// CreateUserToken will create jwt token with claims of user data, and the token is
// valid for `duration` milliseconds.
func CreateUserToken(key string, user model.ChatUser, duration int64, sessionId int64) (string, error) {
	iat := time.Now().UTC()
	exp := iat.Add(time.Duration(duration) * time.Millisecond)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		&UserJWTClaim{
			jwt.RegisteredClaims{
				IssuedAt:  &jwt.NumericDate{Time: iat},
				ExpiresAt: &jwt.NumericDate{Time: exp},
				Subject:   fmt.Sprint(user.Id.Int64()),
				Issuer:    ISSUER,
				Audience:  []string{FRONTEND_AUDIENCE},
			},
			types.NewJsonInt64(sessionId),
			user.SessionVersion,
			user.Username,
			user.Role,
		},
	)

	signedToken, err := token.SignedString([]byte(key))
	return signedToken, err
}

func CreateInternalToken(key string, duration int64, userId int64) (string, error) {
	iat := time.Now().UTC()
	exp := iat.Add(time.Duration(duration) * time.Millisecond)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		&InternalJWTClaim{
			jwt.RegisteredClaims{
				IssuedAt:  &jwt.NumericDate{Time: iat},
				ExpiresAt: &jwt.NumericDate{Time: exp},
				Issuer:    ISSUER,
				Audience:  []string{INTERNAL_AUDIENCE},
				Subject:   fmt.Sprintf("user:%d", userId),
			},
		},
	)

	signedToken, err := token.SignedString([]byte(key))
	return signedToken, err
}

func DecodeToken(key string, tokStr string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokStr, &UserJWTClaim{}, func(t *jwt.Token) (any, error) {
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

func DecodeTokenWithoutClaimsValidation(key string, tokStr string) (*jwt.Token, error) {
	parser := jwt.NewParser(
		jwt.WithoutClaimsValidation(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	token, err := parser.ParseWithClaims(tokStr, &UserJWTClaim{}, func(t *jwt.Token) (any, error) {
		return []byte(key), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}
