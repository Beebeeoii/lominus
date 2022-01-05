// Package sync is deprecated (please do not use)
package sync

import (
	"os"

	"github.com/beebeeoii/lominus/internal/indexing"
)

func Sync() {
	//TODO 	compare new and current IndexMap for differences in ids
	// 		if id in new IndexMap but not in current IndexMap:
	//				download directory/file into the correct location
	//		if id in current IndexMap but not in new IndexMap:
	//				do nothing
	//		if id in current and new IndexMap:
	//				check that the file/dir actually exists on local storage. if non-existent, download.
	//		overwrite current index_map.csv with new IndexMap
}

func RetrieveNewIndexMap() {
	// TODO get all directories and files, and their corresponding ids and lastUpdated
	//		return IndexMap, error
	//
	// indexMap1 := indexing.IndexMap{Id: "1", FileName: "1.txt", LastUpdated: time.Now().Unix()}
	// indexMap2 := indexing.IndexMap{Id: "2", FileName: "2.txt", LastUpdated: time.Now().Unix()}
	// indexMap3 := indexing.IndexMap{Id: "3", FileName: "3.txt", LastUpdated: time.Now().Unix()}
	// indexMap4 := indexing.IndexMap{Id: "4", FileName: "4.txt", LastUpdated: time.Now().Unix()}
	// indexMap5 := indexing.IndexMap{Id: "5", FileName: "5.txt", LastUpdated: time.Now().Unix()}
	// indexMap6 := indexing.IndexMap{Id: "6", FileName: "6.txt", LastUpdated: time.Now().Unix()}
	// indexMap := []indexing.IndexMap{indexMap1, indexMap2, indexMap3, indexMap4, indexMap5, indexMap6}
	// indexing.CreateIndexMap(indexing.IndexMap{IndexMapEntries: indexMapArray})
}

func LoadCurrentIndexMap() (map[string]indexing.IndexMapEntry, error) {
	file, fileErr := os.Open(indexing.INDEX_MAP_FILE_NAME)
	if fileErr != nil {
		return map[string]indexing.IndexMapEntry{}, fileErr
	}
	return indexing.LoadIndexMap(file)
}
