// Package logs provides primitives to initialise and access the logger.
package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appPref "github.com/beebeeoii/lominus/internal/app/pref"
	"github.com/beebeeoii/lominus/internal/lominus"
	log "github.com/sirupsen/logrus"
)

// Logger instance to be used for logging purposes
var (
	Logger *log.Logger
)

// LominusFormatter provides the format for log outputs
type LominusFormatter struct {
	log.TextFormatter
}

// Format returns the formatted log string
func (f *LominusFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("[%s] %s - %s (%s:%s:%d)\n", entry.Time.Format(f.TimestampFormat), strings.ToUpper(entry.Level.String()), entry.Message, entry.Caller.File, entry.Caller.Func.Name(), entry.Caller.Line)), nil
}

// Init initialises the log file and the different loggers: WarningLogger, InfoLogger and ErrorLogger.
func Init() error {
	logPath, getLogPathErr := getLogPath()
	if getLogPathErr != nil {
		return getLogPathErr
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	var logLevel string
	preferencesPath, getPreferencesPathErr := appPref.GetPreferencesPath()
	if getPreferencesPathErr != nil {
		return getPreferencesPathErr
	}

	preferences, loadPrefErr := appPref.LoadPreferences(preferencesPath)
	if loadPrefErr != nil {
		logLevel = "info"
	} else {
		logLevel = preferences.LogLevel
	}

	Logger = log.New()
	Logger.SetOutput(file)
	Logger.SetReportCaller(true)
	Logger.SetFormatter(&LominusFormatter{log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}})
	Logger.SetLevel(getLogLevel()(logLevel))

	return nil
}

// SetLogLevel sets the log level for global Logger by taking in the string representation of a log.Level
func SetLogLevel(logLevel string) error {
	switch logLevel {
	case "panic":
	case "fatal":
	case "error":
	case "warn":
	case "info":
	case "debug":
	case "trace":
		Logger.SetLevel(getLogLevel()(logLevel))
		return nil
	}

	return fmt.Errorf("invalid logLevel provided - must be panic, fatal, error, warn, info, debug, trace")
}

// getLogPath returns the file path to the log file.
func getLogPath() (string, error) {
	var logPath string

	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		return logPath, retrieveBaseDirErr
	}

	logPath = filepath.Join(baseDir, lominus.LOG_FILE_NAME)

	return logPath, nil
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
