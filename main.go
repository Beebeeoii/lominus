package main

import (
	"log"

	"github.com/beebeeoii/lominus/internal/app"
	appLock "github.com/beebeeoii/lominus/internal/app/lock"
	"github.com/beebeeoii/lominus/internal/ui"

	"github.com/juju/fslock"
)

func main() {
	lock := fslock.New(appLock.GetLockPath())
	err := lock.TryLock()

	if err != nil {
		log.Fatalln(err)
	}
	defer lock.Unlock()

	appInitErr := app.Init()
	if appInitErr != nil {
		log.Fatalln(appInitErr)
	}

	uiInitErr := ui.Init()
	if uiInitErr != nil {
		log.Fatalln(uiInitErr)
	}
}
