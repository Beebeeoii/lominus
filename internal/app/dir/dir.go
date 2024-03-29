// Package appDir provides directory generators for Lominus config files.
package appDir

import (
	"os"
	"path/filepath"

	appConstants "github.com/beebeeoii/lominus/internal/constants"
)

// GetBaseDir returns the directory where config files for Lominus will be stored in.
// It uses os.UserConfigDir() under the hood and appends lominus.APP_NAME to it.
// On Unix systems, os.UserConfigDir() returns $XDG_CONFIG_HOME as specified by https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
// if non-empty, else $HOME/.config.
// On Darwin, it returns $HOME/Library/Application Support.
// On Windows, it returns %AppData%. On Plan 9, it returns $home/lib.
// If the location cannot be determined (for example, $HOME is not defined), then it will return an error.
func GetBaseDir() (string, error) {
	var baseDir string

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return baseDir, err
	}

	baseDir = filepath.Join(userConfigDir, appConstants.APP_NAME)

	return baseDir, nil
}
