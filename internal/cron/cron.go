package cron

import (
	"log"
	"time"

	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	"github.com/beebeeoii/lominus/pkg/pref"

	"github.com/go-co-op/gocron"
)

func Init() error {
	preferences, prefLoadErr := pref.LoadPreferences(appPref.GetPreferencesPath())
	if prefLoadErr != nil {
		return prefLoadErr
	}

	if preferences.Frequency == -1 {
		return nil
	}

	scheduler := gocron.NewScheduler(time.Local)
	_, cronErr := scheduler.Every(1).Seconds().Do(func() {
		log.Println(time.Local)
	})

	if cronErr != nil {
		return cronErr
	}

	scheduler.StartAsync()

	return nil
}
