// Package ui provides primitives that initialises the UI.
package ui

import (
	"fyne.io/fyne/v2"
	"github.com/beebeeoii/lominus/internal/cron"
	"github.com/beebeeoii/lominus/internal/lominus"
	"github.com/beebeeoii/lominus/internal/notifications"
)

// TODO Documentation
func BuildSystemTray() *fyne.Menu {
	return fyne.NewMenu(lominus.APP_NAME,
		fyne.NewMenuItem("Sync Now", func() {
			preferences := getPreferences()
			if preferences.Directory == "" {
				notifications.NotificationChannel <- notifications.Notification{Title: "Unable to sync", Content: "Please set the directory to store your Luminus files"}
				return
			}
			if preferences.Frequency == -1 {
				notifications.NotificationChannel <- notifications.Notification{Title: "Unable to sync", Content: "Please choose a sync frequency to sync now."}
				return
			}
			cron.Rerun(getPreferences().Frequency)
		}),
		fyne.NewMenuItem("Open Lominus", func() {
			w.Show()
		}),
	)
}
