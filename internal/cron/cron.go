package cron

import (
	"time"

	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	logs "github.com/beebeeoii/lominus/internal/log"
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

	job, err := createJob(preferences.Frequency)
	if err != nil {
		return err
	}

	mainJob = job
	mainScheduler.StartAsync()

	return nil
}

func UpdateFrequency(frequency int) error {
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
	return mainScheduler.Every(frequency).Seconds().Do(func() {
		logs.InfoLogger.Println(GetLastRan())
		LastRanChannel <- GetLastRan().Format("2 Jan 15:04:05")
	})
}
