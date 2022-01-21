// Package appPref provides path retrievers for Lominus preferences files.
package appPref

import (
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	"github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/internal/lominus"
)

const PREFERENCES_FILE_NAME = lominus.PREFERENCES_FILE_NAME

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

	preferencesPath = filepath.Join(baseDir, lominus.PREFERENCES_FILE_NAME)

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
