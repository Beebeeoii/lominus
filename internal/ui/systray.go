package ui

import (
	"fyne.io/fyne/v2"
	"github.com/getlantern/systray"

	lominus "github.com/beebeeoii/lominus/internal/lominus"
)

func onMinimise() {
	systray.Run(onReady, onExit)
}

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
				systray.Quit()
				fyne.CurrentApp().Run()
			case <-quitButton.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
	// mQuit.SetIcon(resourceAppIconPng.StaticContent)
}

func onExit() {
}
