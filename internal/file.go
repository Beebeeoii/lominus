package file

import (
	"encoding/gob"
	"fmt"
	"os"
)

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

func Exists(name string) bool {
	_, err := os.Stat(name)

	return err == nil
}

type FileNotFoundError struct {
	FileName string
}

func (e *FileNotFoundError) Error() string {
	return fmt.Sprintf("FileNotFoundError: %s cannot be found.", e.FileName)
}
