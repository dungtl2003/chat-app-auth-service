package tests

import (
	"database/sql"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/httpclient"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq" // postgresql driver support
)

type Helper struct {
	Db           *Database
	Client       *httpclient.HttpClient
	AuthURL      string
	logger       *slog.Logger
	JwtSecret    string
	ATDurationMs int
}

// SuckDelay is a function that blocks the current goroutine for a specified
// duration in milliseconds. It uses a busy wait loop to achieve this.
// Sleep is not used to avoid blocking the entire process.
func SuckDelay(ms int) {
	start := time.Now()
	duration := time.Duration(ms) * time.Millisecond
	for start.Add(duration).After(time.Now()) {
	}
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

func GetATFromResponse(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	bodyAsByteArray, err := httpclient.ReadResponse(resp)
	if err != nil {
		return ""
	}

	jsonMap := make(map[string]any)
	err = json.Unmarshal(bodyAsByteArray, &jsonMap)
	if err != nil {
		return ""
	}

	return jsonMap["access_token"].(string)
}

func GetRespJson(resp *http.Response) (map[string]any, error) {
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}
	bodyAsByteArray, err := httpclient.ReadResponse(resp)
	if err != nil {
		return nil, err
	}

	jsonMap := make(map[string]any)
	err = json.Unmarshal(bodyAsByteArray, &jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonMap, nil
}

func NewHelper() *Helper {
	ATDurationMsStr, bool := os.LookupEnv("ACCESS_TOKEN_DURATION_MS")
	if !bool {
		log.Fatal("ACCESS_TOKEN_DURATION_MS is not set")
	}
	ATDurationMs, err := strconv.Atoi(ATDurationMsStr)
	if err != nil || ATDurationMs < 0 {
		log.Fatal("ACCESS_TOKEN_DURATION_MS must be a non-negative number")
	}

	jwtSecret, bool := os.LookupEnv("JWT_SECRET")
	if !bool {
		log.Fatal("JWT_SECRET is not set")
	}

	authUrl, bool := os.LookupEnv("AUTH_URL")
	if !bool {
		log.Fatal("AUTH_URL is not set")
	}

	dbUrl, bool := os.LookupEnv("ADMIN_DATABASE_URL")
	if !bool {
		log.Fatal("ADMIN_DATABASE_URL is not set")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	db, err := NewDb(dbUrl, logger)
	if err != nil {
		log.Fatalf("Error when creating database connection: %v", err)
	}

	client := httpclient.NewWithConfig(&http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:     true,
			ResponseHeaderTimeout: 15 * time.Second,
		},
		Timeout: 15 * time.Second,
	})

	return &Helper{
		Db:           db,
		Client:       client,
		logger:       logger,
		AuthURL:      authUrl,
		JwtSecret:    jwtSecret,
		ATDurationMs: ATDurationMs,
	}
}

func (h *Helper) Snapshot() error {
	h.logger.Info("Taking snapshot")
	return h.Db.Snapshot()
}

func (h *Helper) Rollback() error {
	var err error
	err = nil
	defer func() {
		h.logger.Info("Closing connection")
		h.Db.Close()
	}()
	h.logger.Info("Rolling back")
	err = h.Db.Rollback()

	return err
}

type Database struct {
	client *sql.DB
	logger helper.LoggerWrapper
}

// New creates a new database connection. The function returns a database
// connection and an error.
func NewDb(url string, logger *slog.Logger) (*Database, error) {
	client, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	return &Database{
		client: client,
		logger: helper.NewLoggerWrapper(logger),
	}, nil
}

// Close closes the database connection. The function returns an error.
func (d *Database) Close() error {
	d.logger.Info("closing database connection")
	if err := d.client.Close(); err != nil {
		d.logger.Error("error when closing database connection", "error", err)
		return err
	} else {
		d.logger.Info("database connection closed")
		return nil
	}
}

// Snapshot creates a snapshot of the current database state. The function is currently used for testing purposes.
// The function returns an error. You can use Rollback() to revert the database to the state before the snapshot.
func (d *Database) Snapshot() error {
	tx, err := d.client.Begin()
	if err != nil {
		d.logger.Error("error when starting transaction", "error", err)
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = d.client.Exec(`CREATE TABLE IF NOT EXISTS chat_user.chat_user_snapshot AS SELECT * FROM chat_user.chat_user WHERE false;`) // create an empty table
	if err != nil {
		msg := fmt.Sprintf("error when creating snapshot: error when creating table chat_user_snapshot: %v", err)
		d.logger.Error(msg)
		return err
	}

	_, err = d.client.Exec(`CREATE TABLE IF NOT EXISTS chat_user.device_snapshot AS SELECT * FROM chat_user.device WHERE false;`) // create an empty table
	if err != nil {
		msg := fmt.Sprintf("error when creating snapshot: error when creating table device_snapshot: %v", err)
		d.logger.Error(msg)
		return err
	}

	// we need to make sure that the snapshot tables are empty
	_, err = d.client.Exec(`DELETE FROM chat_user.chat_user_snapshot;`)
	if err != nil {
		msg := fmt.Sprintf("error when creating snapshot: error when deleting data from chat_user_snapshot: %v", err)
		d.logger.Error(msg)
		return err
	}

	_, err = d.client.Exec(`INSERT INTO chat_user.chat_user_snapshot SELECT * FROM chat_user.chat_user;`)
	if err != nil {
		msg := fmt.Sprintf("error when creating snapshot: error when inserting data into chat_user_snapshot: %v", err)
		d.logger.Error(msg)
		return err
	}

	// insert data from the original tables to the snapshot tables
	_, err = d.client.Exec(`INSERT INTO chat_user.device_snapshot SELECT * FROM chat_user.device;`)
	if err != nil {
		msg := fmt.Sprintf("error when creating snapshot: error when inserting data into device_snapshot: %v", err)
		d.logger.Error(msg)
		return err
	}

	return nil
}

// Rollback rolls back the database to the state before the snapshot. The function is currently used for testing purposes.
// The function returns an error. You must call Snapshot() before calling this function.
func (d *Database) Rollback() error {
	tx, err := d.client.Begin()
	if err != nil {
		d.logger.Error("error when starting transaction", "error", err)
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// we will migrate the data from the snapshot tables to the original tables
	_, err = d.client.Exec(`DELETE FROM chat_user.chat_user;`) // cascade delete
	if err != nil {
		msg := fmt.Sprintf("error when rolling back: error when deleting data from chat_user: %v", err)
		d.logger.Error(msg)
		return err
	}

	_, err = d.client.Exec(`INSERT INTO chat_user.chat_user SELECT * FROM chat_user.chat_user_snapshot;`)
	if err != nil {
		msg := fmt.Sprintf("error when rolling back: error when inserting data into chat_user: %v", err)
		d.logger.Error(msg)
		return err
	}

	_, err = d.client.Exec(`INSERT INTO chat_user.device SELECT * FROM chat_user.device_snapshot;`)
	if err != nil {
		msg := fmt.Sprintf("error when rolling back: error when inserting data into device: %v", err)
		d.logger.Error(msg)
		return err
	}

	_, err = d.client.Exec(`DROP TABLE chat_user.device_snapshot;`)
	if err != nil {
		msg := fmt.Sprintf("error when rolling back: error when dropping device_snapshot: %v", err)
		d.logger.Error(msg)
		return err
	}

	_, err = d.client.Exec(`DROP TABLE chat_user.chat_user_snapshot;`)
	if err != nil {
		msg := fmt.Sprintf("error when rolling back: error when dropping chat_user_snapshot: %v", err)
		d.logger.Error(msg)
		return err
	}

	err = tx.Commit()
	if err != nil {
		msg := fmt.Sprintf("error when rolling back: error when committing transaction: %v", err)
		d.logger.Error(msg)
		return err
	}

	return nil
}
