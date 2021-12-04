package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	"github.com/beebeeoii/lominus/internal/file"
	lominus "github.com/beebeeoii/lominus/internal/lominus"
	"github.com/beebeeoii/lominus/pkg/auth"
	"github.com/beebeeoii/lominus/pkg/pref"
	fileDialog "github.com/sqweek/dialog"
)

func Init() error {
	app := app.NewWithID(lominus.APP_NAME)
	app.SetIcon(resourceAppIconPng)

	w := app.NewWindow(fmt.Sprintf("%s v%s", lominus.APP_NAME, lominus.APP_VERSION))
	header := widget.NewLabelWithStyle(fmt.Sprintf("%s v%s", lominus.APP_NAME, lominus.APP_VERSION), fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0})

	credentialsUi, credentialsUiErr := getCredentialsUi(w)
	if credentialsUiErr != nil {
		return credentialsUiErr
	}

	directoryUi, directoryUiErr := getDirectoryUi(w)
	if directoryUiErr != nil {
		return directoryUiErr
	}
	content := container.NewVBox(header, credentialsUi, directoryUi)

	w.SetContent(content)
	w.Resize(fyne.NewSize(600, 600))
	w.SetFixedSize(true)
	w.SetOnClosed(func() {
		onMinimise()
	})
	w.ShowAndRun()
	return nil
}

func getCredentialsUi(parentWindow fyne.Window) (*fyne.Container, error) {
	divider := widget.NewSeparator()
	label := widget.NewLabelWithStyle("Login Info", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0})
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
			return container.NewVBox(), err
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

	return container.NewVBox(label, divider, subLabel, credentialsForm, saveButton), nil
}

func getDirectoryUi(parentWindow fyne.Window) (*fyne.Container, error) {
	divider := widget.NewSeparator()
	label := widget.NewLabelWithStyle("File Directory", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0})
	subLabel := widget.NewLabel("Root directory for your Luminus files:")

	fileDirLabel := widget.NewLabel(getDir())
	fileDirLabel.Wrapping = fyne.TextWrapBreak
	chooseDirButton := widget.NewButton("Choose directory", func() {
		dir, err := fileDialog.Directory().Title("Choose directory").Browse()
		if err != nil {
			if err.Error() != "Cancelled" {
				dialog.NewInformation(lominus.APP_NAME, "An error has occurred :( Please try again or contact us.", parentWindow).Show()
				log.Println(err)
			}
			return
		}

		prefPath := appPref.GetPreferencesPath()
		currentPref, loadPrefErr := pref.LoadPreferences(prefPath)
		if loadPrefErr != nil {
			dialog.NewInformation(lominus.APP_NAME, "An error has occurred :( Please try again or contact us.", parentWindow).Show()
			log.Println(loadPrefErr)
			return
		}

		pref.SavePreferences(prefPath, appPref.Preferences{Directory: dir, Frequency: currentPref.Frequency})
		fileDirLabel.SetText(getDir())
	})

	return container.NewVBox(label, divider, subLabel, fileDirLabel, chooseDirButton), nil
}

func getDir() string {
	rootDir := "Not set"
	prefPath := appPref.GetPreferencesPath()
	if file.Exists(prefPath) {
		preferences, err := pref.LoadPreferences(prefPath)
		if err != nil {
			log.Fatalln(err)
		}

		if preferences.Directory != "" {
			rootDir = preferences.Directory
		}
	}
	return rootDir
}
