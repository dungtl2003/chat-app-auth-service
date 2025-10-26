package config

import (
	"dungtl2003/chat-app-auth-service/internal/logging"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Env string

const (
	Dev  Env = "dev"
	Prod Env = "prod"
	Test Env = "test"
)

type JwtTokenConfig struct {
	JwtSecret    string
	ATDurationMs int64
	RTDurationMs int64
}

type LogConfig struct {
	Level logging.LoggerLevel
	Kind  logging.LoggerKind
}

type UserServiceConfig struct {
	URL string
}

type IdGeneratorConfig struct {
	Addr    string
	CertDir string
}

type PasswordManagerConfig struct {
	Cost int
}

type Config struct {
	ServerPort            int
	Env                   Env
	LogConfig             LogConfig
	UserServiceConfig     UserServiceConfig
	JwtTokenConfig        JwtTokenConfig
	DomainName            string
	PasswordManagerConfig PasswordManagerConfig
	IdGeneratorConfig     IdGeneratorConfig
}

// LoadConfig loads the configuration from env file. It will return Config instance
// or error if occurs.
func LoadConfig() (*Config, error) {
	c := &Config{}
	err := c.setEnv()
	if err != nil {
		return nil, err
	}
	err = c.setServerPort()
	if err != nil {
		return nil, err
	}
	err = c.setLogConfig()
	if err != nil {
		return nil, err
	}
	err = c.setUserServiceConfig()
	if err != nil {
		return nil, err
	}
	err = c.setJwtTokenConfig()
	if err != nil {
		return nil, err
	}
	err = c.setDomainName()
	if err != nil {
		return nil, err
	}
	err = c.setPasswordManagerConfig()
	if err != nil {
		return nil, err
	}
	err = c.setIdGeneratorConfig()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (t JwtTokenConfig) String() string {
	parts := []string{
		fmt.Sprintf("JWT_SECRET: %s", t.JwtSecret),
		fmt.Sprintf("AT_DURATION_MS: %d", t.ATDurationMs),
		fmt.Sprintf("RT_DURATION_MS: %d", t.RTDurationMs),
	}

	return fmt.Sprintf("JwtTokenConfig{%s}", strings.Join(parts, ", "))
}

func (u UserServiceConfig) String() string {
	parts := []string{
		fmt.Sprintf("URL: %s", u.URL),
	}

	return fmt.Sprintf("UserServiceConfig{%s}", strings.Join(parts, ", "))
}

func (l LogConfig) String() string {
	parts := []string{
		fmt.Sprintf("LEVEL: %s", l.Level),
		fmt.Sprintf("KIND: %s", l.Kind),
	}

	return fmt.Sprintf("LogConfig{%s}", strings.Join(parts, ", "))
}

func (s IdGeneratorConfig) String() string {
	parts := []string{
		fmt.Sprintf("ADDR: %s", s.Addr),
		fmt.Sprintf("CERT_DIR: %s", s.CertDir),
	}

	return fmt.Sprintf("IdGeneratorConfig{%s}", strings.Join(parts, ", "))
}

func (p PasswordManagerConfig) String() string {
	parts := []string{
		fmt.Sprintf("PASSWORD_HASH_COST: %d", p.Cost),
	}
	return fmt.Sprintf("PasswordManagerConfig{%s}", strings.Join(parts, ", "))
}

func (c *Config) String() string {
	parts := []string{
		fmt.Sprintf("SERVER_PORT: %d", c.ServerPort),
		fmt.Sprintf("LOGGER: %s", c.LogConfig),
		fmt.Sprintf("ENV: %s", c.Env),
		fmt.Sprintf("USER_SERVICE_CONFIG: %s", c.UserServiceConfig),
		fmt.Sprintf("JWT: %s", c.JwtTokenConfig),
		fmt.Sprintf("DOMAIN_NAME: %s", c.DomainName),
		fmt.Sprintf("PASSWORD_MANAGER_CONFIG: %s", c.PasswordManagerConfig),
		fmt.Sprintf("ID_GENERATOR_CONFIG: %s", c.IdGeneratorConfig),
	}

	return fmt.Sprintf("Config{%s}", strings.Join(parts, ", "))
}

func (c *Config) setIdGeneratorConfig() error {
	c.IdGeneratorConfig = IdGeneratorConfig{}

	addr, has := os.LookupEnv("ID_GENERATOR_SERVICE_ADDR")
	if !has {
		return fmt.Errorf("ID_GENERATOR_SERVICE_ADDR not found")
	}

	certDir, has := os.LookupEnv("ID_GENERATOR_SERVICE_CERT_DIR")
	if !has {
		// return fmt.Errorf("ID_GENERATOR_SERVICE_CERT_DIR not found")
		certDir = ""
	}

	c.IdGeneratorConfig.Addr = addr
	c.IdGeneratorConfig.CertDir = certDir

	return nil
}

func (c *Config) setPasswordManagerConfig() error {
	log.Println("Setting PASSWORD_HASH_COST")
	costStr, has := os.LookupEnv("PASSWORD_HASH_COST")
	if !has {
		log.Println("PASSWORD_HASH_COST not found, setting to 12")
		costStr = "12"
	}

	cost, err := strconv.Atoi(costStr)
	if err != nil {
		return fmt.Errorf("Invalid cost number: %s", costStr)
	}

	if cost < 1 || cost > 31 {
		return fmt.Errorf("Cost number out of range: %s (4-31)", costStr)
	}

	c.PasswordManagerConfig.Cost = cost
	return nil
}

func (c *Config) setDomainName() error {
	log.Println("Setting DOMAIN_NAME")
	domainName, has := os.LookupEnv("DOMAIN_NAME")
	if !has {
		log.Println("DOMAIN_NAME not found, setting to localhost")
		domainName = "localhost"
	}
	c.DomainName = domainName
	return nil
}

func (c *Config) setJwtTokenConfig() error {
	log.Println("Setting JWT_SECRET, ACCESS_TOKEN_DURATION_MS and REFRESH_TOKEN_DURATION_MS")

	secretKey, has := os.LookupEnv("JWT_SECRET")
	if !has {
		return fmt.Errorf("JWT_SECRET is required")
	}

	atDurationMsStr, has := os.LookupEnv("ACCESS_TOKEN_DURATION_MS")
	if !has {
		log.Println("ACCESS_TOKEN_DURATION_MS not found, setting to 900000 (15 minutes)")
		atDurationMsStr = "900000"
	}
	atDurationMs, err := strconv.ParseInt(atDurationMsStr, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid access token duration: %s", atDurationMsStr)
	}
	if atDurationMs < 0 {
		return fmt.Errorf("Access token duration must be non-negative")
	}

	rtDurationMsStr, has := os.LookupEnv("REFRESH_TOKEN_DURATION_MS")
	if !has {
		log.Println("REFRESH_TOKEN_DURATION_MS not found, setting to 172800000 (2 days)")
		rtDurationMsStr = "172800000"
	}
	rtDurationMs, err := strconv.ParseInt(rtDurationMsStr, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid refresh token duration: %s", rtDurationMsStr)
	}
	if rtDurationMs < 0 {
		return fmt.Errorf("Refresh token duration must be non-negative")
	}

	tokConfig := JwtTokenConfig{
		JwtSecret:    secretKey,
		RTDurationMs: rtDurationMs,
		ATDurationMs: atDurationMs,
	}

	c.JwtTokenConfig = tokConfig
	return nil
}

func (c *Config) setUserServiceConfig() error {
	log.Println("Setting USER_SERVICE_URL")
	userServiceURL, has := os.LookupEnv("USER_SERVICE_URL")
	if !has {
		return fmt.Errorf("USER_SERVICE_URL is required")
	}

	c.UserServiceConfig.URL = userServiceURL
	return nil
}

func (c *Config) setLogConfig() error {
	logLevel := logging.INFO // Default log level
	logKind := logging.TEXT  // Default log kind

	logLevelStr, has := os.LookupEnv("LOG_LEVEL")
	if has {
		if !logging.IsValidLoggerLevel(logLevelStr) {
			return fmt.Errorf("`LOG_LEVEL=%s` is invalid", logLevelStr)
		}
		logLevel = logging.LoggerLevel(logLevelStr)
	}

	kind, has := os.LookupEnv("LOG_KIND")
	if has {
		if !logging.IsValidLoggerKind(kind) {
			return fmt.Errorf("`LOG_KIND=%s` is invalid", kind)
		}
		logKind = logging.LoggerKind(kind)
	}

	c.LogConfig.Level = logLevel
	c.LogConfig.Kind = logKind
	return nil
}

func (c *Config) setServerPort() error {
	log.Println("Setting PORT")
	portStr, has := os.LookupEnv("PORT")
	if !has {
		log.Println("PORT not found, setting to 8400")
		portStr = "8400"
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("Invalid port number: %s", portStr)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("Port number out of range: %s (1-65535)", portStr)
	}

	c.ServerPort = port
	return nil
}

func (c *Config) setEnv() error {
	log.Println("Setting ENVIRONMENT")
	envStr, has := os.LookupEnv("ENVIRONMENT")
	if !has {
		log.Println("ENVIRONMENT not found, setting to dev")
		envStr = "dev"
	}

	switch strings.ToLower(envStr) {
	case "dev":
		c.Env = Dev
	case "prod":
		c.Env = Prod
	case "test":
		c.Env = Test
	default:
		return fmt.Errorf("ENVIRONMENT=%s is invalid", envStr)
	}
	return nil
}
