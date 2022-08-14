// Package appLock provides path retrievers for Lominus lock file.
package appLock

import (
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
)

// GetLockPath returns the file path to Lominus lock file.
func GetLockPath() (string, error) {
	var lockPath string

	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		return lockPath, retrieveBaseDirErr
	}

	lockPath = filepath.Join(baseDir, appConstants.LOCK_FILE_NAME)

	return lockPath, nil
}
