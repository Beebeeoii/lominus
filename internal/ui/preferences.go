// Package ui provides primitives that initialises the UI.
package ui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	logs "github.com/beebeeoii/lominus/internal/log"
	fileDialog "github.com/sqweek/dialog"
)

var frequencyMap = map[int]string{
	1:  appConstants.SYNC_FREQUENCY_ONE_HOUR,
	2:  appConstants.SYNC_FREQUENCY_TWO_HOUR,
	4:  appConstants.SYNC_FREQUENCY_FOUR_HOUR,
	6:  appConstants.SYNC_FREQUENCY_SIX_HOUR,
	12: appConstants.SYNC_FREQUENCY_TWELVE_HOUR,
	-1: appConstants.SYNC_FREQUENCY_DISABLED,
}

type PreferencesData struct {
	Directory string
	Frequency int
	LogLevel  string
}

// getPreferencesTab builds the preferences tab in the main UI.
func getPreferencesTab(preferencesData PreferencesData, _ fyne.Window) (*container.TabItem, error) {
	logs.Logger.Debugln("preferences tab loaded")
	tab := container.NewTabItem(appConstants.PREFERENCES_TITLE, container.NewVBox())

	fileDirectoryView, fileDirectoryViewErr := getFileDirectoryView(w, preferencesData.Directory)
	if fileDirectoryViewErr != nil {
		return tab, fileDirectoryViewErr
	}

	syncView, syncViewErr := getSyncView(w, preferencesData.Frequency)
	if syncViewErr != nil {
		return tab, syncViewErr
	}

	advancedView, advancedViewErr := getAdvancedView(w, preferencesData.LogLevel)
	if advancedViewErr != nil {
		return tab, advancedViewErr
	}

	tab.Content = container.NewVBox(fileDirectoryView, syncView, advancedView)

	return tab, nil
}

// getFileDirectoryView builds the view for choosing folder directory for LMS files
// to be stored locally. It is placed in the Preferences tab.
func getFileDirectoryView(parentWindow fyne.Window, directory string) (fyne.CanvasObject, error) {
	logs.Logger.Debugln("file directory view loaded")

	label := widget.NewLabelWithStyle(
		appConstants.FILE_DIRECTORY_TAB_TITLE,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true, Italic: false, Monospace: false, TabWidth: 0},
	)

	if directory == "" {
		directory = appConstants.FILE_DIRECTORY_FOLDER_PATH_DEFAULT
	}

	folderPathLabel := widget.NewLabel(directory)
	folderPathLabel.Wrapping = fyne.TextWrapWord
	chooseDirButton := widget.NewButton(appConstants.FILE_DIRECTORY_SELECT_DIRECTORY_TEXT, func() {
		dir, dirErr := fileDialog.Directory().Title(
			appConstants.FILE_DIRECTORY_SELECT_DIRECTORY_TEXT,
		).Browse()

		if dirErr != nil {
			if dirErr.Error() != "Cancelled" {
				logs.Logger.Debugln("directory selection cancelled")
				dialog.NewInformation(
					appConstants.APP_NAME,
					appConstants.PREFERENCES_FAILED_MESSAGE,
					parentWindow,
				).Show()
				logs.Logger.Errorln(dirErr)
			}
			return
		}
		logs.Logger.Debugf("directory chosen - %s", dir)

		savePrefErr := appPref.SaveRootSyncDirectory(dir)
		if savePrefErr != nil {
			dialog.NewInformation(
				appConstants.APP_NAME,
				appConstants.PREFERENCES_FAILED_MESSAGE,
				parentWindow,
			).Show()
			logs.Logger.Errorln(savePrefErr)
			return
		}

		logs.Logger.Debugln("directory saved")
		folderPathLabel.SetText(dir)
	})

	return container.NewVBox(label, widget.NewSeparator(), folderPathLabel, chooseDirButton), nil
}

// getSyncView builds the view for choosing frequency of sync for LMS files.
// It is placed in the Preferences tab.
func getSyncView(parentWindow fyne.Window, frequency int) (fyne.CanvasObject, error) {
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
		var newFrequency int
		switch s {
		case appConstants.SYNC_FREQUENCY_DISABLED:
			newFrequency = -1
		case appConstants.SYNC_FREQUENCY_ONE_HOUR:
			newFrequency = 1
		case appConstants.SYNC_FREQUENCY_TWO_HOUR:
			newFrequency = 2
		case appConstants.SYNC_FREQUENCY_FOUR_HOUR:
			newFrequency = 4
		case appConstants.SYNC_FREQUENCY_SIX_HOUR:
			newFrequency = 6
		case appConstants.SYNC_FREQUENCY_TWELVE_HOUR:
			newFrequency = 12
		default:
			newFrequency = 1
		}

		logs.Logger.Debugf("frequency selected - %d", newFrequency)

		savePrefErr := appPref.SaveSyncFrequency(newFrequency)
		if savePrefErr != nil {
			dialog.NewInformation(
				appConstants.APP_NAME,
				appConstants.PREFERENCES_FAILED_MESSAGE,
				parentWindow,
			).Show()
			logs.Logger.Errorln(savePrefErr)
			return
		}
		logs.Logger.Debugln("frequency saved")
	})
	frequencySelect.Selected = frequencyMap[frequency]

	return container.NewVBox(label, widget.NewSeparator(), description, frequencySelect), nil
}

// getAdvancedView builds the view for advanced options such as debug mode.
// It is placed in the Preferences tab.
func getAdvancedView(parentWindow fyne.Window, logLevel string) (fyne.CanvasObject, error) {
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
				filepath.FromSlash(
					fmt.Sprintf("file://%s", filepath.Join(baseDir, appConstants.LOG_FILE_NAME)),
				),
			),
		)
	}

	description.Wrapping = fyne.TextWrapWord

	debugCheckbox := widget.NewCheck(appConstants.DEBUG_CHECKBOX_TITLE, func(onDebug bool) {
		var newLogLevel string

		if onDebug {
			newLogLevel = "debug"
		} else {
			newLogLevel = "info"
		}

		logs.SetLogLevel(newLogLevel)
		logs.Logger.Debugf("debug mode changed to - %v", onDebug)

		savePrefErr := appPref.SaveDebugMode(newLogLevel)
		if savePrefErr != nil {
			dialog.NewInformation(
				appConstants.APP_NAME,
				appConstants.PREFERENCES_FAILED_MESSAGE,
				parentWindow,
			).Show()
			logs.Logger.Errorln(savePrefErr)
			return
		}

		dialog.NewInformation(
			appConstants.APP_NAME,
			appConstants.DEBUG_TOGGLE_SUCCESSFUL_MESSAGE,
			parentWindow,
		).Show()
	})

	debugCheckbox.Checked = logLevel == "debug"

	return container.NewVBox(label, widget.NewSeparator(), description, debugCheckbox), nil
}
