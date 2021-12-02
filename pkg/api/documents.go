package api

type folder struct {
	Id string
	Name string
	Subfolder bool
	downloadable bool
	TotalFileCount int
}

const FOLDER_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/files/?populate=totalFileCount%2CsubFolderCount%2CTotalSize&ParentID="

func getAllDownloadableFolderNames() []folder {
	
}

func GetAllFileNames() {
	for item in folders:
		while item.subfolder:
			// call getalldownloadernames --> a series of ids

}

