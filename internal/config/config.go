package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type LogConfig struct {
	Level  string // INFO, DEBUG, WARN, ERROR
	Kind   string // TEXT or JSON
	Logger *slog.Logger
}

type JwtTokenConfig struct {
	JwtSecret    string
	ATDurationMs int64
	RTDurationMs int64
}

func (t JwtTokenConfig) String() string {
	parts := []string{
		fmt.Sprintf("JWT_SECRET: %s", t.JwtSecret),
		fmt.Sprintf("AT_DURATION_MS: %d", t.ATDurationMs),
		fmt.Sprintf("RT_DURATION_MS: %d", t.RTDurationMs),
	}

	return fmt.Sprintf("JwtTokenConfig{%s}", strings.Join(parts, ", "))
}

type Config struct {
	ServerPort     int
	Env            string
	LogConfig      *LogConfig
	UserServiceURL string
	JwtTokenConfig *JwtTokenConfig
	DomainName     string
	Cost           int
}

func (l *LogConfig) String() string {
	parts := []string{
		fmt.Sprintf("LEVEL: %s", l.Level),
		fmt.Sprintf("KIND: %s", l.Kind),
	}

	return fmt.Sprintf("LogConfig{%s}", strings.Join(parts, ", "))
}

func (c *Config) String() string {
	logConfigPart := "nil"
	if c.LogConfig != nil {
		logConfigPart = fmt.Sprintf("%s", c.LogConfig)
	}

	jwtTokenConfigPart := "nil"
	if c.JwtTokenConfig != nil {
		jwtTokenConfigPart = fmt.Sprintf("%s", c.JwtTokenConfig)
	}

	parts := []string{
		fmt.Sprintf("SERVER_PORT: %d", c.ServerPort),
		fmt.Sprintf("LOGGER: %s", logConfigPart),
		fmt.Sprintf("ENV: %s", c.Env),
		fmt.Sprintf("USER_SERVICE_URL: %s", c.UserServiceURL),
		fmt.Sprintf("JWT: %s", jwtTokenConfigPart),
		fmt.Sprintf("DOMAIN_NAME: %s", c.DomainName),
		fmt.Sprintf("COST: %d", c.Cost),
	}

	return fmt.Sprintf("Config{%s}", strings.Join(parts, ", "))
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
	err = c.setUserServiceURL()
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
	err = c.setCost()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) setCost() error {
	log.Println("Setting COST")
	costStr, has := os.LookupEnv("COST")
	if !has {
		log.Println("COST not found, setting to 12")
		costStr = "12"
	}

	cost, err := strconv.Atoi(costStr)
	if err != nil {
		return fmt.Errorf("Invalid cost number: %s", costStr)
	}

	if cost < 1 || cost > 31 {
		return fmt.Errorf("Cost number out of range: %s (4-31)", costStr)
	}

	c.Cost = cost
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

	tokConfig := &JwtTokenConfig{
		JwtSecret:    secretKey,
		RTDurationMs: rtDurationMs,
		ATDurationMs: atDurationMs,
	}

	c.JwtTokenConfig = tokConfig
	return nil
}

func (c *Config) setUserServiceURL() error {
	log.Println("Setting USER_SERVICE_URL")
	userServiceURL, has := os.LookupEnv("USER_SERVICE_URL")
	if !has {
		return fmt.Errorf("USER_SERVICE_URL is required")
	}

	c.UserServiceURL = userServiceURL
	return nil
}

func (c *Config) setLogConfig() error {
	log.Println("Setting LOG_LEVEL and LOG_KIND")
	logLevel, err := getLogLevel()
	if err != nil {
		return err
	}
	logKind, err := getLogKind()
	if err != nil {
		return err
	}

	writer := os.Stdout

	var slogLogLevel slog.Level
	switch logLevel {
	case "DEBUG":
		slogLogLevel = slog.LevelDebug
	case "INFO":
		slogLogLevel = slog.LevelInfo
	case "WARN":
		slogLogLevel = slog.LevelWarn
	case "ERROR":
		slogLogLevel = slog.LevelError
	}

	var handler slog.Handler
	if logKind == "TEXT" {
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slogLogLevel})
	} else {
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slogLogLevel})
	}

	c.LogConfig = &LogConfig{
		Level:  logLevel,
		Kind:   logKind,
		Logger: slog.New(handler),
	}
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
	log.Println("Setting ENV")
	env, has := os.LookupEnv("ENV")
	if !has {
		log.Println("ENV not found, setting to dev")
		env = "dev"
	}
	c.Env = env

	return nil
}

func getLogLevel() (string, error) {
	logLevel, has := os.LookupEnv("LOG_LEVEL")
	if !has {
		logLevel = "INFO"
	}

	if logLevel != "INFO" && logLevel != "DEBUG" && logLevel != "WARN" && logLevel != "ERROR" {
		return "", fmt.Errorf("`LOG_LEVEL=%s` is invalid. It can only be `INFO`, `DEBUG`, `WARN` or `ERROR`\n", logLevel)
	}

	return logLevel, nil
}

func getLogKind() (string, error) {
	kind, has := os.LookupEnv("LOG_KIND")
	if !has {
		kind = "TEXT"
	}

	if kind != "TEXT" && kind != "JSON" {
		return "", fmt.Errorf("`LOG_KIND=%s` is invalid, it can only be `TEXT` or `JSON`", kind)
	}

	return kind, nil
}
