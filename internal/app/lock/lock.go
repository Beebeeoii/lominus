// Package appLock provides path retrievers for Lominus lock file.
package appLock

import (
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	"github.com/beebeeoii/lominus/internal/lominus"
)

// GetLockPath returns the file path to Lominus lock file.
func GetLockPath() string {
	return filepath.Join(appDir.GetBaseDir(), lominus.LOCK_FILE_NAME)
}
