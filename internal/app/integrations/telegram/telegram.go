// Package intTelegram provides path retrievers for Lominus Telegram integration files.
package intTelegram

import (
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	"github.com/beebeeoii/lominus/internal/lominus"
)

// GetTelegramInfoPath returns the file path to user's telegram config file.
func GetTelegramInfoPath() string {
	return filepath.Join(appDir.GetBaseDir(), lominus.TELEGRAM_FILE_NAME)
}
