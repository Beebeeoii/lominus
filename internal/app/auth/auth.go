// Package appAuth provides path retrievers for Lominus auth files.
package appAuth

import (
	"path/filepath"
	"time"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	"github.com/boltdb/bolt"
)

type CanvasCredentials struct {
	CanvasApiToken string
}

// GetJwtPath returns the file path to user's JWT data.
func GetTokensPath() (string, error) {
	var jwtPath string

	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		return jwtPath, retrieveBaseDirErr
	}

	jwtPath = filepath.Join(baseDir, appConstants.TOKENS_FILE_NAME)

	return jwtPath, nil
}

// Save saves the user's Canvas API token locally.
func (cred CanvasCredentials) Save() error {
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
		err := tx.Bucket([]byte("Auth")).Put([]byte("canvasToken"), []byte(cred.CanvasApiToken))
		return err
	})

	return updateErr
}
