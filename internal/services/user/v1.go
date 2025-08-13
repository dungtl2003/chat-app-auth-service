package user

import (
	"bytes"
	"context"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/logging"
	"dungtl2003/chat-app-auth-service/internal/services"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type UserServiceV1 struct {
	status  services.ServiceStatus
	logger  *logging.LoggerWrapper
	userURL string
	client  *http.Client
}

type UserServiceV1Options struct {
	Logger *logging.LoggerWrapper
	Client *http.Client
}

// New creates a new UserServiceV1 instance. Remember to call Close() when done
// to release resources.
func NewUSerServiceV1(userURL string, opts *UserServiceV1Options) (*UserServiceV1, error) {
	var loggerWrapper *logging.LoggerWrapper
	var client *http.Client

	if opts != nil && opts.Logger == nil {
		logger, err := logging.NewLogger(logging.INFO, logging.TEXT)
		if err != nil {
			return nil, fmt.Errorf("failed to create logger: %v", err)
		}
		loggerWrapper = logging.NewLoggerWrapper(logger)
	} else {
		loggerWrapper = opts.Logger
	}

	if opts != nil && opts.Client == nil {
		client = &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:          100,              // Max idle connections across all hosts
				MaxIdleConnsPerHost:   10,               // Max idle connections per host
				IdleConnTimeout:       90 * time.Second, // Keep idle connections for 90s
				TLSHandshakeTimeout:   10 * time.Second, // Timeout for TLS handshake
				ExpectContinueTimeout: 1 * time.Second,  // Wait time for 100-Continue responses
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,  // Connection timeout
					KeepAlive: 30 * time.Second, // TCP keep-alive time
				}).DialContext,
			},
			Timeout: 15 * time.Second, // Overall request timeout
		}
	} else {
		client = opts.Client
	}

	service := &UserServiceV1{
		status:  services.READY,
		client:  client,
		logger:  loggerWrapper,
		userURL: userURL,
	}

	service.logger.Infofln("[%s] Service created with URL: %s", service.Name(), userURL)
	return service, nil
}

// Status checks the health of the service by making a request to the
// healthcheck endpoint. It returns the service status based on the response.
func (s *UserServiceV1) Status() services.ServiceStatus {
	if s.status == services.STOPPED {
		s.logger.Errorfln("[%s] Service is stopped", s.Name())
		return s.status
	}

	resp, err := s.client.Get(fmt.Sprintf("%s/healthcheck", s.userURL))
	if err != nil {
		s.logger.Errorfln("[%s] Failed to check service status: %v", s.Name(), err)
		s.status = services.ERROR
		return s.status
	}

	var healthResp HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		s.logger.Errorfln("[%s] Failed to decode service health response: %v", s.Name(), err)
		s.status = services.ERROR
		return s.status
	}

	if healthResp.Status == UP {
		s.logger.Infofln("[%s] Service is up and running", s.Name())
		s.status = services.READY
	} else {
		s.logger.Errorfln("[%s] Service is down, status: %s", s.Name(), healthResp.Status)
		s.status = services.ERROR
	}

	return s.status
}

// Name returns the name of the service.
func (s *UserServiceV1) Name() string {
	return "User Service V1"
}

// Close stops the service and releases any resources it holds.
func (s *UserServiceV1) Close() error {
	if s.status == services.STOPPED {
		s.logger.Errorfln("[%s] Service is already stopped", s.Name())
		return nil
	}

	s.logger.Infofln("[%s] Closing service", s.Name())
	s.status = services.STOPPED
	s.logger.Infofln("[%s] Service stopped", s.Name())
	return nil
}

func (s *UserServiceV1) GetUserAuth(context context.Context, identifier string) (*UserGetAuthResponse, error) {
	url := helper.EncodeURLPath(fmt.Sprintf(`%s/users/auth-info?identifier=%s`, s.userURL, identifier))
	method := http.MethodGet
	header := http.Header{
		"Content-Type": {"application/json"},
	}

	req, err := http.NewRequestWithContext(context, method, url, nil)
	if err != nil {
		s.logger.Errorfln("[%s] Error creating request: %v", s.Name(), err)
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header = header

	s.logger.Debugfln("[%s] Making %s request to %s", s.Name(), method, url)
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Errorfln("[%s] Error making request: %v", s.Name(), err)
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	var userGetAuthResponse UserGetAuthResponse
	userGetAuthResponse.StatusCode = resp.StatusCode
	if err := json.NewDecoder(resp.Body).Decode(&userGetAuthResponse.Body); err != nil {
		s.logger.Errorfln("[%s] Error decoding response: %v", s.Name(), err)
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	s.logger.Debugfln("[%s] Received response: %v", s.Name(), userGetAuthResponse)
	return &userGetAuthResponse, nil
}

func (s *UserServiceV1) GetUserById(context context.Context, userId int64) (*UserGetResponse, error) {
	url := helper.EncodeURLPath(fmt.Sprintf(`%s/users/%d`, s.userURL, userId))
	method := http.MethodGet
	header := http.Header{
		"Content-Type": {"application/json"},
	}

	req, err := http.NewRequestWithContext(context, method, url, nil)
	if err != nil {
		s.logger.Errorfln("[%s] Error creating request: %v", s.Name(), err)
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header = header

	s.logger.Debugfln("[%s] Making %s request to %s", s.Name(), method, url)
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Errorfln("[%s] Error making request: %v", s.Name(), err)
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	var userGetResponse UserGetResponse
	userGetResponse.StatusCode = resp.StatusCode
	if err := json.NewDecoder(resp.Body).Decode(&userGetResponse.Body); err != nil {
		s.logger.Errorfln("[%s] Error decoding response: %v", s.Name(), err)
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	s.logger.Debugfln("[%s] Received response: %v", s.Name(), userGetResponse)
	return &userGetResponse, nil
}

func (s *UserServiceV1) IncrementSessionVersion(context context.Context, userId int64) error {
	url := helper.EncodeURLPath(fmt.Sprintf(`%s/users/%d/session-version/increment`, s.userURL, userId))
	method := http.MethodPost
	header := http.Header{}

	req, err := http.NewRequestWithContext(context, method, url, nil)
	if err != nil {
		s.logger.Errorfln("[%s] Error creating request: %v", s.Name(), err)
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header = header

	s.logger.Debugfln("[%s] Making %s request to %s", s.Name(), method, url)
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Errorfln("[%s] Error making request: %v", s.Name(), err)
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Errorfln("[%s] Failed to increment session version, status code: %d", s.Name(), resp.StatusCode)
		return fmt.Errorf("failed to increment session version, status code: %d", resp.StatusCode)
	}

	s.logger.Debugfln("[%s] Successfully incremented session version for user %d", s.Name(), userId)
	return nil
}

func (s *UserServiceV1) RevokeAllSessions(context context.Context, userId int64) error {
	url := helper.EncodeURLPath(fmt.Sprintf(`%s/users/%d/sessions/revoke`, s.userURL, userId))
	method := http.MethodPost
	header := http.Header{}

	req, err := http.NewRequestWithContext(context, method, url, nil)
	if err != nil {
		s.logger.Errorfln("[%s] Error creating request: %v", s.Name(), err)
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header = header

	s.logger.Debugfln("[%s] Making %s request to %s", s.Name(), method, url)
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Errorfln("[%s] Error making request: %v", s.Name(), err)
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Errorfln("[%s] Failed to revoke all sessions for user %d, status code: %d", s.Name(), userId, resp.StatusCode)
		return fmt.Errorf("failed to revoke all sessions for user %d, status code: %d", userId, resp.StatusCode)
	}

	s.logger.Debugfln("[%s] Successfully revoked all sessions for user %d", s.Name(), userId)
	return nil
}

func (s *UserServiceV1) RevokeSession(context context.Context, userId int64, sessionId int64) error {
	url := helper.EncodeURLPath(fmt.Sprintf(`%s/users/%d/sessions/%d/revoke`, s.userURL, userId, sessionId))
	method := http.MethodPost
	header := http.Header{}

	req, err := http.NewRequestWithContext(context, method, url, nil)
	if err != nil {
		s.logger.Errorfln("[%s] Error creating request: %v", s.Name(), err)
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header = header

	s.logger.Debugfln("[%s] Making %s request to %s", s.Name(), method, url)
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Errorfln("[%s] Error making request: %v", s.Name(), err)
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Errorfln("[%s] Failed to revoke session %d for user %d, status code: %d", s.Name(), sessionId, userId, resp.StatusCode)
		return fmt.Errorf("failed to revoke session %d for user %d, status code: %d", sessionId, userId, resp.StatusCode)
	}

	s.logger.Debugfln("[%s] Successfully revoked session %d for user %d", s.Name(), sessionId, userId)
	return nil
}

func (s *UserServiceV1) CreateUser(context context.Context, payload UserPostRequestBody) (*UserPostResponse, error) {
	url := helper.EncodeURLPath(fmt.Sprintf(`%s/users`, s.userURL))
	method := http.MethodPost
	header := http.Header{
		"Content-Type": {"application/json"},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.Errorfln("[%s] Error marshalling request body: %v", s.Name(), err)
		return nil, fmt.Errorf("error marshalling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(context, method, url, io.Reader(bytes.NewReader(bodyBytes)))
	if err != nil {
		s.logger.Errorfln("[%s] Error creating request: %v", s.Name(), err)
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header = header

	s.logger.Debugfln("[%s] Making %s request to %s", s.Name(), method, url)
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Errorfln("[%s] Error making request: %v", s.Name(), err)
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	var userResp UserPostResponse
	userResp.StatusCode = resp.StatusCode
	if err := json.NewDecoder(resp.Body).Decode(&userResp.Body); err != nil {
		s.logger.Errorfln("[%s] Error decoding response: %v", s.Name(), err)
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	s.logger.Debugfln("[%s] Received response: %v", s.Name(), userResp)
	return &userResp, nil
}

func (s *UserServiceV1) CreateSession(context context.Context, userId int64, payload SessionPostRequestBody) (*SessionPostResponse, error) {
	url := helper.EncodeURLPath(fmt.Sprintf(`%s/users/%d/sessions`, s.userURL, userId))
	method := http.MethodPost
	header := http.Header{
		"Content-Type": {"application/json"},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.Errorfln("[%s] Error marshalling request body: %v", s.Name(), err)
		return nil, fmt.Errorf("error marshalling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(context, method, url, io.Reader(bytes.NewReader(bodyBytes)))
	if err != nil {
		s.logger.Errorfln("[%s] Error creating request: %v", s.Name(), err)
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header = header

	s.logger.Debugfln("[%s] Making %s request to %s with payload: %s", s.Name(), method, url, string(bodyBytes))
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Errorfln("[%s] Error making request: %v", s.Name(), err)
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	var sessionResp SessionPostResponse
	sessionResp.StatusCode = resp.StatusCode
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp.Body); err != nil {
		s.logger.Errorfln("[%s] Error decoding response: %v", s.Name(), err)
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	s.logger.Debugfln("[%s] Received response: %v", s.Name(), sessionResp)
	return &sessionResp, nil
}
