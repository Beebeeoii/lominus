// Package ui provides primitives that initialises the UI.
package ui

import (
	"fyne.io/fyne/v2"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	"github.com/beebeeoii/lominus/internal/cron"
	"github.com/beebeeoii/lominus/internal/notifications"
)

// BuildSystemTray creates the system tray icon and its menu options, used to be initialised
// when Lominus starts.
func BuildSystemTray() *fyne.Menu {
	return fyne.NewMenu(appConstants.APP_NAME,
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
