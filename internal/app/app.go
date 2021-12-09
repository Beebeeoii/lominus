package app

import (
	"os"
	"path/filepath"
	"runtime"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	"github.com/beebeeoii/lominus/internal/file"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/lominus"
	"github.com/beebeeoii/lominus/pkg/pref"
)

func Init() error {
	baseDir := appDir.GetBaseDir()

	if !file.Exists(baseDir) {
		os.Mkdir(baseDir, os.ModePerm)
	}

	logInitErr := logs.Init()
	if logInitErr != nil {
		return logInitErr
	}

	prefDir := filepath.Join(baseDir, lominus.PREFERENCES_FILE_NAME)
	if !file.Exists(prefDir) {
		preferences := appPref.Preferences{Directory: "", Frequency: -1}

		savePrefErr := pref.SavePreferences(prefDir, preferences)
		if savePrefErr != nil {
			return savePrefErr
		}
		logs.InfoLogger.Println("pref.go created")
	}

	return nil
}

func GetOs() string {
	return runtime.GOOS
}
