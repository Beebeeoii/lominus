// Package app provides primitives to initialise crucial files for Lominus.
package app

import (
	"os"
	"runtime"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	"github.com/beebeeoii/lominus/internal/file"
	logs "github.com/beebeeoii/lominus/internal/log"
)

// Init initialises and ensures log and preference files that Lominus requires are available.
// Directory in Preferences defaults to empty string ("").
// Frequency in Preferences defaults to -1.
func Init() error {
	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		return retrieveBaseDirErr
	}

	if !file.Exists(baseDir) {
		os.Mkdir(baseDir, os.ModePerm)
	}

	logInitErr := logs.Init()
	if logInitErr != nil {
		return logInitErr
	}

	preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
	if getPreferencesPathErr != nil {
		return getPreferencesPathErr
	}

	if !file.Exists(preferencesPath) {
		preferences := appPref.Preferences{
			Directory: "",
			Frequency: -1,
			LogLevel:  "info",
		}

		savePrefErr := appPref.SavePreferences(preferencesPath, preferences)
		if savePrefErr != nil {
			return savePrefErr
		}
	}

	return nil
}

// GetOs returns user's running program's operating system target:
// one of darwin, freebsd, linux, and so on.
// To view possible combinations of GOOS and GOARCH, run "go tool dist list".
func GetOs() string {
	return runtime.GOOS
}
