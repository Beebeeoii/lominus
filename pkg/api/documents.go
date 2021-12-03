package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Folder struct {
	Id           string
	Name         string
	Subfolder    bool
	downloadable bool
}

type Document struct {
	Id   string
	Name string
}

const FOLDER_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/files/?populate=totalFileCount%2CsubFolderCount%2CTotalSize&ParentID="

func (req Request) GetAllDownloadableFolderNames(modules []Module) (map[string][]Folder, error) { //consider making this method private
	folders := make(map[string][]Folder)

	for _, module := range modules {

		var folder []Folder
		request, err := http.NewRequest("GET", req.Url+module.Id, nil)

		if err != nil {
			return folders, err
		}

		request.Header.Add("Authorization", "Bearer "+req.JwtToken)

		cl := &http.Client{}
		response, err := cl.Do(request)

		if err != nil {
			return folders, err
		}

		body, err := ioutil.ReadAll(response.Body)

		if err != nil {
			return folders, err
		}

		var obj RawResponse                                //variable which holds the raw data
		json.Unmarshal([]byte(string([]byte(body))), &obj) //Converting from byte to struct

		for _, content := range obj.Data {

			if _, ok := content["access"]; ok { // only folder that can be accessed will be placed in folders slice

				newStruct := Folder{
					content["id"].(string),
					content["name"].(string),
					content["subFolderCount"].(float64) > 0,
					content["isActive"].(bool) && !content["allowUpload"].(bool), // downloadable = active folder + does not allow uploads
				}
				folder = append(folder, newStruct)
			}
		}
		folders[module.ModuleCode] = folder
	}
	return folders, nil
}

func (req Request) GetAllFileNames(modules []Module) {
	//fols, err := req.GetAllDownloadableFolderNames(modules)
	//return (map[string][]Document, error)
	/*
		result = map of Document slices
		for key,value in map:
			temp = slice of documents
			for folder in folders:
				if !folder.subFolder:
					add files into temp
				else:
					// get to the bottom folder
					// add files to temp
			add temp to result
	*/
}
