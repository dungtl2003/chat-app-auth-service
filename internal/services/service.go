package services

import (
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
)

type ServiceStatus int

const (
	ServiceReady ServiceStatus = iota
	ServiceError
	ServiceStopped
)

type BadResponseError struct {
	ErrBlock types.ErrorBlock
}

func (e BadResponseError) Error() string {
	return fmt.Sprintf("bad response: %s (status code: %d)", e.ErrBlock.Message, e.ErrBlock.Code)
}

type Service interface {
	Status() ServiceStatus
	Name() string
	Close() error
}
