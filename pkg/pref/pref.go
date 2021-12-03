package pref

import "github.com/beebeeoii/lominus/internal/file"

type Preferences struct {
	Directory string
	Frequency int
}

const PREFERENCES_FILE_NAME = "pref.gob"

func SavePreferences(credentials Preferences) error {
	return file.EncodeStructToFile(PREFERENCES_FILE_NAME, credentials)
}

func LoadPreferences() (Preferences, error) {
	preferences := Preferences{}
	if !file.Exists(PREFERENCES_FILE_NAME) {
		return preferences, &file.FileNotFoundError{FileName: PREFERENCES_FILE_NAME}
	}
	err := file.DecodeStructFromFile(PREFERENCES_FILE_NAME, &preferences)

	return preferences, err
}
