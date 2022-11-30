// Package appPref provides path retrievers for Lominus preferences files.
package appPref

import (
	"fmt"
	"path/filepath"
	"time"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	"github.com/beebeeoii/lominus/internal/file"
	"github.com/boltdb/bolt"
)

const PREFERENCES_FILE_NAME = appConstants.PREFERENCES_FILE_NAME

// Preferences struct describes the data being stored in the user's preferences file.
type Preferences struct {
	Directory string
	Frequency int
	LogLevel  string
}

// GetJwtPath returns the file path to user's preferences.
func GetPreferencesPath() (string, error) {
	var preferencesPath string

	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		return preferencesPath, retrieveBaseDirErr
	}

	preferencesPath = filepath.Join(baseDir, appConstants.PREFERENCES_FILE_NAME)

	return preferencesPath, nil
}

// SavePreferences saves the user's preferences data onto local storage.
func SavePreferences(filePath string, preferences Preferences) error {
	return file.EncodeStructToFile(filePath, preferences)
}

// LoadPreferences loads the user's preferences data from local storage.
func LoadPreferences(filePath string) (Preferences, error) {
	preferences := Preferences{}
	if !file.Exists(filePath) {
		return preferences, &file.FileNotFoundError{FileName: filePath}
	}
	err := file.DecodeStructFromFile(filePath, &preferences)

	return preferences, err
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
