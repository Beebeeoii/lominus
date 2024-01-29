// Package ui provides primitives that initialises the UI.
package ui

import (
	"fyne.io/fyne/v2"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	"github.com/beebeeoii/lominus/internal/cron"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/notifications"
)

// BuildSystemTray creates the system tray icon and its menu options, used to be initialised
// when Lominus starts.
func BuildSystemTray() *fyne.Menu {
	return fyne.NewMenu(appConstants.APP_NAME,
		fyne.NewMenuItem("Sync Now", func() {
			pref, err := appPref.GetPreferences()
			if err != nil {
				logs.Logger.Errorln("[systray]: Sync Now - Unable to get preferences")
			}

			if pref.Directory == "" {
				notifications.NotificationChannel <- notifications.Notification{Title: "Unable to sync", Content: "Please set the directory to store your files"}
				return
			}

			if pref.Frequency == -1 {
				notifications.NotificationChannel <- notifications.Notification{Title: "Unable to sync", Content: "Please choose a sync frequency to sync now."}
				return
			}

			cron.Rerun(pref.Directory, pref.Frequency)
		}),
		fyne.NewMenuItem("Open Lominus", func() {
			w.Show()
		}),
	)
}
