// Package ui provides primitives that initialises the UI.
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/pkg/auth"
)

type CredentialsData struct {
	CanvasApiToken string
}

// getCredentialsTab builds the credentials tab in the main UI.
func getCredentialsTab(credentialsData CredentialsData, parentWindow fyne.Window) (*container.TabItem, error) {
	logs.Logger.Debugln("credentials tab loaded")
	tab := container.NewTabItem(appConstants.CREDENTIALS_TITLE, container.NewVBox())

	canvasView, canvasViewErr := getCanvasView(w, credentialsData.CanvasApiToken)
	if canvasViewErr != nil {
		return tab, canvasViewErr
	}

	tab.Content = container.NewVBox(canvasView)

	return tab, nil
}

// getCanvasView builds the view for Canvas credentials placed in the credentials tab.
func getCanvasView(
	parentWindow fyne.Window,
	defaultApiToken string,
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

	canvasTokenEntry.SetText(defaultApiToken)

	canvasCredentialsForm := widget.NewForm(
		widget.NewFormItem(appConstants.CANVAS_TOKEN_TEXT, canvasTokenEntry),
	)

	canvasSaveButton := widget.NewButton(appConstants.SAVE_CREDENTIALS_TEXT, func() {
		canvasCredentials := auth.CanvasCredentials{
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

			saveErr := appAuth.SaveCanvasCredentials(appAuth.CanvasCredentials{
				CanvasApiToken: canvasTokenEntry.Text,
			})

			if saveErr != nil {
				dialog.NewInformation(
					appConstants.APP_NAME,
					appConstants.VERIFICATION_FAILED_MESSAGE,
					parentWindow,
				).Show()
				logs.Logger.Errorln(saveErr)
				return
			}

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
