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
	ApiVersion     string
	LogConfig      *LogConfig
	UserURL        string
	DeviceURL      string
	JwtTokenConfig *JwtTokenConfig
	DomainName     string
}

// New returns a new Config instance. Call Load() to set the configuration values.
func New() *Config {
	return &Config{}
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
		fmt.Sprintf("API_VERSION: %s", c.ApiVersion),
		fmt.Sprintf("USER_URL: %s", c.UserURL),
		fmt.Sprintf("DEVICE_URL: %s", c.DeviceURL),
		fmt.Sprintf("JWT: %s", jwtTokenConfigPart),
		fmt.Sprintf("DOMAIN_NAME: %s", c.DomainName),
	}

	return fmt.Sprintf("Config{%s}", strings.Join(parts, ", "))
}

// Load sets the configuration values.
func (c *Config) Load() {
	c.setEnv()
	c.setApiVersion()
	c.setServerPort()
	c.setLogConfig()
	c.setUserURL()
	c.setDeviceURL()
	c.setJwtTokenConfig()
	c.setDomainName()
}

func (c *Config) setDomainName() {
	log.Println("Setting DOMAIN_NAME")
	domainName, has := os.LookupEnv("DOMAIN_NAME")
	if !has {
		log.Println("DOMAIN_NAME not found, setting to localhost")
		domainName = "localhost"
	}
	c.DomainName = domainName
}

func (c *Config) setJwtTokenConfig() {
	log.Println("Setting JWT_SECRET, ACCESS_TOKEN_DURATION_MS and REFRESH_TOKEN_DURATION_MS")

	secretKey, has := os.LookupEnv("JWT_SECRET")
	if !has {
		log.Fatalln("JWT_SECRET is required")
	}

	atDurationMsStr, has := os.LookupEnv("ACCESS_TOKEN_DURATION_MS")
	if !has {
		log.Println("ACCESS_TOKEN_DURATION_MS not found, setting to 1800000 (30 minutes)")
		atDurationMsStr = "1800000"
	}
	atDurationMs, err := strconv.ParseInt(atDurationMsStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid access token duration: %s", atDurationMsStr)
	}
	if atDurationMs < 0 {
		log.Fatalf("Access token duration must be non-negative")
	}

	rtDurationMsStr, has := os.LookupEnv("REFRESH_TOKEN_DURATION_MS")
	if !has {
		log.Println("REFRESH_TOKEN_DURATION_MS not found, setting to 172800000 (2 days)")
		rtDurationMsStr = "172800000"
	}
	rtDurationMs, err := strconv.ParseInt(rtDurationMsStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid refresh token duration: %s", rtDurationMsStr)
	}
	if rtDurationMs < 0 {
		log.Fatalf("Refresh token duration must be non-negative")
	}

	tokConfig := &JwtTokenConfig{
		JwtSecret:    secretKey,
		RTDurationMs: rtDurationMs,
		ATDurationMs: atDurationMs,
	}

	c.JwtTokenConfig = tokConfig
}

func (c *Config) setUserURL() {
	log.Println("Setting USER_URL")
	userURL, has := os.LookupEnv("USER_URL")
	if !has {
		log.Fatalln("USER_URL is required")
	}

	c.UserURL = userURL
}

func (c *Config) setDeviceURL() {
	log.Println("Setting DEVICE_URL")
	deviceURL, has := os.LookupEnv("DEVICE_URL")
	if !has {
		log.Fatalln("DEVICE_URL is required")
	}

	c.DeviceURL = deviceURL
}

func (c *Config) setLogConfig() {
	log.Println("Setting LOG_LEVEL and LOG_KIND")
	logLevel := getLogLevel()
	logKind := getLogKind()

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
}

func (c *Config) setServerPort() {
	log.Println("Setting PORT")
	portStr, has := os.LookupEnv("PORT")
	if !has {
		log.Println("PORT not found, setting to 8400")
		portStr = "8400"
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid port number: %s", portStr)
	}

	if port < 1 || port > 65535 {
		log.Fatalf("Port number out of range: %s (1-65535)", portStr)
	}

	c.ServerPort = port
}

func (c *Config) setApiVersion() {
	log.Println("Setting API_VERSION")
	apiVersion, has := os.LookupEnv("API_VERSION")
	if !has {
		log.Println("API_VERSION not found, setting to v1")
		apiVersion = "v1"
	}
	c.ApiVersion = apiVersion
}

func (c *Config) setEnv() {
	log.Println("Setting ENV")
	env, has := os.LookupEnv("ENV")
	if !has {
		log.Println("ENV not found, setting to dev")
		env = "dev"
	}
	c.Env = env
}

func getLogLevel() string {
	logLevel, has := os.LookupEnv("LOG_LEVEL")
	if !has {
		logLevel = "INFO"
	}

	if logLevel != "INFO" && logLevel != "DEBUG" && logLevel != "WARN" && logLevel != "ERROR" {
		log.Fatalf("`LOG_LEVEL=%s` is invalid. It can only be `INFO`, `DEBUG`, `WARN` or `ERROR`\n", logLevel)
	}

	return logLevel
}

func getLogKind() string {
	kind, has := os.LookupEnv("LOG_KIND")
	if !has {
		kind = "TEXT"
	}

	if kind != "TEXT" && kind != "JSON" {
		log.Fatalf("`LOG_KIND=%s` is invalid, it can only be `TEXT` or `JSON`", kind)
	}

	return kind
}
