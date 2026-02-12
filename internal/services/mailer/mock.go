package mailer

import "dungtl2003/chat-app-auth-service/internal/services"

type MockMailer struct {
	MockNameFunc   func() string
	MockStatusFunc func() services.ServiceStatus
	MockCloseFunc  func() error
	MockSend       func(req *SendEmailRequest) error
}

func (m *MockMailer) Send(req *SendEmailRequest) error {
	if m.MockSend != nil {
		return m.MockSend(req)
	}

	return nil
}

func (m *MockMailer) Name() string {
	if m.MockNameFunc != nil {
		return m.MockNameFunc()
	}

	return "Mock Mailer Service"
}

func (m *MockMailer) Status() services.ServiceStatus {
	if m.MockStatusFunc != nil {
		return m.MockStatusFunc()
	}

	// Default status is READY
	return services.ServiceReady
}

func (m *MockMailer) Close() error {
	if m.MockCloseFunc != nil {
		return m.MockCloseFunc()
	}

	return nil
}
