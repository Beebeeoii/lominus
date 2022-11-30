// Package appPref provides path retrievers for Lominus preferences files.
package appPref

import (
	"fmt"
	"path/filepath"
	"time"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	"github.com/boltdb/bolt"
)

// Preferences struct describes the data being stored in the user's preferences file.
type Preferences struct {
	Directory string
	Frequency int
	LogLevel  string
}

// SaveRootSyncDirectory saves the user's root sync directory locally.
func SaveRootSyncDirectory(directory string) error {
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
		err := tx.Bucket([]byte("Preferences")).Put([]byte("directory"), []byte(directory))
		return err
	})

	return updateErr
}

// SaveSyncFrequency saves the user's sync frequency locally.
func SaveSyncFrequency(frequency int) error {
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
		err := tx.Bucket([]byte("Preferences")).Put([]byte("frequency"), []byte(fmt.Sprint(frequency)))
		return err
	})

	return updateErr
}

// SaveDebugMode saves the user's chosen debug mode locally.
func SaveDebugMode(logLevel string) error {
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
		err := tx.Bucket([]byte("Preferences")).Put([]byte("logLevel"), []byte(logLevel))
		return err
	})

	return updateErr
}
