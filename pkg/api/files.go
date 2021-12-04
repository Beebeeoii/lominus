package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Folder struct {
	Id           string
	Name         string
	Downloadable bool
	HasSubFolder bool
}

type File struct {
	Id   string
	Name string
}

const FOLDER_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/files/?populate=totalFileCount,subFolderCount,TotalSize&ParentID=%s"
const FILE_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/files/%s/file?populate=Creator,lastUpdatedUser,comment"
const DOWNLOAD_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/files/file/%s/downloadurl"

func (req Request) GetAllFolders() ([]Folder, error) {
	folder := []Folder{}

	rawResponse := RawResponse{}
	err := req.GetRawResponse(&rawResponse)
	if err != nil {
		return folder, err
	}

	for _, content := range rawResponse.Data {
		if _, exists := content["access"]; exists { // only folder that can be accessed will be placed in folders slice
			newFolder := Folder{
				Id:           content["id"].(string),
				Name:         content["name"].(string),
				Downloadable: content["isActive"].(bool) && !content["allowUpload"].(bool), // downloadable = active folder + does not allow uploads
				HasSubFolder: int(content["subFolderCount"].(float64)) > 0,
			}
			folder = append(folder, newFolder)
		}
	}
	return folder, nil
}

func (req Request) GetAllFiles() ([]File, error) {
	files := []File{}

	folders, err := req.GetAllFolders()
	if err != nil {
		return files, err
	}

	for _, folder := range folders {
		newReq := Request{
			Url:       fmt.Sprintf(FOLDER_URL_ENDPOINT, folder.Id),
			JwtToken:  req.JwtToken,
			UserAgent: USER_AGENT,
		}

		subFiles, err := newReq.getRootFiles(folder)
		if err != nil {
			return files, err
		}

		files = append(files, subFiles...)
	}
	return files, nil
}

func (req Request) getRootFiles(folder Folder) ([]File, error) {
	files := []File{}

	if !folder.Downloadable {
		return files, nil
	}

	if folder.HasSubFolder {
		newReq := Request{
			Url:       fmt.Sprintf(FOLDER_URL_ENDPOINT, folder.Id),
			JwtToken:  req.JwtToken,
			UserAgent: USER_AGENT,
		}

		subFolders, err := newReq.GetAllFolders()
		if err != nil {
			return files, err
		}

		for _, subFolder := range subFolders {
			newReq := Request{
				Url:       fmt.Sprintf(FOLDER_URL_ENDPOINT, subFolder.Id),
				JwtToken:  req.JwtToken,
				UserAgent: USER_AGENT,
			}

			subFiles, err := newReq.getRootFiles(subFolder)
			if err != nil {
				return files, err
			}

			files = append(files, subFiles...)
		}
	}

	newReq := Request{
		Url:       fmt.Sprintf(FILE_URL_ENDPOINT, folder.Id),
		JwtToken:  req.JwtToken,
		UserAgent: USER_AGENT,
	}

	rawResponse := RawResponse{}
	err := newReq.GetRawResponse(&rawResponse)
	if err != nil {
		return files, err
	}

	for _, content := range rawResponse.Data {
		newFile := File{
			Id:   content["id"].(string),
			Name: content["name"].(string),
		}
		files = append(files, newFile)
	}

	return files, nil
}

func (req Request) Download(file File) error {
	downloadResponse := DownloadResponse{}
	err := req.GetRawResponse(&downloadResponse)
	if err != nil {
		return err
	}
	fmt.Println(downloadResponse)

	response, err := http.Get(downloadResponse.Data)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	doc, err := os.Create("C:\\Users\\jiaju\\Desktop\\" + file.Name)
	if err != nil {
		return err
	}
	defer doc.Close()
	_, err = io.Copy(doc, response.Body)
	if err != nil {
		return err
	}
	return nil
}
