package user

import (
	"context"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/services"
	"dungtl2003/chat-app-auth-service/internal/types"
)

const (
	UP   = "UP"
	DOWN = "DOWN"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type UserGetResponseBody struct {
	User  model.ChatUser `json:"user"`
	Error string         `json:"error"`
}

type UserGetResponse struct {
	StatusCode int                 `json:"status_code"`
	Body       UserGetResponseBody `json:"body"`
}

type UserGetAuthResponseBody struct {
	User  model.ChatUser `json:"user"`
	Error string         `json:"error"`
}

type UserGetAuthResponse struct {
	StatusCode int                     `json:"status_code"`
	Body       UserGetAuthResponseBody `json:"body"`
}

type SessionPostRequestBody struct {
	Id               int64      `json:"id"`
	Version          int64      `json:"version"`
	DeviceInfo       types.Json `json:"device_info"`
	RefreshTokenHash string     `json:"refresh_token_hash"`
	ExpiresAt        string     `json:"expires_at"`
}

type SessionPostResponseBody struct {
	Session model.Session `json:"session"`
	Error   string        `json:"error"`
}

type SessionPostResponse struct {
	StatusCode int                     `json:"status_code"`
	Body       SessionPostResponseBody `json:"body"`
}

type UserPostRequestBody struct {
	Email    string         `json:"email"`
	Username string         `json:"username"`
	Password string         `json:"password"`
	Role     model.UserRole `json:"role"`
}

type UserPostResponseBody struct {
	User  model.ChatUser `json:"user"`
	Error string         `json:"error"`
}

type UserPostResponse struct {
	StatusCode int                  `json:"status_code"`
	Body       UserPostResponseBody `json:"body"`
}

type UserService interface {
	services.Service
	CreateUser(context context.Context, payload UserPostRequestBody) (*UserPostResponse, error)
	GetUserAuth(context context.Context, identifier string) (*UserGetAuthResponse, error)
	GetUserById(context context.Context, userId int64) (*UserGetResponse, error)
	IncrementSessionVersion(context context.Context, userId int64) error
	RevokeAllSessions(context context.Context, userId int64) error
	RevokeSession(context context.Context, userId int64, sessionId int64) error
	CreateSession(context context.Context, userId int64, payload SessionPostRequestBody) (*SessionPostResponse, error)
}
