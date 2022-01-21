// Package file provides primitives to file indexing for sync operations.
package indexing

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	logs "github.com/beebeeoii/lominus/internal/log"
	"github.com/beebeeoii/lominus/pkg/api"
)

// IndexMap struct contains an array of IndexMapEntry.
// It is used to create a .csv IndexMap file for file comparison during syncs.
type IndexMap struct {
	Entries []IndexMapEntry
}

// IndexMapEntry struct contains the file Id, name and last updated (unix).
// These are the data used for file comparison during syncs.
type IndexMapEntry struct {
	Id          string
	FileName    string
	LastUpdated int64
}

const INDEX_MAP_FILE_NAME = "index_map.csv"

// Build is used to create a map of the current files on the local desktop.
// The built map will be used to compare with the IndexMap to determine whether a file
// needs to be downloaded or updated.
func Build(dir string) (map[string]api.File, error) {
	filesMap := make(map[string]api.File)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == dir {
			return nil
		}

		if !info.IsDir() {
			filesMap[info.Name()] = api.File{
				Name:        info.Name(),
				Ancestors:   strings.Split(path[len(dir)+1:], string(os.PathSeparator)),
				LastUpdated: info.ModTime(),
			}
		}
		return nil
	})

	return filesMap, err
}

// CreateIndexMap writes the IndexMap to a csv file for which can be loaded for the next sync.
func CreateIndexMap(indexMap IndexMap) error {
	logs.Logger.Infof("Creating index map: %s\n", INDEX_MAP_FILE_NAME)

	indexMapPath, getIndexMapPathErr := getIndexMapPath()
	if getIndexMapPathErr != nil {
		return getIndexMapPathErr
	}

	indexMapFile, _ := os.Create(indexMapPath)
	w := csv.NewWriter(indexMapFile)

	for _, entry := range indexMap.Entries {
		err := w.Write([]string{entry.Id, entry.FileName, strconv.FormatInt(entry.LastUpdated, 10)}) //{[id], [fileName], [lastUpdated]}
		if err != nil {
			return err
		}
	}
	w.Flush()

	logs.Logger.Infoln("Index map created successfully.")
	return nil
}

// LoadIndexMap loads the IndexMap csv file back to a map of IndexMapEntry, with the key being the file Id.
func LoadIndexMap(file io.Reader) (map[string]IndexMapEntry, error) {
	logs.Logger.Infof("Loading index map: %s\n", INDEX_MAP_FILE_NAME)
	r := csv.NewReader(file)
	indexMap := map[string]IndexMapEntry{}

	for {
		record, err := r.Read() //record: {[id], [fileName], [lastUpdated]}
		if err == io.EOF {
			break
		}
		if err != nil {
			return indexMap, err
		}

		id := record[0]
		fileName := record[1]
		lastUpdated, _ := strconv.ParseInt(record[2], 10, 64)
		indexMap[id] = IndexMapEntry{Id: id, FileName: fileName, LastUpdated: lastUpdated}
	}
	return indexMap, nil
}

// getIndexMapPath returns the file path to the IndexMap csv file.
func getIndexMapPath() (string, error) {
	var indexMapPath string

	baseDir, retrieveBaseDirErr := appDir.GetBaseDir()
	if retrieveBaseDirErr != nil {
		return indexMapPath, retrieveBaseDirErr
	}

	indexMapPath = filepath.Join(baseDir, INDEX_MAP_FILE_NAME)

	return indexMapPath, nil
}
