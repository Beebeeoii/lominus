// Package appAuth provides path retrievers for Lominus auth files.
package appAuth

import (
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	"github.com/beebeeoii/lominus/internal/lominus"
)

// GetJwtPath returns the file path to user's JWT data.
func GetJwtPath() string {
	return filepath.Join(appDir.GetBaseDir(), lominus.JWT_DATA_FILE_NAME)
}

// GetJwtPath returns the file path to user's credentials.
func GetCredentialsPath() string {
	return filepath.Join(appDir.GetBaseDir(), lominus.CREDENTIALS_FILE_NAME)
}
