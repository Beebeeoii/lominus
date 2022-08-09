// Package updater provides functions to allow for app to self update
package updater

import (
	"fmt"
	"os"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	appConstants "github.com/beebeeoii/lominus/internal/constants"
	logs "github.com/beebeeoii/lominus/internal/log"
	lominus "github.com/beebeeoii/lominus/internal/lominus"
	"github.com/creativeprojects/go-selfupdate"
)

const (
	GITHUB_REPO = "Beebeeoii/lominus"
	VERSION     = lominus.APP_VERSION
)

func DoSelfUpdate(parentWindow fyne.Window) {
	selfupdate.SetLogger(logs.Logger)

	latest, found, err := selfupdate.DetectLatest(GITHUB_REPO)
	if err != nil {
		logs.Logger.Fatalln(fmt.Sprintf("Error occurred while detecting version: %v", err))
	}
	if !found {
		logs.Logger.Fatalln(fmt.Sprintf("Latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH))
	}

	if latest.LessOrEqual(VERSION) {
		logs.Logger.Fatalln(fmt.Sprintf("Current version (%s) is the latest", VERSION))
	}

	exe, err := os.Executable()
	if err != nil {
		logs.Logger.Fatalln("Could not locate executable path")
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, latest.AssetName, exe); err != nil {
		logs.Logger.Fatalln(fmt.Sprintf("Error occurred while updating binary: %v", err))
	}

	updateMessage := fmt.Sprintf(appConstants.UPDATE_DIALOG_MESSAGE, latest.Version())

	// TODO: popup a GUI stating that it is updated
	logs.Logger.Infoln(updateMessage)

	dialog.NewInformation(
		lominus.APP_NAME,
		updateMessage,
		parentWindow,
	).Show()
}
