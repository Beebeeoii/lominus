package appLock

import (
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	"github.com/beebeeoii/lominus/internal/lominus"
)

func GetLockPath() string {
	return filepath.Join(appDir.GetBaseDir(), lominus.LOCK_FILE_NAME)
}
