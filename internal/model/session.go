package model

import (
	"dungtl2003/chat-app-auth-service/internal/types"
)

type Session struct {
	Id               types.JsonInt64    `json:"id"`
	Version          types.JsonInt64    `json:"version"`
	DeviceInfo       types.Json         `json:"device_info"`
	RefreshTokenHash string             `json:"refresh_token_hash"`
	RevokedAt        types.JsonNullTime `json:"revoked_at"`
	IssuedAt         types.JsonTime     `json:"issued_at"`
	ExpiresAt        types.JsonTime     `json:"expires_at"`
	UserId           types.JsonInt64    `json:"user_id"`
	// Indicates whether the session was revoked by the user (owner) or by the
	// system (e.g., admin action, security reasons). When this is true, it means
	// the user themselves initiated the revocation of the session, not an external factor.
	RevokedByOwner bool `json:"revoked_by_owner"`
}
