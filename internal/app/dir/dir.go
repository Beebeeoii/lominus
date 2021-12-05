package appDir

import (
	"log"
	"os"
	"path/filepath"

	"github.com/beebeeoii/lominus/internal/lominus"
)

func GetBaseDir() string {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	return filepath.Join(userConfigDir, lominus.APP_NAME)
}
