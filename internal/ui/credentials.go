// Package ui provides primitives that initialises the UI.
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	"github.com/beebeeoii/lominus/internal/file"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/pkg/auth"
)

var credentialsPath string
var tokensPath string

// getCredentialsTab builds the credentials tab in the main UI.
func getCredentialsTab(parentWindow fyne.Window) (*container.TabItem, error) {
	logs.Logger.Debugln("credentials tab loaded")
	tab := container.NewTabItem(appConstants.CREDENTIALS_TITLE, container.NewVBox())

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

	luminusView, luminusViewErr := getLuminusView(w, credentials.LuminusCredentials)
	if luminusViewErr != nil {
		return tab, luminusViewErr
	}

	canvasView, canvasViewErr := getCanvasView(w, credentials.CanvasCredentials)
	if canvasViewErr != nil {
		return tab, canvasViewErr
	}

	tab.Content = container.NewVBox(luminusView, canvasView)

	return tab, nil
}

// getLuminusView builds the view for Luminus credentials placed in the credentials tab.
func getLuminusView(
	parentWindow fyne.Window,
	defaultCredentials auth.LuminusCredentials,
) (fyne.CanvasObject, error) {
	logs.Logger.Debugln("luminus view loaded")

	label := widget.NewLabelWithStyle(
		appConstants.LUMINUS_TAB_TITLE,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0},
	)
	description := widget.NewRichTextFromMarkdown(appConstants.LUMINUS_TAB_DESCRIPTION)
	description.Wrapping = fyne.TextWrapWord

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder(appConstants.LUMINUS_USERNAME_PLACEHOLDER)
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder(appConstants.LUMINUS_PASSWORD_PLACEHOLDER)

	usernameEntry.SetText(defaultCredentials.Username)
	passwordEntry.SetText(defaultCredentials.Password)

	luminusCredentialsForm := widget.NewForm(
		widget.NewFormItem(appConstants.LUMINUS_USERNAME_TEXT, usernameEntry),
		widget.NewFormItem(appConstants.LUMINUS_PASSWORD_TEXT, passwordEntry),
	)

	luminusSaveButton := widget.NewButton(appConstants.SAVE_CREDENTIALS_TEXT, func() {
		luminusCredentials := auth.LuminusCredentials{
			Username: usernameEntry.Text,
			Password: passwordEntry.Text,
		}

		status := widget.NewLabel(appConstants.VERIFYING_MESSAGE)
		progressBar := widget.NewProgressBarInfinite()

		mainDialog := dialog.NewCustom(
			appConstants.APP_NAME,
			appConstants.CANCEL_TEXT,
			container.NewVBox(status, progressBar),
			parentWindow,
		)
		mainDialog.Show()

		logs.Logger.Debugln("verifying credentials")
		_, err := auth.RetrieveJwtToken(luminusCredentials, true)
		mainDialog.Hide()
		if err != nil {
			logs.Logger.Debugln("verfication failed")
			dialog.NewInformation(
				appConstants.APP_NAME,
				appConstants.VERIFICATION_FAILED_MESSAGE,
				parentWindow,
			).Show()
		} else {
			logs.Logger.Debugln("verfication succesful - saving credentials")
			luminusCredentials.Save(credentialsPath)
			dialog.NewInformation(
				appConstants.APP_NAME,
				appConstants.VERIFICATION_SUCCESSFUL_MESSAGE,
				parentWindow,
			).Show()
		}
	})

	return container.NewVBox(
		label,
		widget.NewSeparator(),
		description,
		luminusCredentialsForm,
		luminusSaveButton,
	), nil
}

// getCanvasView builds the view for Canvas credentials placed in the credentials tab.
func getCanvasView(
	parentWindow fyne.Window,
	defaultCredentials auth.CanvasCredentials,
) (fyne.CanvasObject, error) {
	logs.Logger.Debugln("canvas view loaded")

	label := widget.NewLabelWithStyle(
		appConstants.CANVAS_TAB_TITLE,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0},
	)
	description := widget.NewRichTextFromMarkdown(appConstants.CANVAS_TAB_DESCRIPTION)
	description.Wrapping = fyne.TextWrapWord
	canvasTokenEntry := widget.NewPasswordEntry()
	canvasTokenEntry.SetPlaceHolder(appConstants.CANVAS_TOKEN_PLACEHOLDER)

	canvasTokenEntry.SetText(defaultCredentials.CanvasApiToken)

	canvasCredentialsForm := widget.NewForm(
		widget.NewFormItem(appConstants.CANVAS_TOKEN_TEXT, canvasTokenEntry),
	)

	canvasSaveButton := widget.NewButton(appConstants.SAVE_CREDENTIALS_TEXT, func() {
		canvasCredentials := auth.CanvasCredentials{
			CanvasApiToken: canvasTokenEntry.Text,
		}

		canvasTokens := auth.CanvasTokenData{
			CanvasApiToken: canvasTokenEntry.Text,
		}

		status := widget.NewLabel(appConstants.VERIFYING_MESSAGE)
		progressBar := widget.NewProgressBarInfinite()

		mainDialog := dialog.NewCustom(
			appConstants.APP_NAME,
			appConstants.CANCEL_TEXT,
			container.NewVBox(status, progressBar),
			parentWindow,
		)
		mainDialog.Show()

		logs.Logger.Debugln("verifying credentials")
		err := canvasCredentials.Authenticate()
		mainDialog.Hide()
		if err != nil {
			logs.Logger.Debugln("verfication failed")
			dialog.NewInformation(
				appConstants.APP_NAME,
				appConstants.VERIFICATION_FAILED_MESSAGE,
				parentWindow,
			).Show()
		} else {
			logs.Logger.Debugln("verfication succesful - saving credentials")
			canvasCredentials.Save(credentialsPath)
			canvasTokens.Save(tokensPath)
			dialog.NewInformation(
				appConstants.APP_NAME,
				appConstants.VERIFICATION_SUCCESSFUL_MESSAGE,
				parentWindow,
			).Show()
		}
	})

	return container.NewVBox(
		label,
		widget.NewSeparator(),
		description,
		canvasCredentialsForm,
		canvasSaveButton,
	), nil
}
