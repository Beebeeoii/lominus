// Package api provides functions that link up and communicate with Luminus servers.
package api

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/beebeeoii/lominus/internal/file"
)

// Folder struct is the datapack for containing details about a Folder
// Ancestors the relative folders that precedes the current folder, including itself.
// Eg. Ancestors for a folder with the path: /folder1/folder2/folder3 is ['folder1', 'folder2', 'folder3']
type Folder struct {
	Id           string
	Name         string
	Downloadable bool
	HasSubFolder bool
	Ancestors    []string
}

// File struct is the datapack for containing details about a File
// Ancestors the relative folders that precedes the current folder, including itself.
// Eg. Ancestors for a file with the path: /folder1/folder2/file.pdf is ['folder1', 'folder2', 'file.pdf']
type File struct {
	Id          string
	Name        string
	Ancestors   []string
	LastUpdated time.Time
}

const FOLDER_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/files/?populate=totalFileCount,subFolderCount,TotalSize&ParentID=%s"
const FILE_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/files/%s/file?populate=Creator,lastUpdatedUser,comment"
const DOWNLOAD_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/files/file/%s/downloadurl"

// getFolderFieldsRequired is a helper function that returns a constant array with fields that a Folder response
// returned by Luminus needs.
func getFolderFieldsRequired() []string {
	return []string{"access", "id", "name", "isActive", "allowUpload", "subFolderCount"}
}

// getFileFieldsRequired is a helper function that returns a constant array with fields that a File response
// returned by Luminus needs.
func getFileFieldsRequired() []string {
	return []string{"id", "name", "lastUpdatedDate"}
}

// GetAllFolders returns a slice of Folder objects from a DocumentRequest.
// It will only return folders in the current folder.
// Nested folders will not be returned.
// Ensure that DocumentRequest mode is GET_ALL_FOLDERS (0).
// Find out more about DocumentRequests under request.go.
func (req DocumentRequest) GetAllFolders() ([]Folder, error) {
	folders := []Folder{}
	if req.Mode != GET_ALL_FOLDERS {
		return folders, errors.New("mode mismatched: ensure DocumentRequest mode is GET_ALL_FOLDERS (0)")
	}

	rawResponse := RawResponse{}
	err := req.Request.GetRawResponse(&rawResponse)
	if err != nil {
		return folders, err
	}

	for _, content := range rawResponse.Data {
		if !IsResponseValid(getFolderFieldsRequired(), content) {
			continue
		}

		if _, exists := content["access"]; exists { // only folder that can be accessed will be placed in folders slice
			newFolder := Folder{
				Id:           content["id"].(string),
				Name:         file.CleanseFolderFileName(content["name"].(string)),
				Downloadable: content["isActive"].(bool) && !content["allowUpload"].(bool), // downloadable = active folder + does not allow uploads
				HasSubFolder: int(content["subFolderCount"].(float64)) > 0,
				Ancestors:    append(req.Folder.Ancestors, req.Folder.Name),
			}
			folders = append(folders, newFolder)
		}
	}
	return folders, nil
}

// GetAllFiles returns a slice of File objects that are in a Folder from a DocumentRequest.
// It will only return files in the current folder.
// To return nested files, use GetRootFiles() instead.
// Ensure that DocumentRequest mode is GET_ALL_FILES (1).
// Find out more about DocumentRequests under request.go.
func (req DocumentRequest) GetAllFiles() ([]File, error) {
	files := []File{}
	if req.Mode != GET_ALL_FILES {
		return files, errors.New("mode mismatched: ensure DocumentRequest mode is GET_ALL_FILES (1)")
	}

	rawResponse := RawResponse{}
	err := req.Request.GetRawResponse(&rawResponse)
	if err != nil {
		return files, err
	}

	for _, content := range rawResponse.Data {
		if !IsResponseValid(getFileFieldsRequired(), content) {
			continue
		}

		lastUpdated, timeParseErr := time.Parse(time.RFC3339, content["lastUpdatedDate"].(string))

		if timeParseErr != nil {
			return files, timeParseErr
		}

		file := File{
			Id:          content["id"].(string),
			Name:        file.CleanseFolderFileName(content["name"].(string)),
			LastUpdated: lastUpdated,
			Ancestors:   append(req.Folder.Ancestors, req.Folder.Name),
		}
		files = append(files, file)
	}

	return files, nil
}

// GetRootFiles returns a slice of File objects and nested File objects that are in a Folder from a DocumentRequest.
// It will traverse all nested folders and return all nested files.
// Ensure that DocumentRequest mode is GET_ALL_FILES (1).
// Find out more about DocumentRequests under request.go.
func (req DocumentRequest) GetRootFiles() ([]File, error) {
	files := []File{}
	if req.Mode != GET_ALL_FILES {
		return files, errors.New("mode mismatched: ensure DocumentRequest mode is GET_ALL_FILES (1)")
	}

	if !req.Folder.Downloadable {
		return files, nil
	}

	if req.Folder.HasSubFolder {
		subFolderReq, subFolderReqBuildErr := BuildDocumentRequest(req.Folder, GET_ALL_FOLDERS)
		if subFolderReqBuildErr != nil {
			return files, subFolderReqBuildErr
		}

		subFolders, err := subFolderReq.GetAllFolders()
		if err != nil {
			return files, err
		}

		for _, subFolder := range subFolders {
			rootFilesReq, rootFilesBuildErr := BuildDocumentRequest(subFolder, GET_ALL_FILES)
			if rootFilesBuildErr != nil {
				return files, rootFilesBuildErr
			}

			subFiles, err := rootFilesReq.GetRootFiles()
			if err != nil {
				return files, err
			}

			files = append(files, subFiles...)
		}
	}

	subFiles, err := req.GetAllFiles()
	if err != nil {
		return files, err
	}

	files = append(files, subFiles...)

	return files, nil
}

// Download downloads the specified File in a DocumentRequest into local storage.
// Ensure that DocumentRequest mode is DOWNLOAD_FILE (2).
// Find out more about DocumentRequests under request.go.
func (req DocumentRequest) Download(filePath string) error {
	if req.Mode != DOWNLOAD_FILE {
		return errors.New("mode mismatched: ensure DocumentRequest mode is DOWNLOAD_FILE (2)")
	}

	downloadResponse := DownloadResponse{}
	err := req.Request.GetRawResponse(&downloadResponse)
	if err != nil {
		return err
	}

	response, err := http.Get(downloadResponse.DownloadUrl)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("received non 200 response code")
	}

	file, err := os.Create(filepath.Join(filePath, req.File.Name))
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
