package cron

import (
	"log"
	"os"
	"path/filepath"
	"time"

	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	"github.com/beebeeoii/lominus/internal/indexing"
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
		LastRanChannel <- GetLastRan().Format("2 Jan 15:04:05")

		preferences, prefErr := pref.LoadPreferences(appPref.GetPreferencesPath())
		if prefErr != nil {
			return
		}

		if preferences.Directory != "" {
			moduleRequest, modReqErr := api.BuildModuleRequest()
			if modReqErr != nil {
				log.Println(modReqErr)
				return
			}

			modules, modErr := moduleRequest.GetModules()
			if modErr != nil {
				log.Println(modErr)
				return
			}

			updatedFiles := make([]api.File, 0)
			for _, module := range modules {
				fileRequest, fileReqErr := api.BuildDocumentRequest(module, api.GET_ALL_FILES)
				if fileReqErr != nil {
					log.Println(fileReqErr)
					continue
				}

				files, fileErr := fileRequest.GetAllFiles()
				if fileErr != nil {
					log.Println(fileErr)
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
				log.Println(currentFilesErr)
				return
			}

			for _, file := range updatedFiles {
				if _, exists := currentFiles[file.Name]; !exists || currentFiles[file.Name].LastUpdated.Before(file.LastUpdated) {
					downloadErr := downloadFile(preferences.Directory, file)
					if downloadErr != nil {
						log.Println(downloadErr)
						continue
					}
				}
			}
		}
	})
}

func downloadFile(baseDir string, file api.File) error {
	fileDirSlice := append([]string{baseDir}, file.Ancestors...)
	ensureDir(filepath.Join(append(fileDirSlice, file.Name)...))

	downloadReq, dlReqErr := api.BuildDocumentRequest(file, api.DOWNLOAD_FILE)
	if dlReqErr != nil {
		log.Println(dlReqErr)
		return dlReqErr
	}

	return downloadReq.Download(filepath.Join(fileDirSlice...))
}

func ensureDir(dir string) {
	log.Println(dir)
	dirName := filepath.Dir(dir)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			panic(merr)
		}
	}
}
