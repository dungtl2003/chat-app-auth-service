package tests

import (
	"context"
	"dungtl2003/chat-app-auth-service/internal/logging"
	"dungtl2003/chat-app-auth-service/internal/password"
	"dungtl2003/chat-app-auth-service/internal/server"
	"dungtl2003/chat-app-auth-service/internal/services/database"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type TestHelper struct {
	RedisClient          *redis.Client
	AdminDatabaseService *database.DatabaseService
	Client               *http.Client
	AuthURL              string
	Logger               *logging.LoggerWrapper
	JwtSecret            string
	ATDurationMs         int
	RTDurationMs         int
	DataFileDir          string

	passwordManager password.PasswordManager
	server          *server.AuthServer
}

type IdGeneratorConfig struct {
	TLSAddr     string
	NonTLSAddr  string
	CertDir     string
	FakeCertDir string
}

type SetUpOptions struct {
	DataFile      *database.DataFile
	ServerOptions *server.AuthServerOptions
}

func NewTestHelper() *TestHelper {
	ATDurationMsStr, bool := os.LookupEnv("ACCESS_TOKEN_DURATION_MS")
	if !bool {
		log.Fatal("ACCESS_TOKEN_DURATION_MS is not set")
	}
	ATDurationMs, err := strconv.Atoi(ATDurationMsStr)
	if err != nil || ATDurationMs < 0 {
		log.Fatal("ACCESS_TOKEN_DURATION_MS must be a non-negative number")
	}

	RTDurationMsStr, bool := os.LookupEnv("REFRESH_TOKEN_DURATION_MS")
	if !bool {
		log.Fatal("REFRESH_TOKEN_DURATION_MS is not set")
	}
	RTDurationMs, err := strconv.Atoi(RTDurationMsStr)
	if err != nil || RTDurationMs < 0 {
		log.Fatal("REFRESH_TOKEN_DURATION_MS must be a non-negative number")
	}

	jwtSecret, bool := os.LookupEnv("JWT_SECRET")
	if !bool {
		log.Fatal("JWT_SECRET is not set")
	}

	authUrl, bool := os.LookupEnv("AUTH_URL")
	if !bool {
		log.Fatal("AUTH_URL is not set")
	}

	dataFileDir, bool := os.LookupEnv("DATA_FILE_DIR")
	if !bool {
		log.Fatal("DATA_FILE_DIR is not set")
	}

	logger, err := logging.NewLogger(logging.DEBUG, logging.TEXT)
	if err != nil {
		log.Fatalf("Error when loading logger: %v", err)
	}
	loggerWrapper := logging.NewLoggerWrapper(logger)

	adminDbURL, has := os.LookupEnv("ADMIN_DATABASE_URL")
	if !has {
		log.Fatalf("Error when getting ADMIN_DATABASE_URL")
	}
	db, err := database.New(adminDbURL, loggerWrapper)
	if err != nil {
		log.Fatalf("Error when creating database: %v", err)
	}

	costStr, has := os.LookupEnv("COST")
	if !has {
		log.Println("COST not found, setting to 12")
		costStr = "12"
	}
	cost, err := strconv.Atoi(costStr)
	if err != nil {
		log.Fatalf("Invalid cost number: %s", costStr)
	}
	passwordManager, err := password.NewBcryptPasswordManager(cost)
	if err != nil {
		log.Fatalf("Error when creating password manager: %v", err)
	}

	loggerWrapper.Info("Creating http client")
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:     true,
			ResponseHeaderTimeout: 15 * time.Second,
		},
		Timeout: 15 * time.Second,
	}

	loggerWrapper.Info("Setting up redis client")
	redisUrl, has := os.LookupEnv("REDIS_URL")
	if !has {
		log.Fatalf("Error when getting REDIS_URL")
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisUrl,
	})

	h := &TestHelper{
		RedisClient:          redisClient,
		AdminDatabaseService: db,
		Client:               client,
		Logger:               loggerWrapper,
		AuthURL:              authUrl,
		JwtSecret:            jwtSecret,
		ATDurationMs:         ATDurationMs,
		RTDurationMs:         RTDurationMs,
		DataFileDir:          dataFileDir,
		passwordManager:      passwordManager,
	}

	h.Logger.Info("Test helper setup completed successfully")
	return h
}

func SetUp(t *TestHelper, opts *SetUpOptions) {
	if err := t.clearAllData(); err != nil {
		log.Fatalf("Error when clearing all data: %v", err)
	}

	if opts != nil && opts.DataFile != nil {
		t.Logger.Info("Creating test data")
		if err := t.createTestData(*opts.DataFile); err != nil {
			log.Fatalf("Error when creating test data: %v", err)
		}
	} else {
		t.Logger.Info("No test data provided, skipping data creation")
	}

	var serverOpts *server.AuthServerOptions
	serverOpts = nil
	if opts != nil {
		serverOpts = opts.ServerOptions
	}

	server, err := server.New(serverOpts)
	if err != nil {
		log.Fatalf("Error when creating server: %v", err)
	}
	t.server = server

	t.Logger.Info("Starting server")
	go func() {
		err := server.Run()
		if err != nil {
			log.Fatalf("Error when running server: %v", err)
		}
	}()

	err = waitForServer(fmt.Sprintf("%s/healthcheck", t.AuthURL), 5*time.Second)
	if err != nil {
		t.Logger.Errorfln("Server did not start in time: %v", err)
		log.Fatalf("Error waiting for server to start: %v", err)
	}
}

func TearDown(t *TestHelper) {
	t.Logger.Info("Tearing down test helper")

	t.Logger.Info("Clearing all redis data")
	if err := t.RedisClient.FlushAll(context.Background()).Err(); err != nil {
		log.Fatalf("Error when clearing redis data: %v", err)
	}

	t.Logger.Info("Closing redis client")
	if err := t.RedisClient.Close(); err != nil {
		log.Fatalf("Error when closing redis client: %v", err)
	}

	t.Logger.Info("Clearing all data")
	if err := t.clearAllData(); err != nil {
		log.Fatalf("Error when clearing all data: %v", err)
	}

	t.Logger.Info("Closing admin database service")
	if err := t.AdminDatabaseService.Close(); err != nil {
		log.Fatalf("Error when closing admin database service: %v", err)
	}

	t.Logger.Info("Closing server")
	err := t.server.Close()
	if err != nil {
		log.Fatalf("Error when closing server: %v", err)
	}

	t.Logger.Info("Test helper torn down successfully")
}

func GetRTFromResponse(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			return cookie.Value
		}
	}
	return ""
}

func Get(client *http.Client, url string, header http.Header) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header = header
	return client.Do(req)
}

func Post(client *http.Client, url string, header http.Header, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	if header != nil {
		req.Header = header
	}
	return client.Do(req)
}

func (t *TestHelper) clearAllData() error {
	t.Logger.Info("Clearing all data")
	if err := t.AdminDatabaseService.ClearAllData(); err != nil {
		t.Logger.Errorfln("Failed to clear all data: %v", err)
		return err
	}

	return nil
}

func (h *TestHelper) createTestData(dataFile database.DataFile) error {
	if dataFile.AssetFile != "" {
		dataFile.AssetFile = fmt.Sprintf("%s/%s", h.DataFileDir, dataFile.AssetFile)
	}
	if dataFile.UserFile != "" {
		dataFile.UserFile = fmt.Sprintf("%s/%s", h.DataFileDir, dataFile.UserFile)
	}

	h.Logger.Info("Creating temporary data")
	err := h.AdminDatabaseService.CreateTemporaryData(dataFile, h.passwordManager)
	if err != nil {
		return err
	}
	return nil
}

func waitForServer(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("server did not start at %s within %s", url, timeout)
}
