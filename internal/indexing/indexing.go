package indexing

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type IndexMap struct {
	Entries []IndexMapEntry
}

type IndexMapEntry struct {
	Id          string
	FileName    string
	LastUpdated int64
}

type Index struct {
	Id          string
	FileName    string
	IsDir       bool
	FileSize    int64
	Directory   string
	LastUpdated int64
}

const INDEX_MAP_FILE_NAME = "index_map.csv"
const INDEX_FILE_NAME = "index.csv"

func Build(dir string) error {
	log.Printf("Creating index file: %s", INDEX_FILE_NAME)
	indexFile, _ := os.Create(INDEX_FILE_NAME)
	w := csv.NewWriter(indexFile)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == dir {
			return nil
		}

		w.Write([]string{info.Name(), path[len(dir)+1:], strconv.FormatBool(info.IsDir()), strconv.FormatInt(info.Size(), 10), strconv.FormatInt(info.ModTime().Unix(), 10)})
		w.Flush()
		return nil
	})
	if err != nil {
		return err
	}
	log.Printf("Index created: %s", INDEX_FILE_NAME)
	return nil
}

func CreateIndexMap(indexMap IndexMap) error {
	log.Printf("Creating index map: %s", INDEX_MAP_FILE_NAME)
	indexMapFile, _ := os.Create(INDEX_MAP_FILE_NAME)
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
