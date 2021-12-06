package main

import (
	"log"

	"github.com/beebeeoii/lominus/internal/app"
	appLock "github.com/beebeeoii/lominus/internal/app/lock"
	"github.com/beebeeoii/lominus/internal/ui"
	"github.com/juju/fslock"
)

func main() {
	appInitErr := app.Init()
	if appInitErr != nil {
		log.Fatalln(appInitErr)
	}

	lock := fslock.New(appLock.GetLockPath())
	lockErr := lock.TryLock()

	if lockErr != nil {
		log.Fatalln(lockErr)
	}
	defer lock.Unlock()

	uiInitErr := ui.Init()
	if uiInitErr != nil {
		log.Fatalln(uiInitErr)
	}
}
