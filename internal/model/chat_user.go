package model

import "dungtl2003/chat-app-auth-service/internal/types"

type Gender string
type UserRole string

const (
	MALE   = "MALE"
	FEMALE = "FEMALE"

	ADMIN = "ADMIN"
	USER  = "USER"
)

type ChatUser struct {
	Id             types.JsonInt64      `json:"id"`
	Email          string               `json:"email"`
	Username       string               `json:"username"`
	Password       string               `json:"password"`
	Role           UserRole             `json:"role"`
	FirstName      types.JsonNullString `json:"first_name"`
	LastName       types.JsonNullString `json:"last_name"`
	Birthday       types.JsonNullTime   `json:"birthday"`
	Gender         types.JsonNullString `json:"gender"`
	PhoneNumber    types.JsonNullString `json:"phone_number"`
	Privacy        types.JsonNullString `json:"privacy"`
	Avatar         types.JsonNullString `json:"avatar"`
	SessionVersion types.JsonInt64      `json:"session_version"`
	CreatedAt      types.JsonTime       `json:"created_at"`
	UpdatedAt      types.JsonNullTime   `json:"updated_at"`
	DeletedAt      types.JsonNullTime   `json:"deleted_at"`

	Sessions []Session `json:"sessions"`

	TotalFriends               int64 `json:"total_friends"`
	TotalPendingFriendRequests int64 `json:"total_pending_friend_requests"`
	TotalSentFriendRequests    int64 `json:"total_sent_friend_requests"`
	TotalBlockedUsers          int64 `json:"total_blocked_users"`
}

func IsRole(role string) bool {
	return role == ADMIN || role == USER
}

func IsGender(gender string) bool {
	return gender == MALE || gender == FEMALE
}
