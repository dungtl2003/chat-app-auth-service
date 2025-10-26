package database

import (
	"database/sql"
	"dungtl2003/chat-app-auth-service/internal/helper"
	"dungtl2003/chat-app-auth-service/internal/logging"
	"dungtl2003/chat-app-auth-service/internal/model"
	"dungtl2003/chat-app-auth-service/internal/password"
	"dungtl2003/chat-app-auth-service/internal/services"
	"encoding/json"
	"os"

	_ "github.com/lib/pq" // postgresql driver support
)

type DataFile struct {
	UserFile  string
	AssetFile string
}

type DatabaseService struct {
	client *sql.DB
	logger *logging.LoggerWrapper
	status services.ServiceStatus
}

// New creates a new database connection. The function returns a database connection and an error.
func New(url string, logger *logging.LoggerWrapper) (*DatabaseService, error) {
	client, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	d := &DatabaseService{
		client: client,
		logger: logger,
		status: services.ServiceReady,
	}

	d.logger.Infofln("[%s] Database connection created", d.Name())
	d.logger.Infofln("[%s] Running", d.Name())
	return d, nil
}

func (d *DatabaseService) Name() string {
	return "Database Service"
}

func (d *DatabaseService) Status() services.ServiceStatus {
	if d.status != services.ServiceStopped {
		// check if the database connection is still alive
		if err := d.client.Ping(); err != nil {
			d.logger.Errorfln("[%s] Database connection is not alive: %v", d.Name(), err)
			d.status = services.ServiceError
		} else {
			d.status = services.ServiceReady
		}
	}

	return d.status
}

// Close closes the database connection. The function returns an error.
func (d *DatabaseService) Close() error {
	if d.Status() == services.ServiceStopped {
		d.logger.Errorfln("[%s] Database connection is already closed", d.Name())
		return nil
	}

	err := d.client.Close()
	if err != nil {
		d.logger.Errorfln("[%s] Failed to close database connection: %v", d.Name(), err)
		d.status = services.ServiceError
	} else {
		d.logger.Infofln("[%s] Database connection closed", d.Name())
		d.status = services.ServiceStopped
	}

	return err
}

// CreateTemporaryData creates temporary data in the database for testing purposes.
func (d *DatabaseService) CreateTemporaryData(dataFile DataFile, passwordManager password.PasswordManager) error {
	tx, err := d.client.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// This must be done first to ensure that the assets are created before so
	// that other data can reference them.
	if dataFile.AssetFile != "" {
		assetData, err := os.ReadFile(dataFile.AssetFile)
		if err != nil {
			return err
		}
		var assets []model.Asset
		err = json.Unmarshal(assetData, &assets)
		if err != nil {
			return err
		}
		d.logger.Debugfln("[%s] Loaded %d assets from file: %s", d.Name(), len(assets), dataFile.AssetFile)
		for _, asset := range assets {
			query := `INSERT INTO media.asset (
				id, public_id, width, height, format, resource_type, created_at, bytes, url, secure_url, asset_folder, original_filename, api_key
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
			);`
			args := []any{
				asset.Id, asset.PublicId, asset.Width, asset.Height, asset.Format, asset.ResourceType, asset.CreatedAt, asset.Bytes, asset.Url, asset.SecureUrl, asset.AssetFolder, asset.OriginalFilename, asset.ApiKey,
			}
			d.logger.Debugfln("[%s] Executing query: %s with args: %v", d.Name(), helper.StripWS(query), args)
			_, err = tx.Exec(query, args...)
			if err != nil {
				return err
			}
		}
	}

	if dataFile.UserFile != "" {
		userData, err := os.ReadFile(dataFile.UserFile)
		if err != nil {
			return err
		}
		var users []model.ChatUser
		err = json.Unmarshal(userData, &users)
		if err != nil {
			return err
		}

		d.logger.Debugfln("[%s] Loaded %d users from file: %s", d.Name(), len(users), dataFile.UserFile)
		for _, user := range users {
			hashedPassword, err := passwordManager.Hash(user.Password)
			if err != nil {
				return err
			}
			query := `INSERT INTO chat_user.chat_user (
            id, email, username, password, role, first_name, last_name, birthday, gender, phone_number, privacy, avatar_id, deleted_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
        );`
			args := []any{user.Id, user.Email, user.Username, hashedPassword, user.Role, user.FirstName, user.LastName, user.Birthday, user.Gender, user.PhoneNumber, user.Privacy, user.AvatarId, user.DeletedAt}
			d.logger.Debugfln("[%s] Executing query: %s with args: %v", d.Name(), helper.StripWS(query), args)
			_, err = tx.Exec(query, args...)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// ClearAllData clears all data in the database. This function is used for
// testing purposes.
func (d *DatabaseService) ClearAllData() error {
	queries := []string{
		"DELETE FROM chat_user.chat_user;",
		"DELETE FROM media.asset;",
		"DELETE FROM conversation.conversation;",
	}

	for _, query := range queries {
		d.logger.Debugfln("query: %s", query)
		_, err := d.client.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}
