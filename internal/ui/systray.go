// Package ui provides primitives that initialises the UI.
package ui

import (
	"fmt"

	"github.com/getlantern/systray"

	"github.com/beebeeoii/lominus/internal/cron"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/lominus"
)

var lastRanItem *systray.MenuItem

// onReady builds and initialises the system tray UI.
func onReady() {
	systray.SetIcon(resourceAppIconIco.Content())
	systray.SetTitle(lominus.APP_NAME)
	systray.SetTooltip(lominus.APP_NAME)
	lastRanItem = systray.AddMenuItem("Last sync: Nil", "Shows when Lominus last checked for updates")
	lastRanItem.Disable()
	openButton := systray.AddMenuItem("Open", "Open Lominus")
	systray.AddSeparator()
	quitButton := systray.AddMenuItem("Quit", "Quit Lominus")

	go func() {
		for {
			select {
			case <-openButton.ClickedCh:
				w.Show()
			case <-quitButton.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()

	go func() {
		for {
			lastRanItem.SetTitle(fmt.Sprintf("Last sync: %s", <-cron.LastRanChannel))
		}
	}()
}

// onExit describes the actions taken when user quits the system tray.
func onExit() {
	logs.InfoLogger.Println("lominus quit")
	mainApp.Quit()
}
