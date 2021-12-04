package appLock

import (
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	lominus "github.com/beebeeoii/lominus/internal/lominus"
)

func GetLockPath() string {
	return filepath.Join(appDir.GetBaseDir(), lominus.LOCK_FILE_NAME)
}
