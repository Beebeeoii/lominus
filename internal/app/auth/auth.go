// Package appAuth provides path retrievers for Lominus auth files.
package appAuth

import (
	"path/filepath"

	"github.com/beebeeoii/lominus/internal/app"
	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	"github.com/boltdb/bolt"
)

type CanvasCredentials struct {
	CanvasApiToken string
}

func GetCanvasCredentials() (CanvasCredentials, error) {
	dbInstance := app.GetDBInstance()
	var canvasCredentials CanvasCredentials

	err := dbInstance.View(func(tx *bolt.Tx) error {
		authBucket := tx.Bucket([]byte("Auth"))
		canvasToken := string(authBucket.Get([]byte("canvasToken")))

		canvasCredentials.CanvasApiToken = canvasToken

		return nil
	})

	if err != nil {
		return CanvasCredentials{}, err
	}

	return canvasCredentials, nil
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

// SaveCanvasCredentials saves the user's Canvas API token locally.
func SaveCanvasCredentials(cred CanvasCredentials) error {
	dbInstance := app.GetDBInstance()

	updateErr := dbInstance.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("Auth")).Put([]byte("canvasToken"), []byte(cred.CanvasApiToken))
		return err
	})

	return updateErr
}
