package ui

import (
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	"github.com/beebeeoii/lominus/internal/cron"
	"github.com/beebeeoii/lominus/internal/file"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/lominus"
	"github.com/beebeeoii/lominus/internal/notifications"
	"github.com/beebeeoii/lominus/pkg/auth"
	"github.com/beebeeoii/lominus/pkg/pref"
	"github.com/getlantern/systray"
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

func Init() error {
	if runtime.GOOS == "windows" {
		systray.Register(onReady, onExit)
	}
	mainApp = app.NewWithID(lominus.APP_NAME)
	mainApp.SetIcon(resourceAppIconPng)

	go func() {
		for {
			notification := <-notifications.NotificationChannel
			mainApp.SendNotification(fyne.NewNotification(notification.Title, notification.Content))
		}
	}()

	w = mainApp.NewWindow(lominus.APP_NAME)
	header := widget.NewLabelWithStyle(fmt.Sprintf("%s v%s", lominus.APP_NAME, lominus.APP_VERSION), fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0})

	credentialsTab, credentialsUiErr := getCredentialsTab(w)
	if credentialsUiErr != nil {
		return credentialsUiErr
	}

	preferencesTab, preferencesErr := getPreferencesTab(w)
	if preferencesErr != nil {
		return preferencesErr
	}

	tabsContainer := container.NewAppTabs(credentialsTab, preferencesTab)
	content := container.NewVBox(
		header,
		tabsContainer,
		layout.NewSpacer(),
		getSyncButton(w),
		getQuitButton(),
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(600, 600))
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

func getCredentialsTab(parentWindow fyne.Window) (*container.TabItem, error) {
	tab := container.NewTabItem("Login Info", container.NewVBox())

	label := widget.NewLabelWithStyle("Your Credentials", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0})
	subLabel := widget.NewRichTextFromMarkdown("Credentials are saved **locally**. It is **only** used to login to your [Luminus](https://luminus.nus.edu.sg) account.")
	subLabel.Wrapping = fyne.TextWrapBreak

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Eg: nusstu\\e0123456")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	credentialsPath := appAuth.GetCredentialsPath()

	if file.Exists(credentialsPath) {
		credentials, err := auth.LoadCredentials(credentialsPath)
		if err != nil {
			return tab, err
		}

		usernameEntry.SetText(credentials.Username)
		passwordEntry.SetText(credentials.Password)
	}

	credentialsForm := widget.NewForm(widget.NewFormItem("Username", usernameEntry), widget.NewFormItem("Password", passwordEntry))

	saveButtonText := "Save Credentials"
	if usernameEntry.Text != "" && passwordEntry.Text != "" {
		saveButtonText = "Update Credentials"
	}

	saveButton := widget.NewButton(saveButtonText, func() {
		credentials := appAuth.Credentials{Username: usernameEntry.Text, Password: passwordEntry.Text}

		status := widget.NewLabel("Please wait while we verify your credentials...")
		progressBar := widget.NewProgressBarInfinite()

		mainDialog := dialog.NewCustom(lominus.APP_NAME, "Cancel", container.NewVBox(status, progressBar), parentWindow)
		mainDialog.Show()

		_, err := auth.RetrieveJwtToken(credentials, true)
		mainDialog.Hide()
		if err != nil {
			dialog.NewInformation(lominus.APP_NAME, "Verification failed. Please check your credentials.", parentWindow).Show()
		} else {
			auth.SaveCredentials(credentialsPath, credentials)
			dialog.NewInformation(lominus.APP_NAME, "Verification successful.", parentWindow).Show()
		}
	})

	tab.Content = container.NewVBox(
		label,
		widget.NewSeparator(),
		subLabel,
		credentialsForm,
		saveButton,
	)

	return tab, nil
}

func getPreferencesTab(parentWindow fyne.Window) (*container.TabItem, error) {
	tab := container.NewTabItem("Preferences", container.NewVBox())

	fileDirHeader := widget.NewLabelWithStyle("File Directory", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0})
	fileDirSubHeader := widget.NewLabel("Root directory for your Luminus files:")

	dir := getPreferences().Directory
	if dir == "" {
		dir = "Not set"
	}

	fileDirLabel := widget.NewLabel(dir)
	fileDirLabel.Wrapping = fyne.TextWrapBreak
	chooseDirButton := widget.NewButton("Choose directory", func() {
		dir, dirErr := fileDialog.Directory().Title("Choose directory").Browse()
		if dirErr != nil {
			if dirErr.Error() != "Cancelled" {
				dialog.NewInformation(lominus.APP_NAME, "An error has occurred :( Please try again or contact us.", parentWindow).Show()
				logs.ErrorLogger.Println(dirErr)
			}
			return
		}

		preferences := getPreferences()
		preferences.Directory = dir

		savePrefErr := pref.SavePreferences(appPref.GetPreferencesPath(), preferences)
		if savePrefErr != nil {
			dialog.NewInformation(lominus.APP_NAME, "An error has occurred :( Please try again or contact us.", parentWindow).Show()
			logs.ErrorLogger.Println(savePrefErr)
			return
		}
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

		savePrefErr := pref.SavePreferences(appPref.GetPreferencesPath(), preferences)
		if savePrefErr != nil {
			dialog.NewInformation(lominus.APP_NAME, "An error has occurred :( Please try again or contact us.", parentWindow).Show()
			logs.ErrorLogger.Println(savePrefErr)
			return
		}
	})
	frequencySelect.Selected = frequencyMap[getPreferences().Frequency]

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
	)

	return tab, nil
}

func getPreferences() appPref.Preferences {
	preference, err := pref.LoadPreferences(appPref.GetPreferencesPath())
	if err != nil {
		logs.ErrorLogger.Fatalln(err)
	}

	return preference
}

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

func getQuitButton() *widget.Button {
	return widget.NewButton("Quit Lominus", func() {
		if getOs() == "windows" {
			logs.InfoLogger.Println("systray quit")
			systray.Quit()
		}
		logs.InfoLogger.Println("lominus quit")
		mainApp.Quit()
	})
}

func getOs() string {
	return runtime.GOOS
}
