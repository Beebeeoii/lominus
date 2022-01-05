// Package pref provides primitives that interacts with Preference files.
package pref

import (
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	"github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/internal/lominus"
)

const PREFERENCES_FILE_NAME = lominus.PREFERENCES_FILE_NAME

var Preferences = appPref.Preferences{}

// SavePreferences saves the user's preferences data onto local storage.
func SavePreferences(filePath string, preferences appPref.Preferences) error {
	return file.EncodeStructToFile(filePath, preferences)
}

// LoadPreferences loads the user's preferences data from local storage.
func LoadPreferences(filePath string) (appPref.Preferences, error) {
	preferences := appPref.Preferences{}
	if !file.Exists(filePath) {
		return preferences, &file.FileNotFoundError{FileName: filePath}
	}
	err := file.DecodeStructFromFile(filePath, &preferences)

	return preferences, err
}
