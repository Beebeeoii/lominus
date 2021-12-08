package cron

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	appApp "github.com/beebeeoii/lominus/internal/app"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	"github.com/beebeeoii/lominus/internal/indexing"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/internal/notifications"
	"github.com/beebeeoii/lominus/pkg/api"
	"github.com/beebeeoii/lominus/pkg/pref"

	"github.com/go-co-op/gocron"
)

var mainScheduler *gocron.Scheduler
var mainJob *gocron.Job
var LastRanChannel chan string

func Init() error {
	mainScheduler = gocron.NewScheduler(time.Local)
	LastRanChannel = make(chan string)

	preferences, loadPrefErr := pref.LoadPreferences(appPref.GetPreferencesPath())
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

func GetNextRun() time.Time {
	return mainJob.NextRun()
}

func GetLastRan() time.Time {
	return mainJob.LastRun()
}

func createJob(frequency int) (*gocron.Job, error) {
	return mainScheduler.Every(frequency).Hours().Do(func() {
		notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Syncing in progress"}
		logs.InfoLogger.Printf("job started: %s\n", time.Now().Format(time.RFC3339))
		if appApp.GetOs() == "windows" {
			LastRanChannel <- GetLastRan().Format("2 Jan 15:04:05")
		}

		preferences, prefErr := pref.LoadPreferences(appPref.GetPreferencesPath())
		if prefErr != nil {
			logs.WarningLogger.Println(prefErr)
			return
		}

		if preferences.Directory != "" {
			moduleRequest, modReqErr := api.BuildModuleRequest()
			if modReqErr != nil {
				notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Authentication failed"}
				logs.WarningLogger.Println(modReqErr)
				return
			}

			modules, modErr := moduleRequest.GetModules()
			if modErr != nil {
				notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Unable to retrieve modules"}
				logs.WarningLogger.Println(modErr)
				return
			}

			updatedFiles := make([]api.File, 0)
			for _, module := range modules {
				fileRequest, fileReqErr := api.BuildDocumentRequest(module, api.GET_ALL_FILES)
				if fileReqErr != nil {
					notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Unable to retrieve files"}
					logs.WarningLogger.Println(fileReqErr)
					continue
				}

				files, fileErr := fileRequest.GetAllFiles()
				if fileErr != nil {
					notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: "Unable to retrieve files"}
					logs.WarningLogger.Println(fileErr)
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
				logs.WarningLogger.Println(currentFilesErr)
				return
			}

			filesToUpdate := 0
			filesUpdated := 0
			for _, file := range updatedFiles {
				if _, exists := currentFiles[file.Name]; !exists || currentFiles[file.Name].LastUpdated.Before(file.LastUpdated) {
					filesToUpdate += 1
					downloadErr := downloadFile(preferences.Directory, file)
					if downloadErr != nil {
						notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: fmt.Sprintf("Unable to download file: %s", file.Name)}
						logs.ErrorLogger.Println(downloadErr)
						continue
					}
					filesUpdated += 1
				}
			}

			filesUpdatedNotificationContent := "Your files are up to date"
			if filesToUpdate > 0 {
				filesUpdatedNotificationContent = fmt.Sprintf("%d/%d files updated", filesToUpdate, filesToUpdate)
			}

			notifications.NotificationChannel <- notifications.Notification{Title: "Sync", Content: filesUpdatedNotificationContent}
			logs.InfoLogger.Printf("job completed: %s\n", time.Now().Format(time.RFC3339))
		}
	})
}

func downloadFile(baseDir string, file api.File) error {
	fileDirSlice := append([]string{baseDir}, file.Ancestors...)
	ensureDir(filepath.Join(append(fileDirSlice, file.Name)...))

	downloadReq, dlReqErr := api.BuildDocumentRequest(file, api.DOWNLOAD_FILE)
	if dlReqErr != nil {
		return dlReqErr
	}

	return downloadReq.Download(filepath.Join(fileDirSlice...))
}

func ensureDir(dir string) {
	dirName := filepath.Dir(dir)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			logs.ErrorLogger.Println(merr)
			panic(merr)
		}
	}
}
