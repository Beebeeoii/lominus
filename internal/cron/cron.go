// Package cron provides primitives to initialise and control the main cron scheduler.
package cron

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	appFiles "github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/internal/indexing"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/notifications"
	"github.com/beebeeoii/lominus/pkg/api"
	"github.com/beebeeoii/lominus/pkg/constants"
	"github.com/beebeeoii/lominus/pkg/integrations/telegram"
	"github.com/boltdb/bolt"

	"github.com/go-co-op/gocron"
)

var mainScheduler *gocron.Scheduler
var mainJob *gocron.Job

// Init initialises the cronjob with the desired frequency set by the user.
// If frequency is unset, cronjob is not initialised.
func Init() error {
	mainScheduler = gocron.NewScheduler(time.Local)

	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		return retrieveBaseDirErr
	}

	dbFName := filepath.Join(baseDir, appConstants.DATABASE_FILE_NAME)
	db, dbErr := bolt.Open(dbFName, 0600, &bolt.Options{ReadOnly: true})

	if dbErr != nil {
		return dbErr
	}

	tx, _ := db.Begin(false)
	prefBucket := tx.Bucket([]byte("Preferences"))
	directory := string(prefBucket.Get([]byte("directory")))
	frequency, _ := strconv.Atoi(string(prefBucket.Get([]byte("frequency"))))

	defer db.Close()

	if frequency == -1 {
		return nil
	}

	job, err := createJob(directory, frequency)
	if err != nil {
		return err
	}

	mainJob = job
	mainScheduler.StartAsync()

	return nil
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
		logs.Logger.Infof("job started: %s", time.Now().Format(time.RFC3339))

		// If directory for file sync is not set, exit from job.
		if rootSyncDirectory == "" {
			logs.Logger.Infoln("Root sync directory not set. Exiting from cron job !")
			return
		}

		logs.Logger.Infoln("trying to access database in Read Only for credentials")

		baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
		if retrieveBaseDirErr != nil {
			logs.Logger.Errorln(retrieveBaseDirErr)
			return
		}

		dbFName := filepath.Join(baseDir, appConstants.DATABASE_FILE_NAME)
		db, dbErr := bolt.Open(dbFName, 0600, &bolt.Options{ReadOnly: true})

		if dbErr != nil {
			logs.Logger.Errorln(dbErr)
			return
		}

		logs.Logger.Infoln("database access: successful")

		tx, _ := db.Begin(false)
		canvasToken := string(tx.Bucket([]byte("Auth")).Get([]byte("canvasToken")))
		telegramUserId := string(tx.Bucket([]byte("Integrations")).Get([]byte("telegramUserId")))
		telegramBotId := string(tx.Bucket([]byte("Integrations")).Get([]byte("telegramBotId")))

		if telegramUserId == "" {
			logs.Logger.Debugln("telegram userId - NOT FOUND")
		} else {
			logs.Logger.Debugln("telegram userId - " + telegramUserId)
		}

		if telegramBotId == "" {
			logs.Logger.Debugln("telegram botId - NOT FOUND")
		} else {
			logs.Logger.Debugln("telegram botId - " + telegramBotId)
		}

		tx.Commit()
		db.Close()

		logs.Logger.Debugln("building - module request")

		canvasModules, canvasModErr := getModules(canvasToken, constants.Canvas)
		if canvasModErr != nil {
			// TODO Somehow collate this error and display to user at the end
			// notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: canvasModErr.Error()}
			logs.Logger.Errorln(canvasModErr)
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
				canvasToken,
				module,
			)

			if moduleFolderReqErr != nil {
				logs.Logger.Errorln(moduleFolderReqErr)
			}

			moduleFolder, moduleFolderErr := moduleFolderReq.GetModuleFolder()
			if moduleFolderErr != nil {
				logs.Logger.Errorln(moduleFolderErr)
			}

			foldersReq, foldersReqErr := api.BuildFoldersRequest(
				canvasToken,
				constants.Canvas,
				moduleFolder,
			)
			if foldersReqErr != nil {
				logs.Logger.Errorln(foldersReqErr)
			}

			files, foldersErr := foldersReq.GetRootFiles()
			if foldersErr != nil {
				fmt.Println("testsets")
				logs.Logger.Errorln(foldersErr)
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
					logs.Logger.Errorln(downloadErr)
					continue
				}
				filesUpdated = append(filesUpdated, file)
			}
		}

		if nFilesToUpdate > 0 && telegramUserId != "" && telegramBotId != "" {
			nFilesUpdated := len(filesUpdated)
			updatedFilesModulesNames := []string{}

			// TODO Send one message per module instead of one message per file as there can be many files
			for _, file := range filesUpdated {
				message := telegram.GenerateFileUpdatedMessageFormat(file)
				gradeMsgErr := telegram.SendMessage(telegramBotId, telegramUserId, message)

				if gradeMsgErr != nil {
					logs.Logger.Errorln(gradeMsgErr)
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
