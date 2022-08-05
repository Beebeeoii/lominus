// Package cron provides primitives to initialise and control the main cron scheduler.
package cron

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	appApp "github.com/beebeeoii/lominus/internal/app"
	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	intTelegram "github.com/beebeeoii/lominus/internal/app/integrations/telegram"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	appFiles "github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/internal/indexing"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/lominus"
	"github.com/beebeeoii/lominus/internal/notifications"
	"github.com/beebeeoii/lominus/pkg/api"
	"github.com/beebeeoii/lominus/pkg/auth"
	"github.com/beebeeoii/lominus/pkg/constants"
	"github.com/beebeeoii/lominus/pkg/integrations/telegram"

	"github.com/go-co-op/gocron"
)

var mainScheduler *gocron.Scheduler
var mainJob *gocron.Job
var LastRanChannel chan string

// Init initialises the cronjob with the desired frequency set by the user.
// If frequency is unset, cronjob is not initialised.
func Init() error {
	mainScheduler = gocron.NewScheduler(time.Local)
	LastRanChannel = make(chan string)

	preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
	if getPreferencesPathErr != nil {
		return getPreferencesPathErr
	}

	preferences, loadPrefErr := appPref.LoadPreferences(preferencesPath)
	if loadPrefErr != nil {
		return loadPrefErr
	}

	if preferences.Frequency == -1 {
		return nil
	}

	job, err := createJob(preferences.Frequency)
	if err != nil {
		return err
	}

	mainJob = job
	mainScheduler.StartAsync()

	return nil
}

// Rerun clears the job from the scheduler and reschedules the same job with the new frequency.
func Rerun(frequency int) error {
	mainScheduler.Clear()

	if frequency == -1 {
		return nil
	}

	job, err := createJob(frequency)
	if err != nil {
		return err
	}

	mainJob = job
	mainScheduler.StartAsync()

	return nil
}

// GetNextRun returns the next time the cronjob would run.
func GetNextRun() time.Time {
	return mainJob.NextRun()
}

// GetLastRan returns the last time the cronjob ran.
func GetLastRan() time.Time {
	return mainJob.LastRun()
}

// createJob creates the cronjob that would run at the given frequency.
// It returns a Job which can be used in the main scheduler.
// This is where the bulk of the syncing logic lives.
//
// TODO Cleanup notifications - make it more user friendly. No point
// putting technical logs in notifications.
func createJob(frequency int) (*gocron.Job, error) {
	return mainScheduler.Every(frequency).Hours().Do(func() {
		logs.Logger.Infof("job started: %s", time.Now().Format(time.RFC3339))
		if appApp.GetOs() == "windows" {
			LastRanChannel <- GetLastRan().Format("2 Jan 15:04:05")
		}

		logs.Logger.Debugln("retrieving - preferences path")
		preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
		if getPreferencesPathErr != nil {
			logs.Logger.Errorln(getPreferencesPathErr)
			return
		}

		logs.Logger.Debugln("loading - preferences")
		preferences, loadPrefErr := appPref.LoadPreferences(preferencesPath)
		if loadPrefErr != nil {
			logs.Logger.Errorln(loadPrefErr)
			return
		}

		logs.Logger.Debugln("retrieving - telegram path")
		telegramInfoPath, getTelegramInfoPathErr := intTelegram.GetTelegramInfoPath()
		if getTelegramInfoPathErr != nil {
			logs.Logger.Errorln(getTelegramInfoPathErr)
			return
		}

		logs.Logger.Debugln("loading - telegram")
		telegramInfo, telegramInfoErr := telegram.LoadTelegramData(telegramInfoPath)

		logs.Logger.Debugln("loading - credentials and tokens")
		tokensData, tokensErr := loadTokensData()
		if tokensErr != nil {
			notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: tokensErr.Error()}
			logs.Logger.Errorln(tokensErr)
			return
		}

		logs.Logger.Debugln("building - module request")

		canvasModules, canvasModErr := getModules(tokensData.CanvasToken.CanvasApiToken, constants.Canvas)
		if canvasModErr != nil {
			// TODO Somehow collate this error and display to user at the end
			// notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: canvasModErr.Error()}
			logs.Logger.Errorln(canvasModErr)
		}

		luminusModules, luninusModErr := getModules(tokensData.LuminusToken.JwtToken, constants.Luminus)
		if luninusModErr != nil {
			// TODO Somehow collate this error and display to user at the end
			// notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: luninusModErr.Error()}
			logs.Logger.Errorln(luninusModErr)
		}

		// Note that grades is currently only supported for Luminus
		grades, gradesErr := getGrades(luminusModules)
		if gradesErr != nil {
			logs.Logger.Errorln(gradesErr)
		}

		if telegramInfoErr != nil {
			for _, grade := range grades {
				message := telegram.GenerateGradeMessageFormat(grade)
				sendErr := telegram.SendMessage(telegramInfo.BotApi, telegramInfo.UserId, message)
				if sendErr != nil {
					logs.Logger.Errorln(sendErr)
				}
			}
		}

		// If directory for file sync is not set, exit from job.
		if preferences.Directory == "" {
			return
		}

		logs.Logger.Debugln("building - index map")
		currentFiles, currentFilesErr := indexing.Build(preferences.Directory)
		if currentFilesErr != nil {
			notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Failed to get current downloaded files"}
			logs.Logger.Errorln(currentFilesErr)
			return
		}

		canvasFolders := []api.Folder{}
		for _, module := range canvasModules {
			// TODO Check if it is even possible for files to be in module's root folder
			folders, canvasFoldersErr := getFolders(tokensData.CanvasToken.CanvasApiToken, constants.Canvas, module)
			if canvasFoldersErr != nil {
				// TODO Somehow collate this error and display to user at the end
				// notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: canvasFoldersErr.Error()}
				logs.Logger.Errorln(canvasFoldersErr)
			}
			canvasFolders = append(canvasFolders, folders...)
		}

		luminusFolders := []api.Folder{}
		for _, module := range luminusModules {
			// This ensures that files in the module's root folder are downloaded as well.
			moduleMainFolder := api.Folder{
				Id:           module.Id,
				Name:         module.Name,
				Downloadable: module.IsAccessible,
				HasSubFolder: true,       // doesn't matter
				Ancestors:    []string{}, // main folder does not have any ancestors
			}
			folders, luminusFoldersErr := getFolders(tokensData.LuminusToken.JwtToken, constants.Luminus, module)
			if luminusFoldersErr != nil {
				// TODO Somehow collate this error and display to user at the end
				// notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: luminusFoldersErr.Error()}
				logs.Logger.Errorln(luminusFoldersErr)
			}
			luminusFolders = append(luminusFolders, moduleMainFolder)
			luminusFolders = append(luminusFolders, folders...)
		}

		nFilesToUpdate := 0
		filesUpdated := []api.File{}

		files := []api.File{}
		for _, folder := range canvasFolders {
			canvasFiles, canvasFilesErr := getFiles(tokensData.CanvasToken.CanvasApiToken, constants.Canvas, folder)
			if canvasFilesErr != nil {
				// TODO Somehow collate this error and display to user at the end
				// notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: canvasFilesErr.Error()}
				logs.Logger.Errorln(canvasFilesErr)
			}
			files = append(files, canvasFiles...)

		}

		for _, folder := range luminusFolders {
			luminusFiles, luminusFilesErr := getFiles(tokensData.LuminusToken.JwtToken, constants.Luminus, folder)
			if luminusFilesErr != nil {
				// TODO Somehow collate this error and display to user at the end
				// notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: luminusFilesErr.Error()}
				logs.Logger.Errorln(luminusFilesErr)
			}
			files = append(files, luminusFiles...)
		}

		for _, file := range files {
			key := fmt.Sprintf("%s/%s", strings.Join(file.Ancestors, "/"), file.Name)
			localLastUpdated := currentFiles[key].LastUpdated
			platformLastUpdated := file.LastUpdated

			if _, exists := currentFiles[key]; !exists || localLastUpdated.Before(platformLastUpdated) {
				nFilesToUpdate += 1

				logs.Logger.Debugf("downloading - %s [%s vs %s]", key, localLastUpdated.String(), platformLastUpdated.String())
				fileDirSlice := append([]string{preferences.Directory}, file.Ancestors...)
				filePath := filepath.Join(fileDirSlice...)
				appFiles.EnsureDir(filePath)
				downloadErr := file.Download(filePath)
				if downloadErr != nil {
					notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: fmt.Sprintf("Unable to download file: %s", file.Name)}
					logs.Logger.Errorln(downloadErr)
					continue
				}
				filesUpdated = append(filesUpdated, file)
			}
		}

		if nFilesToUpdate > 0 {
			nFilesUpdated := len(filesUpdated)
			updatedFilesModulesNames := []string{}

			for _, file := range filesUpdated {
				if telegramInfoErr == nil {
					message := telegram.GenerateFileUpdatedMessageFormat(file)
					gradeMsgErr := telegram.SendMessage(telegramInfo.BotApi, telegramInfo.UserId, message)

					if gradeMsgErr != nil {
						logs.Logger.Errorln(gradeMsgErr)
						continue
					}
				}

				updatedFilesModulesNames = append(updatedFilesModulesNames, fmt.Sprintf("[%s] %s ", file.Ancestors[0], file.Name))
			}

			var updatedFileNamesString string

			if nFilesUpdated > 4 {
				updatedFileNamesString = strings.Join(append(updatedFilesModulesNames[:3], "..."), "\n")
			} else {
				updatedFileNamesString = strings.Join(updatedFilesModulesNames, "\n")
			}

			notifications.NotificationChannel <- notifications.Notification{
				Title:   fmt.Sprintf("Sync: %d/%d updated", nFilesUpdated, nFilesToUpdate),
				Content: updatedFileNamesString,
			}
		} else {
			notifications.NotificationChannel <- notifications.Notification{
				Title:   "Sync",
				Content: "Your files are up to date",
			}
		}

		logs.Logger.Infof("job completed: %s", time.Now().Format(time.RFC3339))
	})
}

func loadTokensData() (auth.TokensData, error) {
	var tokensData auth.TokensData

	tokensPath, credsErr := appAuth.GetTokensPath()
	if credsErr != nil {
		return tokensData, credsErr
	}

	tokensData, tokensErr := auth.LoadTokensData(tokensPath, true)
	if tokensErr != nil {
		return tokensData, tokensErr
	}

	return tokensData, nil
}

func getModules(token string, platform constants.Platform) ([]api.Module, error) {
	modules := []api.Module{}

	modulesRequest, modulesReqErr := api.BuildModulesRequest(token, platform)
	if modulesReqErr != nil {
		return modules, modulesReqErr
	}

	modules, modulesErr := modulesRequest.GetModules()
	if modulesErr != nil {
		return modules, modulesErr
	}

	return modules, nil
}

func getFolders(token string, platform constants.Platform, module api.Module) ([]api.Folder, error) {
	folders := []api.Folder{}

	foldersReq, foldersReqErr := api.BuildFoldersRequest(token, platform, module)
	if foldersReqErr != nil {
		return folders, foldersReqErr
	}

	folders, foldersErr := foldersReq.GetFolders()
	if foldersErr != nil {
		return folders, foldersErr
	}

	return folders, nil
}

func getFiles(token string, platform constants.Platform, folder api.Folder) ([]api.File, error) {
	files := []api.File{}

	filesReq, filesReqErr := api.BuildFilesRequest(token, platform, folder)
	if filesReqErr != nil {
		return files, filesReqErr
	}

	files, filesErr := filesReq.GetFiles()
	if filesErr != nil {
		return files, filesErr
	}

	return files, nil
}

func getGrades(modules []api.Module) ([]api.Grade, error) {
	grades := []api.Grade{}

	var lastSync time.Time
	baseDir, _ := appDir.GetBaseDir()

	existingGradeErr := appFiles.DecodeStructFromFile(filepath.Join(baseDir, lominus.GRADES_FILE_NAME), &lastSync)
	if existingGradeErr != nil {
		return grades, existingGradeErr
	}

	for _, module := range modules {
		logs.Logger.Debugln("building - grade request")
		gradeRequest, gradeReqErr := api.BuildGradeRequest(module)
		if gradeReqErr != nil {
			logs.Logger.Errorln(gradeReqErr)
			continue
		}

		logs.Logger.Debugln("retrieving - grade")
		allGrades, gradesErr := gradeRequest.GetGrades()
		if gradesErr != nil {
			logs.Logger.Errorln(gradesErr)
			continue
		}

		grades = append(grades, allGrades...)
	}

	err := appFiles.EncodeStructToFile(filepath.Join(baseDir, lominus.GRADES_FILE_NAME), time.Now())
	if err != nil {
		return []api.Grade{}, err
	}

	return grades, nil
}
