// Package cron provides primitives to initialise and control the main cron scheduler.
package cron

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	appApp "github.com/beebeeoii/lominus/internal/app"
	intTelegram "github.com/beebeeoii/lominus/internal/app/integrations/telegram"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	files "github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/internal/indexing"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/notifications"
	"github.com/beebeeoii/lominus/pkg/api"
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
		notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Syncing in progress"}
		logs.Logger.Infof("job started: %s", time.Now().Format(time.RFC3339))
		if appApp.GetOs() == "windows" {
			LastRanChannel <- GetLastRan().Format("2 Jan 15:04:05")
		}

		preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
		if getPreferencesPathErr != nil {
			logs.Logger.Errorln(getPreferencesPathErr)
			return
		}

		preferences, loadPrefErr := appPref.LoadPreferences(preferencesPath)
		if loadPrefErr != nil {
			logs.Logger.Errorln(loadPrefErr)
			return
		}

		moduleRequest, modReqErr := api.BuildModuleRequest()
		if modReqErr != nil {
			notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Authentication failed"}
			logs.Logger.Errorln(modReqErr)
			return
		}

		modules, modErr := moduleRequest.GetModules()
		if modErr != nil {
			notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Unable to retrieve modules"}
			logs.Logger.Errorln(modErr)
			return
		}

		if preferences.Directory != "" {
			updatedFiles := make([]api.File, 0)
			for _, module := range modules {
				fileRequest, fileReqErr := api.BuildDocumentRequest(module, api.GET_ALL_FILES)
				if fileReqErr != nil {
					notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Unable to retrieve files"}
					logs.Logger.Errorln(fileReqErr)
					continue
				}

				files, fileErr := fileRequest.GetRootFiles()
				if fileErr != nil {
					notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Unable to retrieve files"}
					logs.Logger.Errorln(fileErr)
					continue
				}

				updatedFiles = append(updatedFiles, files...)
			}

			indexMapEntries := make([]indexing.IndexMapEntry, 0)
			for _, file := range updatedFiles {
				indexMapEntries = append(indexMapEntries, indexing.IndexMapEntry{
					Id:          file.Id,
					FileName:    file.Name,
					LastUpdated: file.LastUpdated.Unix(),
				})
			}

			indexing.CreateIndexMap(indexing.IndexMap{
				Entries: indexMapEntries,
			})

			currentFiles, currentFilesErr := indexing.Build(preferences.Directory)
			if currentFilesErr != nil {
				notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Unable to sync files"}
				logs.Logger.Errorln(currentFilesErr)
				return
			}

			nFilesToUpdate := 0
			filesUpdated := []api.File{}

			for _, file := range updatedFiles {
				if _, exists := currentFiles[file.Name]; !exists || currentFiles[file.Name].LastUpdated.Before(file.LastUpdated) {
					nFilesToUpdate += 1
					downloadErr := downloadFile(preferences.Directory, file)
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

		telegramInfoPath, getTelegramInfoPathErr := intTelegram.GetTelegramInfoPath()
		if getTelegramInfoPathErr != nil {
			logs.Logger.Errorln(getTelegramInfoPathErr)
			return
		}

		telegramInfo, telegramInfoErr := telegram.LoadTelegramData(telegramInfoPath)
		if telegramInfoErr != nil {
			logs.Logger.Errorln(telegramInfoErr)
			return
		}

		for _, module := range modules {
			gradeRequest, gradeReqErr := api.BuildGradeRequest(module)
			if gradeReqErr != nil {
				notifications.NotificationChannel <- notifications.Notification{Title: "Grades", Content: "Unable to retrieve grades"}
				logs.Logger.Errorln(gradeReqErr)
				continue
			}

			grades, gradesErr := gradeRequest.GetGrades()
			if gradesErr != nil {
				notifications.NotificationChannel <- notifications.Notification{Title: "Grades", Content: "Unable to retrieve grades"}
				logs.Logger.Errorln(gradesErr)
				continue
			}

			for _, grade := range grades {
				frequencyDuration, freqErr := time.ParseDuration(fmt.Sprintf("%dh", preferences.Frequency))
				if freqErr != nil {
					logs.Logger.Errorln(freqErr)
					continue
				}

				if time.Since(time.Unix(grade.LastUpdated, 0)) > frequencyDuration {
					continue
				}

				gradeMsgErr := telegram.SendMessage(telegramInfo.BotApi, telegramInfo.UserId, telegram.GenerateGradeMessageFormat(module.ModuleCode, grade.Name, grade.Comments, grade.Marks, grade.MaxMarks))

				if gradeMsgErr != nil {
					logs.Logger.Errorln(gradeMsgErr)
					continue
				}
			}
		}
	})
}

// downloadFile is a helper function to download the respective files into their corresponding
// directory based on the File's Ancestors.
func downloadFile(baseDir string, file api.File) error {
	fileDirSlice := append([]string{baseDir}, file.Ancestors...)
	files.EnsureDir(filepath.Join(append(fileDirSlice, file.Name)...))

	downloadReq, dlReqErr := api.BuildDocumentRequest(file, api.DOWNLOAD_FILE)
	if dlReqErr != nil {
		return dlReqErr
	}

	return downloadReq.Download(filepath.Join(fileDirSlice...))
}
