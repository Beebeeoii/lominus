package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/beebeeoii/lominus/internal/lominus"

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

// getPreferencesTab builds the preferences tab in the main UI.
func getPreferencesTab(parentWindow fyne.Window) (*container.TabItem, error) {
	logs.Logger.Debugln("preferences tab loaded")
	tab := container.NewTabItem(appConstants.PREFERENCES_TITLE, container.NewVBox())
	// debugCheckbox := widget.NewCheck("Debug Mode", func(onDebug bool) {
	// 	preferences := getPreferences()
	// 	preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
	// 	if getPreferencesPathErr != nil {
	// 		dialog.NewInformation(lominus.APP_NAME, "An error has occurred :( Please try again", parentWindow).Show()
	// 		logs.Logger.Errorln(getPreferencesPathErr)
	// 		return
	// 	}

	// 	if onDebug {
	// 		preferences.LogLevel = "debug"
	// 	} else {
	// 		preferences.LogLevel = "info"
	// 	}

	// 	logs.SetLogLevel(preferences.LogLevel)
	// 	logs.Logger.Debugf("debug mode changed to - %v", onDebug)

	// 	savePrefErr := appPref.SavePreferences(preferencesPath, preferences)
	// 	if savePrefErr != nil {
	// 		dialog.NewInformation(lominus.APP_NAME, "An error has occurred :( Please try again", parentWindow).Show()
	// 		logs.Logger.Errorln(savePrefErr)
	// 		return
	// 	}

	// 	dialog.NewInformation(lominus.APP_NAME, "Please restart Lominus for changes to take place.", parentWindow).Show()
	// })

	// debugCheckbox.Checked = getPreferences().LogLevel == "debug"

	fileDirectorySubTab, fileDirectorySubTabErr := getFileDirectorySubTab(w)
	if fileDirectorySubTabErr != nil {
		return tab, fileDirectorySubTabErr
	}

	syncSubTab, syncSubTabErr := getSyncSubTab(w)
	if syncSubTabErr != nil {
		return tab, syncSubTabErr
	}

	tabsContainer := container.NewAppTabs(fileDirectorySubTab, syncSubTab)
	tab.Content = container.NewVBox(tabsContainer)

	return tab, nil
}

func getFileDirectorySubTab(parentWindow fyne.Window) (*container.TabItem, error) {
	logs.Logger.Debugln("file directory tab loaded")
	tab := container.NewTabItem(appConstants.FILE_DIRECTORY_TAB_TITLE, container.NewVBox())

	description := widget.NewRichTextFromMarkdown(appConstants.FILE_DIRECTORY_TAB_DESCRIPTION)
	description.Wrapping = fyne.TextWrapWord

	dir := getPreferences().Directory
	if dir == "" {
		dir = appConstants.FILE_DIRECTORY_FOLDER_PATH_DEFAULT
	}

	folderPathLabel := widget.NewLabel(dir)
	folderPathLabel.Wrapping = fyne.TextWrapWord
	chooseDirButton := widget.NewButton(appConstants.FILE_DIRECTORY_SELECT_DIRECTORY_TEXT, func() {
		dir, dirErr := fileDialog.Directory().Title(
			appConstants.FILE_DIRECTORY_SELECT_DIRECTORY_TEXT,
		).Browse()

		if dirErr != nil {
			if dirErr.Error() != "Cancelled" {
				logs.Logger.Debugln("directory selection cancelled")
				dialog.NewInformation(
					lominus.APP_NAME,
					appConstants.SAVE_PREFERENCES_FAILED_MESSAGE,
					parentWindow,
				).Show()
				logs.Logger.Errorln(dirErr)
			}
			return
		}
		logs.Logger.Debugf("directory chosen - %s", dir)

		preferences := getPreferences()
		preferences.Directory = dir

		preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
		if getPreferencesPathErr != nil {
			dialog.NewInformation(
				lominus.APP_NAME,
				appConstants.SAVE_PREFERENCES_FAILED_MESSAGE,
				parentWindow,
			).Show()
			logs.Logger.Errorln(getPreferencesPathErr)
			return
		}

		savePrefErr := appPref.SavePreferences(preferencesPath, preferences)
		if savePrefErr != nil {
			dialog.NewInformation(
				lominus.APP_NAME,
				appConstants.SAVE_PREFERENCES_FAILED_MESSAGE,
				parentWindow,
			).Show()
			logs.Logger.Errorln(savePrefErr)
			return
		}
		logs.Logger.Debugln("directory saved")
		folderPathLabel.SetText(preferences.Directory)
	})

	tab.Content = container.NewVBox(description, folderPathLabel, chooseDirButton)

	return tab, nil
}

func getSyncSubTab(parentWindow fyne.Window) (*container.TabItem, error) {
	logs.Logger.Debugln("sync tab loaded")
	tab := container.NewTabItem(appConstants.SYNC_TAB_TITLE, container.NewVBox())

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
				appConstants.SAVE_PREFERENCES_FAILED_MESSAGE,
				parentWindow,
			).Show()
			logs.Logger.Errorln(getPreferencesPathErr)
			return
		}

		savePrefErr := appPref.SavePreferences(preferencesPath, preferences)
		if savePrefErr != nil {
			dialog.NewInformation(
				lominus.APP_NAME,
				appConstants.SAVE_PREFERENCES_FAILED_MESSAGE,
				parentWindow,
			).Show()
			logs.Logger.Errorln(savePrefErr)
			return
		}
		logs.Logger.Debugln("frequency saved")
	})
	frequencySelect.Selected = frequencyMap[getPreferences().Frequency]

	tab.Content = container.NewVBox(description, frequencySelect)

	return tab, nil
}
