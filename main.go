package main

import (
	"log"

	"github.com/beebeeoii/lominus/internal/app"
	"github.com/beebeeoii/lominus/internal/ui"
)

func main() {
	appInitErr := app.Init()
	if appInitErr != nil {
		log.Fatalln(appInitErr)
	}

	uiInitErr := ui.Init()
	if uiInitErr != nil {
		log.Fatalln(uiInitErr)
	}
}
