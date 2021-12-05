package appPref

import (
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	lominus "github.com/beebeeoii/lominus/internal/lominus"
)

type Preferences struct {
	Directory string
	Frequency int
}

func GetPreferencesPath() string {
	return filepath.Join(appDir.GetBaseDir(), lominus.PREFERENCES_FILE_NAME)
}
