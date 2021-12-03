package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	lominus "github.com/beebeeoii/lominus/internal/app"
	"github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/pkg/auth"
	"github.com/beebeeoii/lominus/pkg/pref"
	fileDialog "github.com/sqweek/dialog"
)

func Init() {
	app := app.New()

	w := app.NewWindow(fmt.Sprintf("%s v%s", lominus.APP_NAME, lominus.APP_VERSION))
	header := widget.NewLabelWithStyle(fmt.Sprintf("%s v%s", lominus.APP_NAME, lominus.APP_VERSION), fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0})

	content := container.NewVBox(header, getCredentialsUi(w), getDirectoryUi())

	w.SetContent(content)
	w.Resize(fyne.NewSize(600, 600))
	w.SetFixedSize(true)
	w.ShowAndRun()
}

func getCredentialsUi(parentWindow fyne.Window) *fyne.Container {
	divider := widget.NewSeparator()
	label := widget.NewLabelWithStyle("Login Info", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0})
	subLabel := widget.NewLabel("Your credentials are saved locally.\nIt is ONLY used to login to your account on https://luminus.nus.edu.sg.")

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Eg: nusstu\\e0123456")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")
	if file.Exists(auth.CREDENTIALS_FILE_NAME) {
		credentials, err := auth.LoadCredentials()
		if err != nil {
			log.Fatalln(err)
		}

		usernameEntry.SetText(credentials.Username)
		passwordEntry.SetText(credentials.Password)
	}

	saveButtonText := "Save Credentials"
	if usernameEntry.Text != "" && passwordEntry.Text != "" {
		saveButtonText = "Update Credentials"
	}

	saveButton := widget.NewButton(saveButtonText, func() {
		credentials := auth.Credentials{Username: usernameEntry.Text, Password: passwordEntry.Text}

		status := widget.NewLabel("Please wait while we verify your credentials...")
		progressBar := widget.NewProgressBarInfinite()

		mainDialog := dialog.NewCustom(lominus.APP_NAME, "Cancel", container.NewVBox(status, progressBar), parentWindow)
		mainDialog.Show()

		_, err := auth.RetrieveJwtToken(credentials, true)
		mainDialog.Hide()
		if err != nil {
			dialog.NewInformation(lominus.APP_NAME, "Verification failed. Please check your credentials.", parentWindow).Show()
		} else {
			auth.SaveCredentials(credentials)
			dialog.NewInformation(lominus.APP_NAME, "Verification successful.", parentWindow).Show()
		}
	})

	return container.NewVBox(label, divider, subLabel, usernameEntry, passwordEntry, saveButton)
}

func getDirectoryUi() *fyne.Container {

	divider := widget.NewSeparator()
	label := widget.NewLabelWithStyle("File Directory", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0})
	subLabel := widget.NewLabel("Root directory for your Luminus files:")

	fileDirLabel := widget.NewLabel(getDir())
	chooseDirButton := widget.NewButton("Choose directory", func() {
		dir, err := fileDialog.Directory().Title("Choose directory").Browse()
		if err != nil {
			log.Println(err)
		}
		pref.SavePreferences(pref.Preferences{Directory: dir, Frequency: 0})
		fileDirLabel.SetText(getDir())
	})

	return container.NewVBox(label, divider, subLabel, fileDirLabel, chooseDirButton)
}

func getDir() string {
	rootDir := "Not set"
	if file.Exists(pref.PREFERENCES_FILE_NAME) {
		preferences, err := pref.LoadPreferences()
		if err != nil {
			log.Fatalln(err)
		}

		if preferences.Directory != "" {
			rootDir = preferences.Directory
		}
	}
	return rootDir
}
