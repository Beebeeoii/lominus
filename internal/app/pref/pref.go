// Package appPref provides path retrievers for Lominus preferences files.
package appPref

import (
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	"github.com/beebeeoii/lominus/internal/lominus"
)

// Preferences struct describes the data being stored in the user's preferences file.
type Preferences struct {
	Directory string
	Frequency int
}

// GetJwtPath returns the file path to user's preferences.
func GetPreferencesPath() string {
	return filepath.Join(appDir.GetBaseDir(), lominus.PREFERENCES_FILE_NAME)
}
