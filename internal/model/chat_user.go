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
	Id       types.JsonInt64 `json:"id"`
	Email    string          `json:"email"`
	Username string          `json:"username"`
	Password string          `json:"password"`
	Role     UserRole        `json:"role"`

	Devices []Device `json:"devices"`
}

func IsRole(role string) bool {
	return role == ADMIN || role == USER
}

func IsGender(gender string) bool {
	return gender == MALE || gender == FEMALE
}
