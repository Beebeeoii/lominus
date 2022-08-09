// Package updater provides functions to allow for app to self update
package updater

import (
	"os"
	"runtime"

	logs "github.com/beebeeoii/lominus/internal/log"
	lominus "github.com/beebeeoii/lominus/internal/lominus"
	"github.com/creativeprojects/go-selfupdate"
)

const (
	GITHUB_REPO = "Beebeeoii/lominus"
	VERSION     = lominus.APP_VERSION
)

func DoSelfUpdate() {
	selfupdate.SetLogger(logs.Logger)

	latest, found, err := selfupdate.DetectLatest(GITHUB_REPO)
	if err != nil {
		logs.Logger.Fatalln("Error occurred while detecting version: %v", err)
	}
	if !found {
		logs.Logger.Fatalln("Latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}

	if latest.LessOrEqual(VERSION) {
		logs.Logger.Fatalln("Current version (%s) is the latest", VERSION)
	}

	exe, err := os.Executable()
	if err != nil {
		logs.Logger.Fatalln("Could not locate executable path")
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, latest.AssetName, exe); err != nil {
		logs.Logger.Fatalln("Error occurred while updating binary: %v", err)
	}

	// TODO: popup a GUI stating that it is updated
	logs.Logger.Infoln("Successfully updated to version %s. Restart app to see changes.", latest.Version())
}
