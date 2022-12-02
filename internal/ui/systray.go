// Package ui provides primitives that initialises the UI.
package ui

import (
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	"github.com/beebeeoii/lominus/internal/cron"
	"github.com/beebeeoii/lominus/internal/notifications"
	"github.com/boltdb/bolt"
)

// BuildSystemTray creates the system tray icon and its menu options, used to be initialised
// when Lominus starts.
func BuildSystemTray() *fyne.Menu {
	return fyne.NewMenu(appConstants.APP_NAME,
		fyne.NewMenuItem("Sync Now", func() {
			baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
			if retrieveBaseDirErr != nil {
				return
			}

			dbFName := filepath.Join(baseDir, appConstants.DATABASE_FILE_NAME)
			db, dbErr := bolt.Open(dbFName, 0600, &bolt.Options{ReadOnly: true})

			if dbErr != nil {
				return
			}

			tx, _ := db.Begin(false)
			prefBucket := tx.Bucket([]byte("Preferences"))
			directory := string(prefBucket.Get([]byte("directory")))
			frequency, _ := strconv.Atoi(string(prefBucket.Get([]byte("frequency"))))

			defer db.Close()

			if directory == "" {
				notifications.NotificationChannel <- notifications.Notification{Title: "Unable to sync", Content: "Please set the directory to store your Luminus files"}
				return
			}

			if frequency == -1 {
				notifications.NotificationChannel <- notifications.Notification{Title: "Unable to sync", Content: "Please choose a sync frequency to sync now."}
				return
			}

			cron.Rerun(directory, frequency)
		}),
		fyne.NewMenuItem("Open Lominus", func() {
			w.Show()
		}),
	)
}
