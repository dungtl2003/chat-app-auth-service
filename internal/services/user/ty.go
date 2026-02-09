package user

import (
	"context"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/services"
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
)

const (
	UP   = "UP"
	DOWN = "DOWN"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func (r HealthResponse) String() string {
	return fmt.Sprintf("HealthResponse{Status: %s}", r.Status)
}

type UserGetResponse struct {
	User model.ChatUser `json:"user"`
}

func (r UserGetResponse) String() string {
	return fmt.Sprintf("UserGetResponse{User: %+v}", r.User)
}

type UserGetAuthResponse struct {
	User model.ChatUser `json:"user"`
}

func (r UserGetAuthResponse) String() string {
	return fmt.Sprintf("UserGetAuthResponse{User: %+v}", r.User)
}

type SessionPostRequest struct {
	Id               int64      `json:"id"`
	Version          int64      `json:"version"`
	DeviceInfo       types.Json `json:"device_info"`
	RefreshTokenHash string     `json:"refresh_token_hash"`
	ExpiresAt        string     `json:"expires_at"`
}

func (r SessionPostRequest) String() string {
	return fmt.Sprintf("SessionPostRequestBody{Id: %d, Version: %d, DeviceInfo: %s, RefreshTokenHash (hashed): %s, ExpiresAt: %s}", r.Id, r.Version, r.DeviceInfo, r.RefreshTokenHash, r.ExpiresAt)
}

type SessionPostResponse struct {
	Session model.Session `json:"session"`
}

func (r SessionPostResponse) String() string {
	return fmt.Sprintf("SessionPostResponse{Session: %+v}", r.Session)
}

type UserPostRequest struct {
	Email    string         `json:"email"`
	Username string         `json:"username"`
	Password string         `json:"password"`
	Role     model.UserRole `json:"role"`
}

func (r UserPostRequest) String() string {
	return fmt.Sprintf("UserPostRequest{Email: %s, Username: %s, Role: %s, Password (hashed): %s}", r.Email, r.Username, r.Role, r.Password)
}

type UserPostResponse struct {
	User model.ChatUser `json:"user"`
}

func (r UserPostResponse) String() string {
	return fmt.Sprintf("UserPostResponse{User: %+v}", r.User)
}

type GetUserByIdRequest struct {
	UserId    int64  `json:"user_id"`
	SessionId *int64 `json:"session_id"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	NewPassword string `json:"new_password"`
}

type UserService interface {
	services.Service
	CreateUser(context context.Context, payload UserPostRequest) (*UserPostResponse, error)
	GetUserAuth(context context.Context, identifier string) (*UserGetAuthResponse, error)
	GetUserById(context context.Context, request *GetUserByIdRequest) (*UserGetResponse, error)
	IncrementSessionVersion(context context.Context, userId int64) error
	RevokeAllSessions(context context.Context, userId int64) error
	RevokeSession(context context.Context, userId int64, sessionId int64) error
	CreateSession(context context.Context, userId int64, payload SessionPostRequest) (*SessionPostResponse, error)
	ResetPassword(context context.Context, request *ResetPasswordRequest) error
}
