package intTelegram

import (
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	"github.com/beebeeoii/lominus/internal/lominus"
)

func GetTelegramInfoPath() string {
	return filepath.Join(appDir.GetBaseDir(), lominus.TELEGRAM_FILE_NAME)
}
