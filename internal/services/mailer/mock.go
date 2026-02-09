package mailer

import "dungtl2003/chat-app-auth-service/internal/services"

type MockMailer struct {
	MockNameFunc   func() string
	MockStatusFunc func() services.ServiceStatus
	MockCloseFunc  func() error
	MockSend       func(to string, subject string, body string) error
}

func (m *MockMailer) Send(to string, subject string, body string) error {
	if m.MockSend != nil {
		return m.MockSend(to, subject, body)
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
