// Package file provides util primitives to file operations.
package file

import (
	"encoding/gob"
	"fmt"
	"os"
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

// FileNotFoundError struct is an error struct that contains the custom error that will be thrown when file is not found.
type FileNotFoundError struct {
	FileName string
}

// FileNotFoundError is an error that will be thrown when file is not found.
func (e *FileNotFoundError) Error() string {
	return fmt.Sprintf("FileNotFoundError: %s cannot be found.", e.FileName)
}
