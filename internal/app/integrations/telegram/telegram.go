// Package intTelegram provides path retrievers for Lominus Telegram integration files.
package intTelegram

import (
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	"github.com/beebeeoii/lominus/internal/lominus"
)

// GetTelegramInfoPath returns the file path to user's telegram config file.
func GetTelegramInfoPath() (string, error) {
	var telegramInfoPath string

	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		return telegramInfoPath, retrieveBaseDirErr
	}

	telegramInfoPath = filepath.Join(baseDir, lominus.TELEGRAM_FILE_NAME)

	return telegramInfoPath, nil
}
