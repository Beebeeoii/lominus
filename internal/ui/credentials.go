package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
	"github.com/beebeeoii/lominus/internal/file"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/lominus"
	"github.com/beebeeoii/lominus/pkg/auth"
)

const (
	SAVE_CREDENTIALS_TEXT = "Save Credentials"
)

var credentialsPath string
var tokensPath string

// getCredentialsTab builds the credentials tab in the main UI.
func getCredentialsTab(parentWindow fyne.Window) (*container.TabItem, error) {
	logs.Logger.Debugln("credentials tab loaded")
	tab := container.NewTabItem("Credentials", container.NewVBox())

	var credentials auth.CredentialsData
	cPath, getCredentialsPathErr := appAuth.GetCredentialsPath()
	if getCredentialsPathErr != nil {
		return tab, getCredentialsPathErr
	}
	credentialsPath = cPath

	if file.Exists(credentialsPath) {
		logs.Logger.Debugf("credentials exists - loading from %s", credentialsPath)
		c, err := auth.LoadCredentialsData(credentialsPath)
		if err != nil {
			return tab, err
		}

		credentials = c
	}

	tPath, getTokensPathErr := appAuth.GetTokensPath()
	if getTokensPathErr != nil {
		return tab, getTokensPathErr
	}
	tokensPath = tPath

	luminusSubTab, luminusSubTabErr := getLuminusSubTab(w, credentials.LuminusCredentials)
	if luminusSubTabErr != nil {
		return tab, luminusSubTabErr
	}

	canvasSubTab, canvasSubTabErr := getCanvasSubTab(w, credentials.CanvasCredentials)
	if canvasSubTabErr != nil {
		return tab, canvasSubTabErr
	}

	tabsContainer := container.NewAppTabs(luminusSubTab, canvasSubTab)
	tab.Content = container.NewVBox(tabsContainer)

	return tab, nil
}

func getLuminusSubTab(parentWindow fyne.Window, defaultCredentials auth.LuminusCredentials) (*container.TabItem, error) {
	logs.Logger.Debugln("luminus tab loaded")
	tab := container.NewTabItem("Luminus", container.NewVBox())

	description := widget.NewRichTextFromMarkdown("Credentials are saved **locally**. It is used to login to [Luminus](https://luminus.nus.edu.sg) **only**.")
	description.Wrapping = fyne.TextWrapWord

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Eg: nusstu\\e0123456")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	usernameEntry.SetText(defaultCredentials.Username)
	passwordEntry.SetText(defaultCredentials.Password)

	luminusCredentialsForm := widget.NewForm(widget.NewFormItem("Username", usernameEntry), widget.NewFormItem("Password", passwordEntry))

	luminusSaveButton := widget.NewButton(SAVE_CREDENTIALS_TEXT, func() {
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

	tab.Content = container.NewVBox(
		description,
		luminusCredentialsForm,
		luminusSaveButton,
	)

	return tab, nil
}

func getCanvasSubTab(parentWindow fyne.Window, defaultCredentials auth.CanvasCredentials) (*container.TabItem, error) {
	logs.Logger.Debugln("canvas tab loaded")
	tab := container.NewTabItem("Canvas", container.NewVBox())

	description := widget.NewRichTextFromMarkdown("Token is saved **locally**. It is used to access [Canvas](https://canvas.nus.edu.sg/) **only**.")
	description.Wrapping = fyne.TextWrapWord
	canvasTokenEntry := widget.NewPasswordEntry()
	canvasTokenEntry.SetPlaceHolder("Account > Settings > New access token > Generate Token")

	canvasTokenEntry.SetText(defaultCredentials.CanvasApiToken)

	canvasCredentialsForm := widget.NewForm(widget.NewFormItem("Canvas Token", canvasTokenEntry))

	canvasSaveButton := widget.NewButton(SAVE_CREDENTIALS_TEXT, func() {
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
		description,
		canvasCredentialsForm,
		canvasSaveButton,
	)

	return tab, nil
}
