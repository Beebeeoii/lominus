// Package file provides util primitives to file operations.
package file

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	logs "github.com/beebeeoii/lominus/internal/log"
)

// EncodeStructToFile takes in any struct and encodes it into a file specified by fileName.
// If the file already exists, it is truncated.
// If the file does not exist, it is created with mode 0666 (before umask).
// Provide absolute path else file will be written to the current working directory.
func EncodeStructToFile(fileName string, data interface{}) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	encoder := gob.NewEncoder(file)
	encoder.Encode(data)

	return nil
}

// DecodeStructFromFile takes a file that has been encoded by a struct and decodes it back to the struct.
// Provide absolute path else file may not be found.
func DecodeStructFromFile(fileName string, data interface{}) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(data)

	return err
}

// Exists checks if the given file exists.
func Exists(name string) bool {
	_, err := os.Stat(name)

	return err == nil
}

// EnsureDir is a helper function that ensures that the directory exists by creating them
// if they do not already exist.
func EnsureDir(dir string) {
	dirName := filepath.Dir(dir)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			logs.ErrorLogger.Println(merr)
			panic(merr)
		}
	}
}

// CleanseFolderFileName is a helper function that ensures folders' and files' names are valid,
// that they do not contain prohibited characters. However, some are still not caught for
// unlikeliness and simplicity reasons.
// The following are reserved file names for Windows that are uncaught:
// CON, PRN, AUX, NUL, COM1, COM2, COM3, COM4, COM5, COM6, COM7, COM8, COM9, LPT1, LPT2, LPT3, LPT4, LPT5, LPT6, LPT7, LPT8, LPT9.
// The following are non-printable characters that are uncaught:
// ASCII 0-31.
func CleanseFolderFileName(name string) string {
	name = strings.Replace(name, "/", " ", -1)
	name = strings.Replace(name, "\\", " ", -1)
	name = strings.Replace(name, "<", " ", -1)
	name = strings.Replace(name, ">", " ", -1)
	name = strings.Replace(name, ":", " ", -1)
	name = strings.Replace(name, "\"", " ", -1)
	name = strings.Replace(name, "|", " ", -1)
	name = strings.Replace(name, "?", " ", -1)
	name = strings.Replace(name, "*", " ", -1)
	name = strings.TrimSpace(name)

	return name
}

// FileNotFoundError struct is an error struct that contains the custom error that will be thrown when file is not found.
type FileNotFoundError struct {
	FileName string
}

// FileNotFoundError is an error that will be thrown when file is not found.
func (e *FileNotFoundError) Error() string {
	return fmt.Sprintf("FileNotFoundError: %s cannot be found.", e.FileName)
}
