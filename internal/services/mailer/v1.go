package mailer

import (
	"dungtl2003/chat-app-auth-service/internal/logging"
	"dungtl2003/chat-app-auth-service/internal/services"
	"fmt"
	"net/smtp"
)

type SmtpMailer struct {
	status services.ServiceStatus
	logger *logging.LoggerWrapper

	Host string
	Port int
}

type SmtpMailerOptions struct {
	Logger *logging.LoggerWrapper
	Host   string
	Port   int
}

func NewSmtpMailerService(opts *SmtpMailerOptions) (*SmtpMailer, error) {
	s := &SmtpMailer{
		status: services.ServiceReady,
		logger: opts.Logger,
		Host:   opts.Host,
		Port:   opts.Port,
	}

	s.logger.Infofln("[%s] Service started", s.Name())
	return s, nil
}

func (s *SmtpMailer) Send(to string, subject string, body string) error {
	// Note: The "From" header usually gets overwritten by Gmail to match your account
	msg := []byte("Subject: " + subject + "\r\n" +
		"To: " + to + "\r\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		body + "\r\n")

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	// Send WITHOUT authentication (The container does the auth for you)
	// We pass 'nil' for the auth parameter
	err := smtp.SendMail(addr, nil, "Dunlyn Security <dunlyn.services@gmail.com>", []string{to}, msg)

	return err
}

func (s *SmtpMailer) Status() services.ServiceStatus {
	return s.status
}

func (s *SmtpMailer) Name() string {
	return "SMTP Service"
}

func (s *SmtpMailer) Close() error {
	if s.status == services.ServiceStopped {
		s.logger.Errorfln("[%s] Service is already stopped", s.Name())
		return nil
	}

	s.logger.Infofln("[%s] Closing service", s.Name())
	s.status = services.ServiceStopped
	s.logger.Infofln("[%s] Service stopped", s.Name())
	return nil
}
