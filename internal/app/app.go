// Package app provides primitives to initialise crucial files for Lominus.
package app

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
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

	preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
	if getPreferencesPathErr != nil {
		return getPreferencesPathErr
	}

	if !file.Exists(preferencesPath) {
		preferences := appPref.Preferences{
			Directory:  "",
			Frequency:  -1,
			LogLevel:   "info",
			AutoUpdate: false,
		}

		savePrefErr := appPref.SavePreferences(preferencesPath, preferences)
		if savePrefErr != nil {
			return savePrefErr
		}
	} else {
		preferences, getPreferencesErr := appPref.LoadPreferences(preferencesPath)
		if getPreferencesErr != nil {
			return getPreferencesErr
		}

		if preferences.LogLevel != "info" && preferences.LogLevel != "debug" {
			preferences.LogLevel = "info"
		}

		savePrefErr := appPref.SavePreferences(preferencesPath, preferences)
		if savePrefErr != nil {
			return savePrefErr
		}
	}

	logInitErr := logs.Init()
	if logInitErr != nil {
		return logInitErr
	}

	// TODO Consider moving this to its own module in the future.
	gradesPath := filepath.Join(baseDir, appConstants.GRADES_FILE_NAME)

	if !file.Exists(gradesPath) {
		gradeFileErr := file.EncodeStructToFile(gradesPath, time.Now())

		if gradeFileErr != nil {
			return gradeFileErr
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
