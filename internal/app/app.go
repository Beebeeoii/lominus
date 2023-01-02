// Package app provides primitives to initialise crucial files for Lominus.
package app

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	"github.com/beebeeoii/lominus/internal/file"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/boltdb/bolt"
)

var dbInstance *bolt.DB

// Init initialises and ensures log and preference files that Lominus requires are available.
// Directory in Preferences defaults to empty string ("").
// Frequency in Preferences defaults to -1.
func Init() (*bolt.DB, error) {
	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		return nil, retrieveBaseDirErr
	}

	if !file.Exists(baseDir) {
		os.Mkdir(baseDir, os.ModePerm)
	}

	dbFName := filepath.Join(baseDir, appConstants.DATABASE_FILE_NAME)
	db, dbErr := bolt.Open(dbFName, 0600, &bolt.Options{Timeout: 3 * time.Second})

	if dbErr != nil {
		return nil, dbErr
	}

	err := db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("Auth"))
		tx.CreateBucketIfNotExists([]byte("Integrations"))
		prefBucket, prefBucketErr := tx.CreateBucketIfNotExists([]byte("Preferences"))
		if prefBucketErr != nil {
			return prefBucketErr
		}

		if prefBucket.Get([]byte("frequency")) == nil {
			prefBucket.Put([]byte("frequency"), []byte("-1"))
		}

		logLevel := prefBucket.Get([]byte("logLevel"))
		if logLevel == nil {
			logLevel = []byte("info")
			prefBucket.Put([]byte("logLevel"), []byte(logLevel))

		}

		logInitErr := logs.Init(string(logLevel))
		if logInitErr != nil {
			return logInitErr
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	dbInstance = db

	return db, nil
}

// GetOs returns user's running program's operating system target:
// one of darwin, freebsd, linux, and so on.
// To view possible combinations of GOOS and GOARCH, run "go tool dist list".
func GetOs() string {
	return runtime.GOOS
}

func GetDBInstance() *bolt.DB {
	return dbInstance
}
