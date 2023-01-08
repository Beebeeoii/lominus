// Package appPref provides path retrievers for Lominus preferences files.
package appPref

import (
	"fmt"
	"strconv"

	"github.com/beebeeoii/lominus/internal/app"
	"github.com/boltdb/bolt"
)

// Preferences struct describes the data being stored in the user's preferences file.
type Preferences struct {
	Directory string
	Frequency int
	LogLevel  string
}

func GetPreferences() (Preferences, error) {
	dbInstance := app.GetDBInstance()
	var pref Preferences

	err := dbInstance.View(func(tx *bolt.Tx) error {
		prefBucket := tx.Bucket([]byte("Preferences"))
		directory := string(prefBucket.Get([]byte("directory")))
		frequency, _ := strconv.Atoi(string(prefBucket.Get([]byte("frequency"))))
		logLevel := string(prefBucket.Get([]byte("logLevel")))

		pref.Directory = directory
		pref.Frequency = frequency
		pref.LogLevel = logLevel

		return nil
	})

	if err != nil {
		return Preferences{}, err
	}

	return pref, nil
}

// SaveRootSyncDirectory saves the user's root sync directory locally.
func SaveRootSyncDirectory(directory string) error {
	dbInstance := app.GetDBInstance()

	updateErr := dbInstance.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("Preferences")).Put([]byte("directory"), []byte(directory))
		return err
	})

	return updateErr
}

// SaveSyncFrequency saves the user's sync frequency locally.
func SaveSyncFrequency(frequency int) error {
	dbInstance := app.GetDBInstance()

	updateErr := dbInstance.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("Preferences")).Put([]byte("frequency"), []byte(fmt.Sprint(frequency)))
		return err
	})

	return updateErr
}

// SaveDebugMode saves the user's chosen debug mode locally.
func SaveDebugMode(logLevel string) error {
	dbInstance := app.GetDBInstance()

	updateErr := dbInstance.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("Preferences")).Put([]byte("logLevel"), []byte(logLevel))
		return err
	})

	return updateErr
}
