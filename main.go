// Package main is where Lominus starts from.
package main

import (
	"github.com/beebeeoii/lominus/internal/app"
	appLock "github.com/beebeeoii/lominus/internal/app/lock"
	"github.com/beebeeoii/lominus/internal/cron"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/notifications"
	"github.com/beebeeoii/lominus/internal/ui"
	"github.com/blang/semver"
	"github.com/juju/fslock"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

func doSelfUpdate() {
	v := semver.MustParse("1.2.4")
	selfupdate.EnableLog()

	// println(selfupdate.DetectVersion("Beebeeoii/lominus", "1.2.4"))
	// println(selfupdate.DetectLatest("Beebeeoii/lominus"))

	latest, err := selfupdate.UpdateSelf(v, "Beebeeoii/lominus")
	if err != nil {
		println("Binary update failed:", err)
		return
	}

	println(latest)
	println(latest.Version.String())

	if latest.Version.Equals(v) {
		// latest version is the same as current version. It means current binary is up to date.
		println("Current binary is the latest version 1.2.3")
	} else {
		println("Successfully updated to version", latest.Version.String())
		println("Release note:\n", latest.ReleaseNotes)
	}
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
