// Package ui provides primitives that initialises the UI.
package ui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	"github.com/beebeeoii/lominus/internal/cron"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/notifications"
	"github.com/boltdb/bolt"
)

var mainApp fyne.App
var w fyne.Window

// Init builds and initialises the UI.
func Init() error {
	mainApp = app.NewWithID(appConstants.APP_NAME)
	mainApp.SetIcon(resourceAppIconPng)

	var canvasToken, directory, logLevel, telegramUserId, telegramBotId string
	var frequency int

	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		return retrieveBaseDirErr
	}

	dbFName := filepath.Join(baseDir, appConstants.DATABASE_FILE_NAME)
	db, dbErr := bolt.Open(dbFName, 0600, &bolt.Options{Timeout: 3 * time.Second})

	if dbErr != nil {
		return dbErr
	}

	tx, _ := db.Begin(false)
	authBucket := tx.Bucket([]byte("Auth"))
	canvasToken = string(authBucket.Get([]byte("canvasToken")))

	prefBucket := tx.Bucket([]byte("Preferences"))
	directory = string(prefBucket.Get([]byte("directory")))
	frequency, _ = strconv.Atoi(string(prefBucket.Get([]byte("frequency"))))
	logLevel = string(prefBucket.Get([]byte("logLevel")))

	intBucket := tx.Bucket([]byte("Integrations"))
	telegramUserId = string(intBucket.Get([]byte("telegramUserId")))
	telegramBotId = string(intBucket.Get([]byte("telegramBotId")))
	tx.Rollback()

	db.Close()

	go func() {
		for {
			notification := <-notifications.NotificationChannel
			mainApp.SendNotification(fyne.NewNotification(notification.Title, notification.Content))
		}
	}()

	w = mainApp.NewWindow(fmt.Sprintf("%s v%s", appConstants.APP_NAME, appConstants.APP_VERSION))

	if desk, ok := mainApp.(desktop.App); ok {
		m := BuildSystemTray()
		desk.SetSystemTrayMenu(m)
	}

	credentialsTab, credentialsUiErr := getCredentialsTab(CredentialsData{
		CanvasApiToken: canvasToken,
	}, w)
	if credentialsUiErr != nil {
		return credentialsUiErr
	}

	preferencesTab, preferencesErr := getPreferencesTab(PreferencesData{
		Directory: directory,
		Frequency: frequency,
		LogLevel:  logLevel,
	}, w)
	if preferencesErr != nil {
		return preferencesErr
	}

	integrationsTab, integrationsErr := getIntegrationsTab(IntegrationData{
		TelegramUserId: telegramUserId,
		TelegramBotId:  telegramBotId,
	}, w)
	if integrationsErr != nil {
		return integrationsErr
	}

	tabsContainer := container.NewAppTabs(credentialsTab, integrationsTab, preferencesTab)
	content := container.NewVBox(
		tabsContainer,
		layout.NewSpacer(),
		getSyncButton(w),
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(600, 600))
	w.SetPadded(true)
	w.SetFixedSize(true)
	w.SetMaster()
	w.SetCloseIntercept(func() {
		w.Hide()
	})
	mainApp.Lifecycle().SetOnEnteredForeground(func() {
		w.Show()
	})
	w.ShowAndRun()
	return nil
}

// getPreferences is a util function that retrieves the user's preferences.
func getPreferences() appPref.Preferences {
	preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
	if getPreferencesPathErr != nil {
		logs.Logger.Fatalln(getPreferencesPathErr)
	}

	preference, err := appPref.LoadPreferences(preferencesPath)
	if err != nil {
		logs.Logger.Fatalln(err)
	}

	return preference
}

// getSyncButton builds the sync button in the main UI.
func getSyncButton(parentWindow fyne.Window) *widget.Button {
	return widget.NewButton(appConstants.SYNC_TEXT, func() {
		preferences := getPreferences()
		if preferences.Directory == "" {
			dialog.NewInformation(
				appConstants.APP_NAME,
				appConstants.NO_FOLDER_DIRECTORY_SELECTED,
				parentWindow,
			).Show()
			return
		}

		if preferences.Frequency == -1 {
			dialog.NewInformation(
				appConstants.APP_NAME,
				appConstants.NO_FREQUENCY_SELECTED,
				parentWindow,
			).Show()
			return
		}
		cron.Rerun(getPreferences().Frequency)
	})
}
