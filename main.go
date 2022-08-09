// Package main is where Lominus starts from.
package main

import (
	"log"
	"os"
	"runtime"

	"github.com/beebeeoii/lominus/internal/app"
	appLock "github.com/beebeeoii/lominus/internal/app/lock"
	"github.com/beebeeoii/lominus/internal/cron"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/notifications"
	"github.com/beebeeoii/lominus/internal/ui"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/juju/fslock"
)

func doSelfUpdate() {
	version := "1.2.4"
	selfupdate.SetLogger(log.New(os.Stderr, "", log.LstdFlags))

	latest, found, err := selfupdate.DetectLatest("Beebeeoii/lominus")
	if err != nil {
		println("error occurred while detecting version: %v", err)
	}
	if !found {
		println("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}

	if latest.LessOrEqual(version) {
		println("Current version (%s) is the latest", version)
	}

	exe, err := os.Executable()
	if err != nil {
		println("could not locate executable path")
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, latest.AssetName, exe); err != nil {
		println("error occurred while updating binary: %v", err)
	}
	log.Printf("Successfully updated to version %s", latest.Version())
}

// Main is the starting point of where magic begins.
func main() {
	appInitErr := app.Init()
	if appInitErr != nil {
		logs.Logger.Fatalln(appInitErr)
	}
	logs.Logger.Infoln("app initialised")

	lockPath, getLockPathErr := appLock.GetLockPath()
	if getLockPathErr != nil {
		logs.Logger.Fatalln(getLockPathErr)
	}

	lock := fslock.New(lockPath)
	lockErr := lock.TryLock()

	if lockErr != nil {
		logs.Logger.Fatalln(lockErr)
	}
	defer lock.Unlock()
	logs.Logger.Infoln("lock initialised")

	notifications.Init()
	logs.Logger.Infoln("notifications initialised")

	cronInitErr := cron.Init()
	if cronInitErr != nil {
		logs.Logger.Fatalln(cronInitErr)
	}
	logs.Logger.Infoln("cron initialised")

	doSelfUpdate()

	uiInitErr := ui.Init()

	if uiInitErr != nil {
		logs.Logger.Fatalln(uiInitErr)
	}
}
