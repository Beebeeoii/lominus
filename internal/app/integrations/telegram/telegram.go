// Package intTelegram provides path retrievers for Lominus Telegram integration files.
package intTelegram

import (
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
)

// GetTelegramInfoPath returns the file path to user's telegram config file.
func GetTelegramInfoPath() (string, error) {
	var telegramInfoPath string

	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		return telegramInfoPath, retrieveBaseDirErr
	}

	telegramInfoPath = filepath.Join(baseDir, appConstants.TELEGRAM_FILE_NAME)

	return telegramInfoPath, nil
}
