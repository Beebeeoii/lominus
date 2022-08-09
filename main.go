// Package main is where Lominus starts from.
package main

import (
	"github.com/beebeeoii/lominus/internal/app"
	appLock "github.com/beebeeoii/lominus/internal/app/lock"
	"github.com/beebeeoii/lominus/internal/cron"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/notifications"
	"github.com/beebeeoii/lominus/internal/ui"

	updater "github.com/beebeeoii/lominus/internal/app/updater"

	"github.com/juju/fslock"
)

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

	updater.DoSelfUpdate()

	uiInitErr := ui.Init()

	if uiInitErr != nil {
		logs.Logger.Fatalln(uiInitErr)
	}
}
