// Package ui provides primitives that initialises the UI.
package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
	appInt "github.com/beebeeoii/lominus/internal/app/integrations/telegram"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	"github.com/beebeeoii/lominus/internal/cron"
	"github.com/beebeeoii/lominus/internal/notifications"
)

var mainApp fyne.App
var w fyne.Window

// Init builds and initialises the UI.
func Init() error {
	mainApp = app.NewWithID(appConstants.APP_NAME)
	mainApp.SetIcon(resourceAppIconPng)

	canvasCredentials, credErr := appAuth.GetCanvasCredentials()
	if credErr != nil {
		return credErr
	}

	pref, prefErr := appPref.GetPreferences()
	if prefErr != nil {
		return prefErr
	}

	telegramIds, tIdsErr := appInt.GetTelegramIds()
	if tIdsErr != nil {
		return tIdsErr
	}

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
		CanvasApiToken: canvasCredentials.CanvasApiToken,
	}, w)
	if credentialsUiErr != nil {
		return credentialsUiErr
	}

	preferencesTab, preferencesErr := getPreferencesTab(PreferencesData{
		Directory: pref.Directory,
		Frequency: pref.Frequency,
		LogLevel:  pref.LogLevel,
	}, w)
	if preferencesErr != nil {
		return preferencesErr
	}

	integrationsTab, integrationsErr := getIntegrationsTab(IntegrationData{
		TelegramUserId: telegramIds.UserId,
		TelegramBotId:  telegramIds.BotId,
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

// getSyncButton builds the sync button in the main UI.
func getSyncButton(parentWindow fyne.Window) *widget.Button {
	return widget.NewButton(appConstants.SYNC_TEXT, func() {
		pref, prefErr := appPref.GetPreferences()
		if prefErr != nil {
			return
		}

		if pref.Directory == "" {
			dialog.NewInformation(
				appConstants.APP_NAME,
				appConstants.NO_FOLDER_DIRECTORY_SELECTED,
				parentWindow,
			).Show()

			return
		}

		if pref.Frequency == -1 {
			dialog.NewInformation(
				appConstants.APP_NAME,
				appConstants.NO_FREQUENCY_SELECTED,
				parentWindow,
			).Show()

			return
		}

		cron.Rerun(pref.Directory, pref.Frequency)
	})
}
