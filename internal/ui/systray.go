package ui

import (
	"github.com/getlantern/systray"

	lominus "github.com/beebeeoii/lominus/internal/lominus"
)

func onReady() {
	systray.SetIcon(resourceAppIconIco.Content())
	systray.SetTitle(lominus.APP_NAME)
	systray.SetTooltip(lominus.APP_NAME)
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
}

func onExit() {
	mainApp.Quit()
}
