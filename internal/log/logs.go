// Package logs provides primitives to initialise and access the logger.
package logs

import (
	"os"
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	"github.com/beebeeoii/lominus/internal/lominus"
	log "github.com/sirupsen/logrus"
)

var (
	Logger *log.Logger
)

// Init initialises the log file and the different loggers: WarningLogger, InfoLogger and ErrorLogger.
func Init() error {
	file, err := os.OpenFile(getLogPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	var logLevel string
	preferences, loadPrefErr := appPref.LoadPreferences(appPref.GetPreferencesPath())
	if loadPrefErr != nil {
		logLevel = "info"
	} else {
		logLevel = preferences.LogLevel
	}

	Logger = log.New()
	Logger.SetOutput(file)
	Logger.SetReportCaller(true)
	Logger.SetFormatter(&log.JSONFormatter{
		PrettyPrint: true,
	})
	Logger.SetLevel(getLogLevel()(logLevel))

	return nil
}

// getLogPath returns the file path to the log file.
func getLogPath() string {
	return filepath.Join(appDir.GetBaseDir(), lominus.LOG_FILE_NAME)
}

// getLogLevel returns log.Level corresponding to the string representative of the log level
func getLogLevel() func(string) log.Level {
	innerMap := map[string]log.Level{
		"panic": log.PanicLevel,
		"fatal": log.FatalLevel,
		"error": log.ErrorLevel,
		"warn":  log.WarnLevel,
		"info":  log.InfoLevel,
		"debug": log.DebugLevel,
		"trace": log.TraceLevel,
	}

	return func(key string) log.Level {
		return innerMap[key]
	}
}
