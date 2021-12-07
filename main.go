package main

import (
	"github.com/beebeeoii/lominus/internal/app"
	appLock "github.com/beebeeoii/lominus/internal/app/lock"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/ui"
	"github.com/juju/fslock"
)

func main() {
	appInitErr := app.Init()
	if appInitErr != nil {
		logs.ErrorLogger.Fatalln(appInitErr)
	}

	lock := fslock.New(appLock.GetLockPath())
	lockErr := lock.TryLock()

	if lockErr != nil {
		logs.ErrorLogger.Fatalln(lockErr)
	}
	defer lock.Unlock()

	uiInitErr := ui.Init()
	if uiInitErr != nil {
		logs.ErrorLogger.Fatalln(uiInitErr)
	}
}
