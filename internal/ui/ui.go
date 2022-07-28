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

	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
	intTelegram "github.com/beebeeoii/lominus/internal/app/integrations/telegram"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	"github.com/beebeeoii/lominus/internal/cron"
	"github.com/beebeeoii/lominus/internal/file"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/lominus"
	"github.com/beebeeoii/lominus/internal/notifications"
	"github.com/beebeeoii/lominus/pkg/auth"
	"github.com/beebeeoii/lominus/pkg/integrations/telegram"

	fileDialog "github.com/sqweek/dialog"
)

const (
	FREQUENCY_DISABLED    = "Disabled"
	FREQUENCY_ONE_HOUR    = "1 hour"
	FREQUENCY_TWO_HOUR    = "2 hour"
	FREQUENCY_FOUR_HOUR   = "4 hour"
	FREQUENCY_SIX_HOUR    = "6 hour"
	FREQUENCY_TWELVE_HOUR = "12 hour"
)

var frequencyMap = map[int]string{
	1:  FREQUENCY_ONE_HOUR,
	2:  FREQUENCY_TWO_HOUR,
	4:  FREQUENCY_FOUR_HOUR,
	6:  FREQUENCY_SIX_HOUR,
	12: FREQUENCY_TWELVE_HOUR,
	-1: FREQUENCY_DISABLED,
}

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

// getCredentialsTab builds the credentials tab in the main UI.
func getCredentialsTab(parentWindow fyne.Window) (*container.TabItem, error) {
	logs.Logger.Debugln("credentials tab loaded")
	tab := container.NewTabItem("Login Info", container.NewVBox())

	label := widget.NewLabelWithStyle("Your Credentials", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0})
	subLabel := widget.NewRichTextFromMarkdown("Credentials are saved **locally**. It is used to login to [Luminus](https://luminus.nus.edu.sg) **only**.")
	subLabel.Wrapping = fyne.TextWrapWord

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Eg: nusstu\\e0123456")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")
	canvasTokenEntry := widget.NewPasswordEntry()
	canvasTokenEntry.SetPlaceHolder("Account > Settings > New access token > Generate Token")

	credentialsPath, getCredentialsPathErr := appAuth.GetCredentialsPath()
	if getCredentialsPathErr != nil {
		return tab, getCredentialsPathErr
	}

	tokensPath, getTokensPathErr := appAuth.GetTokensPath()
	if getTokensPathErr != nil {
		return tab, getTokensPathErr
	}

	if file.Exists(credentialsPath) {
		logs.Logger.Debugf("credentials exists - loading from %s", credentialsPath)
		credentials, err := auth.LoadCredentialsData(credentialsPath)
		if err != nil {
			return tab, err
		}

		usernameEntry.SetText(credentials.LuminusCredentials.Username)
		passwordEntry.SetText(credentials.LuminusCredentials.Password)
		canvasTokenEntry.SetText(credentials.CanvasCredentials.CanvasApiToken)
	}

	luminusCredentialsForm := widget.NewForm(widget.NewFormItem("Username", usernameEntry), widget.NewFormItem("Password", passwordEntry))
	canvasCredentialsForm := widget.NewForm(widget.NewFormItem("Canvas Token", canvasTokenEntry))

	saveButtonText := "Save Credentials"
	if usernameEntry.Text != "" && passwordEntry.Text != "" {
		saveButtonText = "Update Credentials"
	}

	luminusSaveButton := widget.NewButton(saveButtonText, func() {
		luminusCredentials := auth.LuminusCredentials{
			Username: usernameEntry.Text,
			Password: passwordEntry.Text,
		}

		status := widget.NewLabel("Please wait while we verify your credentials...")
		progressBar := widget.NewProgressBarInfinite()

		mainDialog := dialog.NewCustom(lominus.APP_NAME, "Cancel", container.NewVBox(status, progressBar), parentWindow)
		mainDialog.Show()

		logs.Logger.Debugln("verifying credentials")
		_, err := auth.RetrieveJwtToken(luminusCredentials, true)
		mainDialog.Hide()
		if err != nil {
			logs.Logger.Debugln("verfication failed")
			dialog.NewInformation(lominus.APP_NAME, "Verification failed. Please check your credentials.", parentWindow).Show()
		} else {
			logs.Logger.Debugln("verfication succesful - saving credentials")
			luminusCredentials.Save(credentialsPath)
			dialog.NewInformation(lominus.APP_NAME, "Verification successful.", parentWindow).Show()
		}
	})

	canvasSaveButton := widget.NewButton(saveButtonText, func() {
		canvasCredentials := auth.CanvasCredentials{
			CanvasApiToken: canvasTokenEntry.Text,
		}

		canvasTokens := auth.CanvasTokenData{
			CanvasApiToken: canvasTokenEntry.Text,
		}

		status := widget.NewLabel("Please wait while we verify your credentials...")
		progressBar := widget.NewProgressBarInfinite()

		mainDialog := dialog.NewCustom(lominus.APP_NAME, "Cancel", container.NewVBox(status, progressBar), parentWindow)
		mainDialog.Show()

		logs.Logger.Debugln("verifying credentials")
		err := canvasCredentials.Authenticate()
		mainDialog.Hide()
		if err != nil {
			logs.Logger.Debugln("verfication failed")
			dialog.NewInformation(lominus.APP_NAME, "Verification failed. Please check your credentials.", parentWindow).Show()
		} else {
			logs.Logger.Debugln("verfication succesful - saving credentials")
			canvasCredentials.Save(credentialsPath)
			canvasTokens.Save(tokensPath)
			dialog.NewInformation(lominus.APP_NAME, "Verification successful.", parentWindow).Show()
		}
	})

	tab.Content = container.NewVBox(
		label,
		widget.NewSeparator(),
		subLabel,
		luminusCredentialsForm,
		luminusSaveButton,
		widget.NewSeparator(),
		canvasCredentialsForm,
		canvasSaveButton,
	)

	return tab, nil
}

// getPreferencesTab builds the preferences tab in the main UI.
func getPreferencesTab(parentWindow fyne.Window) (*container.TabItem, error) {
	logs.Logger.Debugln("preferences tab loaded")
	tab := container.NewTabItem("Preferences", container.NewVBox())

	fileDirHeader := widget.NewLabelWithStyle("File Directory", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0})
	fileDirSubHeader := widget.NewLabel("Root directory for your Luminus files:")

	dir := getPreferences().Directory
	if dir == "" {
		dir = "Not set"
	}

	fileDirLabel := widget.NewLabel(dir)
	fileDirLabel.Wrapping = fyne.TextWrapWord
	chooseDirButton := widget.NewButton("Choose directory", func() {
		dir, dirErr := fileDialog.Directory().Title("Choose directory").Browse()
		if dirErr != nil {
			if dirErr.Error() != "Cancelled" {
				logs.Logger.Debugln("directory selection cancelled")
				dialog.NewInformation(lominus.APP_NAME, "An error has occurred :( Please try again", parentWindow).Show()
				logs.Logger.Errorln(dirErr)
			}
			return
		}
		logs.Logger.Debugf("directory chosen - %s", dir)

		preferences := getPreferences()
		preferences.Directory = dir

		preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
		if getPreferencesPathErr != nil {
			dialog.NewInformation(lominus.APP_NAME, "An error has occurred :( Please try again", parentWindow).Show()
			logs.Logger.Errorln(getPreferencesPathErr)
			return
		}

		savePrefErr := appPref.SavePreferences(preferencesPath, preferences)
		if savePrefErr != nil {
			dialog.NewInformation(lominus.APP_NAME, "An error has occurred :( Please try again", parentWindow).Show()
			logs.Logger.Errorln(savePrefErr)
			return
		}
		logs.Logger.Debugln("directory saved")
		fileDirLabel.SetText(preferences.Directory)
	})

	frequencyHeader := widget.NewLabelWithStyle("Sync Frequency", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0})
	frequencySubHeader1 := widget.NewRichTextFromMarkdown("Lominus helps to sync files and more from [Luminus](https://luminus.nus.edu.sg) **automatically**.")
	frequencySubHeader2 := widget.NewRichTextFromMarkdown("Frequency denotes the number of **hours** between each sync.")

	frequencySelect := widget.NewSelect([]string{FREQUENCY_DISABLED, FREQUENCY_ONE_HOUR, FREQUENCY_TWO_HOUR, FREQUENCY_FOUR_HOUR, FREQUENCY_SIX_HOUR, FREQUENCY_TWELVE_HOUR}, func(s string) {
		preferences := getPreferences()
		switch s {
		case FREQUENCY_DISABLED:
			preferences.Frequency = -1
		case FREQUENCY_ONE_HOUR:
			preferences.Frequency = 1
		case FREQUENCY_TWO_HOUR:
			preferences.Frequency = 2
		case FREQUENCY_FOUR_HOUR:
			preferences.Frequency = 4
		case FREQUENCY_SIX_HOUR:
			preferences.Frequency = 6
		case FREQUENCY_TWELVE_HOUR:
			preferences.Frequency = 12
		default:
			preferences.Frequency = 1
		}

		logs.Logger.Debugf("frequency selected - %d", preferences.Frequency)

		preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
		if getPreferencesPathErr != nil {
			dialog.NewInformation(lominus.APP_NAME, "An error has occurred :( Please try again", parentWindow).Show()
			logs.Logger.Errorln(getPreferencesPathErr)
			return
		}

		savePrefErr := appPref.SavePreferences(preferencesPath, preferences)
		if savePrefErr != nil {
			dialog.NewInformation(lominus.APP_NAME, "An error has occurred :( Please try again", parentWindow).Show()
			logs.Logger.Errorln(savePrefErr)
			return
		}
		logs.Logger.Debugln("frequency saved")
	})
	frequencySelect.Selected = frequencyMap[getPreferences().Frequency]

	debugCheckbox := widget.NewCheck("Debug Mode", func(onDebug bool) {
		preferences := getPreferences()
		preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
		if getPreferencesPathErr != nil {
			dialog.NewInformation(lominus.APP_NAME, "An error has occurred :( Please try again", parentWindow).Show()
			logs.Logger.Errorln(getPreferencesPathErr)
			return
		}

		if onDebug {
			preferences.LogLevel = "debug"
		} else {
			preferences.LogLevel = "info"
		}

		logs.SetLogLevel(preferences.LogLevel)
		logs.Logger.Debugf("debug mode changed to - %v", onDebug)

		savePrefErr := appPref.SavePreferences(preferencesPath, preferences)
		if savePrefErr != nil {
			dialog.NewInformation(lominus.APP_NAME, "An error has occurred :( Please try again", parentWindow).Show()
			logs.Logger.Errorln(savePrefErr)
			return
		}

		dialog.NewInformation(lominus.APP_NAME, "Please restart Lominus for changes to take place.", parentWindow).Show()
	})

	debugCheckbox.Checked = getPreferences().LogLevel == "debug"

	tab.Content = container.NewVBox(
		fileDirHeader,
		widget.NewSeparator(),
		fileDirSubHeader,
		fileDirLabel,
		chooseDirButton,
		frequencyHeader,
		widget.NewSeparator(),
		frequencySubHeader1,
		frequencySubHeader2,
		frequencySelect,
		widget.NewSeparator(),
		debugCheckbox,
	)

	return tab, nil
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
