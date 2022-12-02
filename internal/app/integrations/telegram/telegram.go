// Package intTelegram provides path retrievers for Lominus Telegram integration files.
package intTelegram

import (
	"path/filepath"
	"time"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	"github.com/boltdb/bolt"
)

// SaveTelegramCredentials saves the user's Telegram userId and botId locally.
func SaveTelegramCredentials(userId string, botId string) error {
	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		return retrieveBaseDirErr
	}

	dbFName := filepath.Join(baseDir, appConstants.DATABASE_FILE_NAME)
	db, dbErr := bolt.Open(dbFName, 0600, &bolt.Options{Timeout: 3 * time.Second})

	if dbErr != nil {
		return dbErr
	}

	defer db.Close()

	updateErr := db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("Integrations")).Put([]byte("telegramUserId"), []byte(userId))
		err1 := tx.Bucket([]byte("Integrations")).Put([]byte("telegramBotId"), []byte(botId))

		if err != nil {
			return err
		}

		return err1
	})

	return updateErr
}
