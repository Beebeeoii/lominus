package appAuth

import (
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	"github.com/beebeeoii/lominus/internal/lominus"
)

func GetJwtPath() string {
	return filepath.Join(appDir.GetBaseDir(), lominus.JWT_DATA_FILE_NAME)
}

func GetCredentialsPath() string {
	return filepath.Join(appDir.GetBaseDir(), lominus.CREDENTIALS_FILE_NAME)
}
