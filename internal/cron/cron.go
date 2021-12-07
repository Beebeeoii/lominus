package cron

import (
	"os"
	"path/filepath"
	"time"

	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	"github.com/beebeeoii/lominus/internal/indexing"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/pkg/api"
	"github.com/beebeeoii/lominus/pkg/pref"
	"github.com/gen2brain/beeep"

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
		logs.InfoLogger.Printf("job started: %s\n", time.Now().Format(time.RFC3339))
		LastRanChannel <- GetLastRan().Format("2 Jan 15:04:05")

		preferences, prefErr := pref.LoadPreferences(appPref.GetPreferencesPath())
		if prefErr != nil {
			logs.WarningLogger.Println(prefErr)
			return
		}

		if preferences.Directory != "" {
			moduleRequest, modReqErr := api.BuildModuleRequest()
			if modReqErr != nil {
				logs.WarningLogger.Println(modReqErr)
				err := beeep.Notify("Sync", "Authentication failed", "assets/app-icon.png")
				if err != nil {
					logs.ErrorLogger.Println(err)
					panic(err)
				}
				return
			}

			modules, modErr := moduleRequest.GetModules()
			if modErr != nil {
				logs.WarningLogger.Println(modErr)
				err := beeep.Notify("Sync", "Unable to retrieve modules.", "assets/app-icon.png")
				if err != nil {
					logs.ErrorLogger.Println(err)
					panic(err)
				}
				return
			}

			updatedFiles := make([]api.File, 0)
			for _, module := range modules {
				fileRequest, fileReqErr := api.BuildDocumentRequest(module, api.GET_ALL_FILES)
				if fileReqErr != nil {
					logs.WarningLogger.Println(fileReqErr)
					err := beeep.Notify("Sync", "Unable to retrieve files.", "assets/app-icon.png")
					if err != nil {
						logs.ErrorLogger.Println(err)
						panic(err)
					}
					continue
				}

				files, fileErr := fileRequest.GetAllFiles()
				if fileErr != nil {
					logs.WarningLogger.Println(fileErr)
					err := beeep.Notify("Sync", "Unable to retrieve files.", "assets/app-icon.png")
					if err != nil {
						logs.ErrorLogger.Println(err)
						panic(err)
					}
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
				logs.WarningLogger.Println(currentFilesErr)
				err := beeep.Notify("Sync", "Unable to sync files.", "assets/app-icon.png")
				if err != nil {
					logs.ErrorLogger.Println(err)
					panic(err)
				}
				return
			}

			for _, file := range updatedFiles {
				if _, exists := currentFiles[file.Name]; !exists || currentFiles[file.Name].LastUpdated.Before(file.LastUpdated) {
					downloadErr := downloadFile(preferences.Directory, file)
					if downloadErr != nil {
						logs.ErrorLogger.Println(downloadErr)
						err := beeep.Notify("Sync", "Unable to download files.", "assets/app-icon.png")
						if err != nil {
							logs.ErrorLogger.Println(err)
							panic(err)
						}
						continue
					}
				}
			}

			logs.InfoLogger.Printf("job completed: %s\n", time.Now().Format(time.RFC3339))
			err := beeep.Notify("Sync", "Your files are up to date.", "assets/app-icon.png")
			if err != nil {
				logs.ErrorLogger.Println(err)
				panic(err)
			}
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
