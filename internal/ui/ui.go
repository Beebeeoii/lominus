// Package ui provides primitives that initialises the UI.
package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	intTelegram "github.com/beebeeoii/lominus/internal/app/integrations/telegram"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	"github.com/beebeeoii/lominus/internal/cron"
	"github.com/beebeeoii/lominus/internal/file"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/lominus"
	"github.com/beebeeoii/lominus/internal/notifications"
	"github.com/beebeeoii/lominus/pkg/integrations/telegram"
)

var mainApp fyne.App
var w fyne.Window

// Init builds and initialises the UI.
func Init() error {
	mainApp = app.NewWithID(lominus.APP_NAME)
	mainApp.SetIcon(resourceAppIconPng)

	go func() {
		for {
			notification := <-notifications.NotificationChannel
			mainApp.SendNotification(fyne.NewNotification(notification.Title, notification.Content))
		}
	}()

	w = mainApp.NewWindow(fmt.Sprintf("%s v%s", lominus.APP_NAME, lominus.APP_VERSION))

	if desk, ok := mainApp.(desktop.App); ok {
		m := BuildSystemTray()
		desk.SetSystemTrayMenu(m)
	}

	credentialsTab, credentialsUiErr := getCredentialsTab(w)
	if credentialsUiErr != nil {
		return credentialsUiErr
	}

	preferencesTab, preferencesErr := getPreferencesTab(w)
	if preferencesErr != nil {
		return preferencesErr
	}

	integrationsTab, integrationsErr := getIntegrationsTab(w)
	if integrationsErr != nil {
		return integrationsErr
	}

	tabsContainer := container.NewAppTabs(credentialsTab, preferencesTab, integrationsTab)
	content := container.NewVBox(
		tabsContainer,
		layout.NewSpacer(),
		getSyncButton(w),
		getQuitButton(),
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

// getIntegrationsTab builds the integrations tab in the main UI.
func getIntegrationsTab(parentWindow fyne.Window) (*container.TabItem, error) {
	logs.Logger.Debugln("integrations tab loaded")
	tab := container.NewTabItem("Integrations", container.NewVBox())

	label := widget.NewLabelWithStyle("Telegram", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0})
	subLabel := widget.NewRichTextFromMarkdown("Lominus can be linked to your Telegram bot to notify you when new grades are released.")
	subLabel.Wrapping = fyne.TextWrapWord

	botApiEntry := widget.NewPasswordEntry()
	botApiEntry.SetPlaceHolder("Your bot's API token")
	userIdEntry := widget.NewEntry()
	userIdEntry.SetPlaceHolder("Your account's ID")

	telegramInfoPath, getTelegramInfoPathErr := intTelegram.GetTelegramInfoPath()
	if getTelegramInfoPathErr != nil {
		return tab, getTelegramInfoPathErr
	}

	if file.Exists(telegramInfoPath) {
		telegramInfo, err := telegram.LoadTelegramData(telegramInfoPath)
		if err != nil {
			return tab, err
		}

		botApiEntry.SetText(telegramInfo.BotApi)
		userIdEntry.SetText(telegramInfo.UserId)
	}

	telegramForm := widget.NewForm(widget.NewFormItem("Bot API Token", botApiEntry), widget.NewFormItem("User ID", userIdEntry))

	saveButtonText := "Save Telegram Info"
	if botApiEntry.Text != "" && userIdEntry.Text != "" {
		saveButtonText = "Update Telegram Info"
	}

	saveButton := widget.NewButton(saveButtonText, func() {
		botApi := botApiEntry.Text
		userId := userIdEntry.Text

		status := widget.NewLabel("Please wait while we send you a test message...")
		progressBar := widget.NewProgressBarInfinite()

		mainDialog := dialog.NewCustom(lominus.APP_NAME, "Cancel", container.NewVBox(status, progressBar), parentWindow)
		mainDialog.Show()

		logs.Logger.Debugln("sending telegram test message")
		err := telegram.SendMessage(botApi, userId, "Thank you for using Lominus! You have succesfully integrated Telegram with Lominus!\n\nBy integrating Telegram with Lominus, you will be notified of the following whenever Lominus polls for new update based on the intervals set:\n💥 new grades releases\n💥 new announcements (TBC)")
		mainDialog.Hide()
		if err != nil {
			errMessage := fmt.Sprintf("%s: %s", err.Error()[:13], err.Error()[strings.Index(err.Error(), "description")+14:len(err.Error())-2])
			logs.Logger.Debugln("telegram test message failed to send")
			dialog.NewInformation(lominus.APP_NAME, errMessage, parentWindow).Show()
		} else {
			telegram.SaveTelegramData(telegramInfoPath, telegram.TelegramInfo{BotApi: botApi, UserId: userId})
			logs.Logger.Debugln("telegram test message sent successfully")
			dialog.NewInformation(lominus.APP_NAME, "Test message sent!\nTelegram info saved successfully.", parentWindow).Show()
		}
	})

	tab.Content = container.NewVBox(
		label,
		widget.NewSeparator(),
		subLabel,
		telegramForm,
		saveButton,
	)

	return tab, nil
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
	return widget.NewButton("Sync Now", func() {
		preferences := getPreferences()
		if preferences.Directory == "" {
			dialog.NewInformation(lominus.APP_NAME, "Please set the directory to store your Luminus files", parentWindow).Show()
			return
		}
		if preferences.Frequency == -1 {
			dialog.NewInformation(lominus.APP_NAME, "Sync is currently disabled. Please choose a sync frequency to sync now.", parentWindow).Show()
			return
		}
		cron.Rerun(getPreferences().Frequency)
	})
}

// getQuitButton builds the quit button in the main UI.
func getQuitButton() *widget.Button {
	return widget.NewButton("Quit Lominus", func() {
		logs.Logger.Infoln("lominus quit")
		mainApp.Quit()
	})
}
