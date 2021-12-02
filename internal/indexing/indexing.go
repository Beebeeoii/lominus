package indexing

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type Map struct {
	IndexMap []IndexMap
}

type IndexMap struct {
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
		w.Write([]string{info.Name(), path, strconv.FormatBool(info.IsDir()), strconv.FormatInt(info.Size(), 10), strconv.FormatInt(info.ModTime().Unix(), 10)})
		w.Flush()
		return nil
	})
	if err != nil {
		return err
	}
	log.Printf("Index created: %s", INDEX_FILE_NAME)
	return nil
}

func CreateIndexMap(files Map) error {
	log.Printf("Creating index map: %s", INDEX_MAP_FILE_NAME)
	indexMapFile, _ := os.Create(INDEX_MAP_FILE_NAME)
	w := csv.NewWriter(indexMapFile)

	for _, indexMap := range files.IndexMap {
		err := w.Write([]string{indexMap.Id, indexMap.FileName, strconv.FormatInt(indexMap.LastUpdated, 10)})
		if err != nil {
			return err
		}
	}
	w.Flush()

	log.Println("Index map created successfully.")
	return nil
}

func LoadIndexMap(reader io.Reader) (map[string]IndexMap, error) {
	log.Printf("Loading index map: %s", INDEX_MAP_FILE_NAME)
	r := csv.NewReader(reader)
	rows := map[string]IndexMap{}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return rows, err
		}

		id := record[0]
		fileName := record[1]
		lastUpdated, _ := strconv.ParseInt(record[2], 10, 64)
		rows[id] = IndexMap{Id: id, FileName: fileName, LastUpdated: lastUpdated}
	}
	return rows, nil
}

func (files Map) AppendToIndexMap(dir string) error {
	indexMapFile, _ := os.Open(INDEX_MAP_FILE_NAME)
	w := csv.NewWriter(indexMapFile)

	for _, indexMap := range files.IndexMap {
		err := w.Write([]string{indexMap.Id, indexMap.FileName, strconv.FormatInt(indexMap.LastUpdated, 10)})
		if err != nil {
			return err
		}
	}
	w.Flush()

	log.Println("Index map appended successfully.")
	return nil
}
