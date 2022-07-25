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
		tokensPath, credsErr := appAuth.GetTokensPath()
		if credsErr != nil {
			notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Credentials path load failed"}
			logs.Logger.Errorln(credsErr)
			return
		}
		tokensData, tokensErr := auth.LoadTokensData(tokensPath)
		if tokensErr != nil {
			notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: tokensErr.Error()}
			logs.Logger.Errorln(tokensErr)
			return
		}

		logs.Logger.Debugln("building - module request")
		modules := []api.Module{}

		canvasModulesRequest, canvasModulesReqErr := api.BuildModulesRequest(tokensData.CanvasToken.CanvasApiToken, constants.Canvas)
		if canvasModulesReqErr != nil {
			notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: canvasModulesReqErr.Error()}
			logs.Logger.Errorln(canvasModulesReqErr)
			return
		}
		canvasModules, canvasModErr := canvasModulesRequest.GetModules()
		if canvasModErr != nil {
			notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: canvasModErr.Error()}
			logs.Logger.Errorln(canvasModErr)
			return
		}
		modules = append(modules, canvasModules...)

		luminusModulesRequest, luminusModulesReqErr := api.BuildModulesRequest(tokensData.LuminusToken.JwtToken, constants.Luminus)
		if luminusModulesReqErr != nil {
			notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: luminusModulesReqErr.Error()}
			logs.Logger.Errorln(luminusModulesReqErr)
			return
		}
		luminusModules, luminusModErr := luminusModulesRequest.GetModules()
		if luminusModErr != nil {
			notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: luminusModErr.Error()}
			logs.Logger.Errorln(luminusModErr)
			return
		}
		modules = append(modules, luminusModules...)

		if preferences.Directory != "" {
			folders := []api.Folder{}
			for _, module := range modules {
				canvasFoldersReq, canvasFoldersReqErr := api.BuildFoldersRequest(tokensData.CanvasToken.CanvasApiToken, constants.Canvas, module)
				if canvasFoldersReqErr != nil {
					notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: canvasFoldersReqErr.Error()}
					logs.Logger.Errorln(canvasFoldersReqErr)
					return
				}
				canvasFolders, canvasFoldersErr := canvasFoldersReq.GetFolders()
				if canvasFoldersErr != nil {
					notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: canvasFoldersErr.Error()}
					logs.Logger.Errorln(canvasFoldersErr)
					return
				}
				folders = append(folders, canvasFolders...)

				luminusFoldersReq, luminusFoldersReqErr := api.BuildFoldersRequest(tokensData.LuminusToken.JwtToken, constants.Luminus, module)
				if luminusFoldersReqErr != nil {
					notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: luminusFoldersReqErr.Error()}
					logs.Logger.Errorln(luminusFoldersReqErr)
					return
				}
				luminusFolders, luminusFoldersErr := luminusFoldersReq.GetFolders()
				if luminusFoldersErr != nil {
					notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: luminusFoldersErr.Error()}
					logs.Logger.Errorln(luminusFoldersErr)
					return
				}
				folders = append(folders, luminusFolders...)
			}

			logs.Logger.Debugln("building - index map")
			currentFiles, currentFilesErr := indexing.Build(preferences.Directory)
			if currentFilesErr != nil {
				notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Unable to sync files"}
				logs.Logger.Errorln(currentFilesErr)
				return
			}

			nFilesToUpdate := 0
			filesUpdated := []api.File{}
			files := []api.File{}

			for _, folder := range folders {
				canvasFilesReq, canvasFilesReqErr := api.BuildFilesRequest(tokensData.CanvasToken.CanvasApiToken, constants.Canvas, folder)
				if canvasFilesReqErr != nil {
					notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: canvasFilesReqErr.Error()}
					logs.Logger.Errorln(canvasFilesReqErr)
					return
				}
				canvasFiles, canvasFilesErr := canvasFilesReq.GetFiles()
				if canvasFilesErr != nil {
					notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: canvasFilesErr.Error()}
					logs.Logger.Errorln(canvasFilesErr)
					return
				}
				files = append(files, canvasFiles...)

				luminusFilesReq, luminusFilesReqErr := api.BuildFilesRequest(tokensData.LuminusToken.JwtToken, constants.Luminus, folder)
				if luminusFilesReqErr != nil {
					notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: luminusFilesReqErr.Error()}
					logs.Logger.Errorln(luminusFilesReqErr)
					return
				}
				luminusFiles, luminusFilesErr := luminusFilesReq.GetFiles()
				if luminusFilesErr != nil {
					notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: luminusFilesErr.Error()}
					logs.Logger.Errorln(luminusFilesErr)
					return
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
		}

		var lastSync time.Time
		baseDir, _ := appDir.GetBaseDir()

		existingGradeErr := appFiles.DecodeStructFromFile(filepath.Join(baseDir, lominus.GRADES_FILE_NAME), &lastSync)
		if existingGradeErr != nil {
			logs.Logger.Debugln(existingGradeErr)
		}

		for _, module := range modules {
			logs.Logger.Debugln("building - grade request")
			gradeRequest, gradeReqErr := api.BuildGradeRequest(module)
			if gradeReqErr != nil {
				notifications.NotificationChannel <- notifications.Notification{Title: "Grades", Content: "Unable to retrieve grades"}
				logs.Logger.Errorln(gradeReqErr)
				continue
			}

			logs.Logger.Debugln("retrieving - grade")
			allGrades, gradesErr := gradeRequest.GetGrades()
			if gradesErr != nil {
				notifications.NotificationChannel <- notifications.Notification{Title: "Grades", Content: "Unable to retrieve grades"}
				logs.Logger.Errorln(gradesErr)
				continue
			}

			for _, grade := range allGrades {
				if time.Unix(grade.LastUpdated, 0).Before(lastSync) {
					continue
				}

				if telegramInfoErr == nil {
					message := telegram.GenerateGradeMessageFormat(grade)
					gradeMsgErr := telegram.SendMessage(telegramInfo.BotApi, telegramInfo.UserId, message)

					if gradeMsgErr != nil {
						logs.Logger.Errorln(gradeMsgErr)
						continue
					}
				}
			}
		}

		err := appFiles.EncodeStructToFile(filepath.Join(baseDir, lominus.GRADES_FILE_NAME), time.Now())
		if err != nil {
			logs.Logger.Debugln(err)
		}
	})
}
