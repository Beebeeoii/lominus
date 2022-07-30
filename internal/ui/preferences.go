package ui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/beebeeoii/lominus/internal/lominus"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	logs "github.com/beebeeoii/lominus/internal/log"
)

var frequencyMap = map[int]string{
	1:  appConstants.SYNC_FREQUENCY_ONE_HOUR,
	2:  appConstants.SYNC_FREQUENCY_TWO_HOUR,
	4:  appConstants.SYNC_FREQUENCY_FOUR_HOUR,
	6:  appConstants.SYNC_FREQUENCY_SIX_HOUR,
	12: appConstants.SYNC_FREQUENCY_TWELVE_HOUR,
	-1: appConstants.SYNC_FREQUENCY_DISABLED,
}

// getPreferencesTab builds the preferences tab in the main UI.
func getPreferencesTab(parentWindow fyne.Window) (*container.TabItem, error) {
	logs.Logger.Debugln("preferences tab loaded")
	tab := container.NewTabItem(appConstants.PREFERENCES_TITLE, container.NewVBox())

	fileDirectoryView, fileDirectoryViewErr := getFileDirectoryView(w)
	if fileDirectoryViewErr != nil {
		return tab, fileDirectoryViewErr
	}

	syncView, syncViewErr := getSyncView(w)
	if syncViewErr != nil {
		return tab, syncViewErr
	}

	advancedView, advancedViewErr := getAdvancedView(w)
	if advancedViewErr != nil {
		return tab, advancedViewErr
	}

	tab.Content = container.NewVBox(fileDirectoryView, syncView, advancedView)

	return tab, nil
}

func getFileDirectoryView(parentWindow fyne.Window) (fyne.CanvasObject, error) {
	logs.Logger.Debugln("file directory view loaded")

	label := widget.NewLabelWithStyle(
		appConstants.FILE_DIRECTORY_TAB_TITLE,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0},
	)
	description := widget.NewRichTextFromMarkdown(appConstants.FILE_DIRECTORY_TAB_DESCRIPTION)
	description.Wrapping = fyne.TextWrapWord

	dir := getPreferences().Directory
	if dir == "" {
		dir = appConstants.FILE_DIRECTORY_FOLDER_PATH_DEFAULT
	}

	folderPathLabel := widget.NewLabel(dir)
	folderPathLabel.Wrapping = fyne.TextWrapWord
	chooseDirButton := widget.NewButton(appConstants.FILE_DIRECTORY_SELECT_DIRECTORY_TEXT, func() {
		dialog.ShowFolderOpen(func(lu fyne.ListableURI, dirErr error) {
			if dirErr != nil {
				logs.Logger.Debugln("directory selection cancelled")
				dialog.NewInformation(
					lominus.APP_NAME,
					appConstants.PREFERENCES_FAILED_MESSAGE,
					parentWindow,
				).Show()
				logs.Logger.Errorln(dirErr)
				return
			}

			if lu == nil {
				return
			}

			dir = lu.Path()

			logs.Logger.Debugf("directory chosen - %s", dir)

			preferences := getPreferences()
			preferences.Directory = dir

			preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
			if getPreferencesPathErr != nil {
				dialog.NewInformation(
					lominus.APP_NAME,
					appConstants.PREFERENCES_FAILED_MESSAGE,
					parentWindow,
				).Show()
				logs.Logger.Errorln(getPreferencesPathErr)
				return
			}

			savePrefErr := appPref.SavePreferences(preferencesPath, preferences)
			if savePrefErr != nil {
				dialog.NewInformation(
					lominus.APP_NAME,
					appConstants.PREFERENCES_FAILED_MESSAGE,
					parentWindow,
				).Show()
				logs.Logger.Errorln(savePrefErr)
				return
			}
			logs.Logger.Debugln("directory saved")
			folderPathLabel.SetText(preferences.Directory)
		}, parentWindow)
	})

	return container.NewVBox(label, widget.NewSeparator(), description, folderPathLabel, chooseDirButton), nil
}

func getSyncView(parentWindow fyne.Window) (fyne.CanvasObject, error) {
	logs.Logger.Debugln("sync view loaded")

	label := widget.NewLabelWithStyle(
		appConstants.SYNC_TAB_TITLE,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0},
	)
	description := widget.NewRichTextFromMarkdown(appConstants.SYNC_TAB_DESCRIPTION)
	description.Wrapping = fyne.TextWrapWord

	frequencySelect := widget.NewSelect([]string{
		appConstants.SYNC_FREQUENCY_DISABLED,
		appConstants.SYNC_FREQUENCY_ONE_HOUR,
		appConstants.SYNC_FREQUENCY_TWO_HOUR,
		appConstants.SYNC_FREQUENCY_FOUR_HOUR,
		appConstants.SYNC_FREQUENCY_SIX_HOUR,
		appConstants.SYNC_FREQUENCY_TWELVE_HOUR,
	}, func(s string) {
		preferences := getPreferences()
		switch s {
		case appConstants.SYNC_FREQUENCY_DISABLED:
			preferences.Frequency = -1
		case appConstants.SYNC_FREQUENCY_ONE_HOUR:
			preferences.Frequency = 1
		case appConstants.SYNC_FREQUENCY_TWO_HOUR:
			preferences.Frequency = 2
		case appConstants.SYNC_FREQUENCY_FOUR_HOUR:
			preferences.Frequency = 4
		case appConstants.SYNC_FREQUENCY_SIX_HOUR:
			preferences.Frequency = 6
		case appConstants.SYNC_FREQUENCY_TWELVE_HOUR:
			preferences.Frequency = 12
		default:
			preferences.Frequency = 1
		}

		logs.Logger.Debugf("frequency selected - %d", preferences.Frequency)

		preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
		if getPreferencesPathErr != nil {
			dialog.NewInformation(
				lominus.APP_NAME,
				appConstants.PREFERENCES_FAILED_MESSAGE,
				parentWindow,
			).Show()
			logs.Logger.Errorln(getPreferencesPathErr)
			return
		}

		savePrefErr := appPref.SavePreferences(preferencesPath, preferences)
		if savePrefErr != nil {
			dialog.NewInformation(
				lominus.APP_NAME,
				appConstants.PREFERENCES_FAILED_MESSAGE,
				parentWindow,
			).Show()
			logs.Logger.Errorln(savePrefErr)
			return
		}
		logs.Logger.Debugln("frequency saved")
	})
	frequencySelect.Selected = frequencyMap[getPreferences().Frequency]

	return container.NewVBox(label, widget.NewSeparator(), description, frequencySelect), nil
}

func getAdvancedView(parentWindow fyne.Window) (fyne.CanvasObject, error) {
	logs.Logger.Debugln("advanced view loaded")

	label := widget.NewLabelWithStyle(
		appConstants.ADVANCED_TAB_TITLE,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0},
	)

	var description *widget.RichText

	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		description = widget.NewRichTextFromMarkdown(
			appConstants.DEBUG_CHECKBOX_WO_LINK_DESCRIPTION,
		)
	} else {
		description = widget.NewRichTextFromMarkdown(
			fmt.Sprintf(
				appConstants.DEBUG_CHECKBOX_W_LINK_DESCRIPTION,
				filepath.Join(baseDir, lominus.LOG_FILE_NAME),
			),
		)
	}

	description.Wrapping = fyne.TextWrapWord

	debugCheckbox := widget.NewCheck(appConstants.DEBUG_CHECKBOX_TITLE, func(onDebug bool) {
		preferences := getPreferences()
		preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
		if getPreferencesPathErr != nil {
			dialog.NewInformation(
				lominus.APP_NAME,
				appConstants.PREFERENCES_FAILED_MESSAGE,
				parentWindow,
			).Show()
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
			dialog.NewInformation(
				lominus.APP_NAME,
				appConstants.PREFERENCES_FAILED_MESSAGE,
				parentWindow,
			).Show()
			logs.Logger.Errorln(savePrefErr)
			return
		}

		dialog.NewInformation(
			lominus.APP_NAME,
			appConstants.DEBUG_TOGGLE_SUCCESSFUL_MESSAGE,
			parentWindow,
		).Show()
	})

	debugCheckbox.Checked = getPreferences().LogLevel == "debug"

	return container.NewVBox(label, widget.NewSeparator(), description, debugCheckbox), nil
}
