package model

import (
	"dungtl2003/chat-app-auth-service/internal/types"
)

type DeviceStatus string

const (
	ONLINE    = "ONLINE"
	OFFLINE   = "OFFLINE"
	AWAY      = "AWAY"
	BUSY      = "BUSY"
	INVISIBLE = "INVISIBLE"
)

type Device struct {
	Id         types.JsonInt64 `json:"id"`
	DeviceName string          `json:"device_name"`
	DeviceType string          `json:"device_type"`
	Os         string          `json:"os"`
	Status     DeviceStatus    `json:"status"`
}

func IsDeviceStatus(status string) bool {
	return status == ONLINE || status == OFFLINE || status == AWAY || status == BUSY || status == INVISIBLE
}

type ById []Device

func (a ById) Len() int           { return len(a) }
func (a ById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ById) Less(i, j int) bool { return a[i].Id.Int64() < a[j].Id.Int64() }
