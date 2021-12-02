package sync

import (
	"os"

	"github.com/beebeeoii/lominus/internal/indexing"
)

func sync() {
	//TODO
}

func CreateIndexMap(indexMapArray []indexing.IndexMap) {
	// indexMap1 := indexing.IndexMap{Id: "1", FileName: "1.txt", LastUpdated: time.Now().Unix()}
	// indexMap2 := indexing.IndexMap{Id: "2", FileName: "2.txt", LastUpdated: time.Now().Unix()}
	// indexMap3 := indexing.IndexMap{Id: "3", FileName: "3.txt", LastUpdated: time.Now().Unix()}
	// indexMap4 := indexing.IndexMap{Id: "4", FileName: "4.txt", LastUpdated: time.Now().Unix()}
	// indexMap5 := indexing.IndexMap{Id: "5", FileName: "5.txt", LastUpdated: time.Now().Unix()}
	// indexMap6 := indexing.IndexMap{Id: "6", FileName: "6.txt", LastUpdated: time.Now().Unix()}
	// indexMap := []indexing.IndexMap{indexMap1, indexMap2, indexMap3, indexMap4, indexMap5, indexMap6}
	indexing.CreateIndexMap(indexing.Map{IndexMap: indexMapArray})
}

func LoadIndexMap() (map[string]indexing.IndexMap, error) {
	var indexMap map[string]indexing.IndexMap
	file, fileErr := os.Open(indexing.INDEX_MAP_FILE_NAME)
	if fileErr != nil {
		return indexMap, fileErr
	}
	return indexing.LoadIndexMap(file)
}
