package mailer

import (
	"dungtl2003/chat-app-auth-service/internal/services"
	"dungtl2003/chat-app-auth-service/internal/types"
	"fmt"
)

type BadResponseError struct {
	ErrBlock types.ErrorBlock
}

func (e BadResponseError) Error() string {
	return fmt.Sprintf("bad response: %s (status code: %d)", e.ErrBlock.Message, e.ErrBlock.Code)
}

type HealthResponse struct {
	Status string `json:"status"`
}

type SendEmailRequest struct {
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`
	To          string `json:"to"`
	Subject     string `json:"subject"`
	Body        string `json:"body"`
}

type MailerService interface {
	services.Service
	Send(req *SendEmailRequest) error
}
