// Package ui provides primitives that initialises the UI.
package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/pkg/integrations/telegram"

	appInt "github.com/beebeeoii/lominus/internal/app/integrations/telegram"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
)

type IntegrationData struct {
	TelegramUserId string
	TelegramBotId  string
}

// getIntegrationsTab builds the integrations tab in the main UI.
func getIntegrationsTab(integrationData IntegrationData, parentWindow fyne.Window) (*container.TabItem, error) {
	logs.Logger.Debugln("integrations tab loaded")
	tab := container.NewTabItem(appConstants.INTEGRATIONS_TITLE, container.NewVBox())

	label := widget.NewLabelWithStyle(
		appConstants.TELEGRAM_TITLE,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0},
	)
	description := widget.NewRichTextFromMarkdown(appConstants.TELEGRAM_DESCRIPTION)
	description.Wrapping = fyne.TextWrapWord

	botApiEntry := widget.NewPasswordEntry()
	botApiEntry.SetPlaceHolder(appConstants.TELEGRAM_BOT_TOKEN_PLACEHOLDER)
	userIdEntry := widget.NewEntry()
	userIdEntry.SetPlaceHolder(appConstants.TELEGRAM_USER_ID_PLACEHOLDER)

	botApiEntry.SetText(integrationData.TelegramBotId)
	userIdEntry.SetText(integrationData.TelegramUserId)

	telegramForm := widget.NewForm(
		widget.NewFormItem(appConstants.TELEGRAM_BOT_TOKEN_TEXT, botApiEntry),
		widget.NewFormItem(appConstants.TELEGRAM_USER_ID_TEXT, userIdEntry),
	)

	saveButton := widget.NewButton(appConstants.SAVE_TELEGRAM_DATA_TEXT, func() {
		botApi := botApiEntry.Text
		userId := userIdEntry.Text

		status := widget.NewLabel(appConstants.TELEGRAM_TESTING_MESSAGE)
		progressBar := widget.NewProgressBarInfinite()

		mainDialog := dialog.NewCustom(
			appConstants.APP_NAME,
			appConstants.CANCEL_TEXT,
			container.NewVBox(status, progressBar),
			parentWindow,
		)
		mainDialog.Show()

		logs.Logger.Debugln("sending telegram test message")
		err := telegram.SendMessage(botApi, userId, appConstants.TELEGRAM_DEFAULT_TEST_MESSAGE)
		mainDialog.Hide()
		if err != nil {
			errMessage := fmt.Sprintf(
				"%s: %s",
				err.Error()[:13],
				err.Error()[strings.Index(err.Error(), "description")+14:len(err.Error())-2],
			)
			logs.Logger.Errorln(errMessage)
			dialog.NewInformation(
				appConstants.APP_NAME,
				appConstants.TELEGRAM_TESTING_FAILED_MESSAGE,
				parentWindow,
			).Show()
		} else {
			saveErr := appInt.SaveTelegramCredentials(userId, botApi)

			if saveErr != nil {
				dialog.NewInformation(
					appConstants.APP_NAME,
					appConstants.TELEGRAM_TESTING_FAILED_MESSAGE,
					parentWindow,
				).Show()
				return
			}

			logs.Logger.Debugln("telegram test message sent successfully")
			dialog.NewInformation(
				appConstants.APP_NAME,
				appConstants.TELEGRAM_TESTING_SUCCESSFUL_MESSAGE,
				parentWindow,
			).Show()
		}
	})

	tab.Content = container.NewVBox(
		label,
		widget.NewSeparator(),
		description,
		telegramForm,
		saveButton,
	)

	return tab, nil
}
