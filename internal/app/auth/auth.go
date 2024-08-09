// Package appAuth provides path retrievers for Lominus auth files.
package appAuth

import (
	"github.com/beebeeoii/lominus/internal/app"
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

// SaveCanvasCredentials saves the user's Canvas API token locally.
func SaveCanvasCredentials(cred CanvasCredentials) error {
	dbInstance := app.GetDBInstance()

	updateErr := dbInstance.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("Auth")).Put([]byte("canvasToken"), []byte(cred.CanvasApiToken))
		return err
	})

	return updateErr
}
