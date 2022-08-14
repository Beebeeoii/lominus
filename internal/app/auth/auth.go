// Package appAuth provides path retrievers for Lominus auth files.
package appAuth

import (
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
)

// GetJwtPath returns the file path to user's JWT data.
func GetTokensPath() (string, error) {
	var jwtPath string

	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		return jwtPath, retrieveBaseDirErr
	}

	jwtPath = filepath.Join(baseDir, appConstants.TOKENS_FILE_NAME)

	return jwtPath, nil
}

// GetJwtPath returns the file path to user's credentials.
func GetCredentialsPath() (string, error) {
	var credentialsPath string

	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		return credentialsPath, retrieveBaseDirErr
	}

	credentialsPath = filepath.Join(baseDir, appConstants.CREDENTIALS_FILE_NAME)

	return credentialsPath, nil
}
