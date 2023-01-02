// Package cron provides primitives to initialise and control the main cron scheduler.
package cron

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
	appInt "github.com/beebeeoii/lominus/internal/app/integrations/telegram"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	appFiles "github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/internal/indexing"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/notifications"
	"github.com/beebeeoii/lominus/pkg/api"
	"github.com/beebeeoii/lominus/pkg/constants"
	"github.com/beebeeoii/lominus/pkg/integrations/telegram"

	"github.com/go-co-op/gocron"
)

var mainScheduler *gocron.Scheduler
var mainJob *gocron.Job

// Init initialises the cronjob with the desired frequency set by the user.
// If frequency is unset, cronjob is not initialised.
func Init() error {
	mainScheduler = gocron.NewScheduler(time.Local)

	pref, err := appPref.GetPreferences()

	if err != nil {
		return err
	}

	if pref.Frequency == -1 {
		return nil
	}

	job, err := createJob(pref.Directory, pref.Frequency)
	if err != nil {
		return err
	}

	mainJob = job
	mainScheduler.StartAsync()

	return err
}

// Rerun clears the job from the scheduler and reschedules the same job with the new frequency.
func Rerun(rootSyncDirectory string, frequency int) error {
	mainScheduler.Clear()

	if frequency == -1 {
		return nil
	}

	job, err := createJob(rootSyncDirectory, frequency)
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
func createJob(rootSyncDirectory string, frequency int) (*gocron.Job, error) {
	return mainScheduler.Every(frequency).Hours().Do(func() {
		notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Sync Started!"}

		logs.Logger.Infof("job started: %s", time.Now().Format(time.RFC3339))

		// If directory for file sync is not set, exit from job.
		if rootSyncDirectory == "" {
			logs.Logger.Infoln("Root sync directory not set. Exiting from cron job !")
			return
		}

		canvasCredentials, credErr := appAuth.GetCanvasCredentials()
		if credErr != nil {
			logs.Logger.Warnln(credErr)
		} else {
			logs.Logger.Infoln("canvasCredentials access: successful")
		}

		telegramIds, tIdsErr := appInt.GetTelegramIds()
		if tIdsErr != nil {
			logs.Logger.Warnln(tIdsErr)
		} else {
			logs.Logger.Infoln("telegramIds access: successful")
		}

		logs.Logger.Debugln("building - module request")

		canvasModules, canvasModErr := getModules(canvasCredentials.CanvasApiToken, constants.Canvas)
		if canvasModErr != nil {
			// TODO Somehow collate this error and display to user at the end
			// notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: canvasModErr.Error()}
			logs.Logger.Warnln(canvasModErr)
		}

		logs.Logger.Debugln("building - index map")
		currentFiles, currentFilesErr := indexing.Build(rootSyncDirectory)
		if currentFilesErr != nil {
			notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Failed to get current downloaded files"}
			logs.Logger.Errorln(currentFilesErr)
			return
		}

		lmsFiles := []api.File{}
		for _, module := range canvasModules {
			moduleFolderReq, moduleFolderReqErr := api.BuildModuleFolderRequest(
				canvasCredentials.CanvasApiToken,
				module,
			)

			if moduleFolderReqErr != nil {
				logs.Logger.Warnln(moduleFolderReqErr)
			}

			moduleFolder, moduleFolderErr := moduleFolderReq.GetModuleFolder()
			if moduleFolderErr != nil {
				logs.Logger.Warnln(moduleFolderErr)
			}

			foldersReq, foldersReqErr := api.BuildFoldersRequest(
				canvasCredentials.CanvasApiToken,
				constants.Canvas,
				moduleFolder,
			)
			if foldersReqErr != nil {
				logs.Logger.Warnln(foldersReqErr)
			}

			files, foldersErr := foldersReq.GetRootFiles()
			if foldersErr != nil {
				logs.Logger.Warnln(foldersErr)
			}

			lmsFiles = append(lmsFiles, files...)
		}

		nFilesToUpdate := 0
		filesUpdated := []api.File{}

		for _, file := range lmsFiles {
			key := fmt.Sprintf("%s/%s", strings.Join(file.Ancestors, "/"), file.Name)
			localLastUpdated := currentFiles[key].LastUpdated
			platformLastUpdated := file.LastUpdated

			if _, exists := currentFiles[key]; !exists || localLastUpdated.Before(platformLastUpdated) {
				nFilesToUpdate += 1

				logs.Logger.Debugf("downloading - %s [%s vs %s]", key, localLastUpdated.String(), platformLastUpdated.String())
				fileDirSlice := append([]string{rootSyncDirectory}, file.Ancestors...)
				filePath := filepath.Join(fileDirSlice...)
				appFiles.EnsureDir(filePath)
				downloadErr := file.Download(filePath)
				if downloadErr != nil {
					notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: fmt.Sprintf("Unable to download file: %s", file.Name)}
					logs.Logger.Warnln(downloadErr)
					continue
				}
				filesUpdated = append(filesUpdated, file)
			}
		}

		if nFilesToUpdate > 0 && telegramIds.UserId != "" && telegramIds.BotId != "" {
			nFilesUpdated := len(filesUpdated)
			updatedFilesModulesNames := []string{}

			// TODO Send one message per module instead of one message per file as there can be many files
			for _, file := range filesUpdated {
				message := telegram.GenerateFileUpdatedMessageFormat(file)
				gradeMsgErr := telegram.SendMessage(telegramIds.BotId, telegramIds.UserId, message)

				if gradeMsgErr != nil {
					logs.Logger.Warnln(gradeMsgErr)
					continue
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

// getModules is a helper function that retrieves Module objects based on the platform
// passed in the arguments.
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
