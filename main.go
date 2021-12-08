package main

import (
	"github.com/beebeeoii/lominus/internal/app"
	appLock "github.com/beebeeoii/lominus/internal/app/lock"
	"github.com/beebeeoii/lominus/internal/cron"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/notifications"
	"github.com/beebeeoii/lominus/internal/ui"
	"github.com/juju/fslock"
)

func main() {
	appInitErr := app.Init()
	if appInitErr != nil {
		logs.ErrorLogger.Fatalln(appInitErr)
	}
	logs.InfoLogger.Println("app initialised")

	lock := fslock.New(appLock.GetLockPath())
	lockErr := lock.TryLock()

	if lockErr != nil {
		logs.ErrorLogger.Fatalln(lockErr)
	}
	defer lock.Unlock()

	notifications.Init()
	logs.InfoLogger.Println("notifications initialised")

	cronInitErr := cron.Init()
	if cronInitErr != nil {
		logs.ErrorLogger.Fatalln(cronInitErr)
	}
	logs.InfoLogger.Println("cron initialised")

	uiInitErr := ui.Init()
	if uiInitErr != nil {
		logs.ErrorLogger.Fatalln(uiInitErr)
	}
}
