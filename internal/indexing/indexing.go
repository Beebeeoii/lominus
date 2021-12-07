package indexing

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	appDir "github.com/beebeeoii/lominus/internal/app/dir"
	"github.com/beebeeoii/lominus/pkg/api"
)

type IndexMap struct {
	Entries []IndexMapEntry
}

type IndexMapEntry struct {
	Id          string
	FileName    string
	LastUpdated int64
}

const INDEX_MAP_FILE_NAME = "index_map.csv"

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

func CreateIndexMap(indexMap IndexMap) error {
	log.Printf("Creating index map: %s", INDEX_MAP_FILE_NAME)
	indexMapFile, _ := os.Create(getIndexMapPath())
	w := csv.NewWriter(indexMapFile)

	for _, entry := range indexMap.Entries {
		err := w.Write([]string{entry.Id, entry.FileName, strconv.FormatInt(entry.LastUpdated, 10)}) //{[id], [fileName], [lastUpdated]}
		if err != nil {
			return err
		}
	}
	w.Flush()

	log.Println("Index map created successfully.")
	return nil
}

func LoadIndexMap(file io.Reader) (map[string]IndexMapEntry, error) {
	log.Printf("Loading index map: %s", INDEX_MAP_FILE_NAME)
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

func getIndexMapPath() string {
	return filepath.Join(appDir.GetBaseDir(), INDEX_MAP_FILE_NAME)
}
