// Package logs provides primitives to initialise and access the logger.
package logs

import (
	"log"
	"os"
	"path/filepath"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	"github.com/beebeeoii/lominus/internal/lominus"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

// Init initialises the log file and the different loggers: WarningLogger, InfoLogger and ErrorLogger.
func Init() error {
	file, err := os.OpenFile(getLogPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	log.SetOutput(file)

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}

// getLogPath returns the file path to the log file.
func getLogPath() string {
	return filepath.Join(appDir.GetBaseDir(), lominus.LOG_FILE_NAME)
}
