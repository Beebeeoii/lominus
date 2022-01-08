// Package api provides functions that link up and communicate with Luminus servers.
package api

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// GetAllFolders returns a slice of Folder objects from a DocumentRequest.
// Ensure DocumentRequest mode is GET_FOLDERS (0).
// Find out more about DocumentRequests under request.go.
func (req DocumentRequest) GetAllFolders() ([]Folder, error) {
	folder := []Folder{}
	if req.Mode != GET_FOLDERS {
		return folder, errors.New("mode mismatched: ensure DocumentRequest mode is GET_FOLDERS (0)")
	}

	rawResponse := RawResponse{}
	err := req.Request.GetRawResponse(&rawResponse)
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
				Ancestors:    []string{strings.TrimSpace(req.Module.ModuleCode)},
			}
			folder = append(folder, newFolder)
		}
	}
	return folder, nil
}

// Deprecated - build DocumentRequest with a Folder instead of a module instead, and call getRootFiles() directly.
// GetAllFiles returns a slice of File objects that are in a Folder using a DocumentRequest.
// Ensure DocumentRequest mode is GET_ALL_FILES (1).
// Find out more about DocumentRequests under request.go.
func (req DocumentRequest) GetAllFiles() ([]File, error) {
	files := []File{}
	if req.Mode != GET_ALL_FILES {
		return files, errors.New("mode mismatched: ensure DocumentRequest mode is GET_ALL_FILES (1)")
	}

	rootFilesReq, rootFilesBuildErr := BuildDocumentRequest(Folder{
		Id:           req.Module.Id,
		Name:         req.Module.ModuleCode,
		Downloadable: true,
		Ancestors:    []string{strings.TrimSpace(req.Module.ModuleCode)},
		HasSubFolder: true,
	}, GET_FILES)
	if rootFilesBuildErr != nil {
		return files, rootFilesBuildErr
	}

	baseFiles, err := rootFilesReq.getRootFiles()
	if err != nil {
		return files, err
	}
	files = append(files, baseFiles...)

	return files, nil
}

// getRootFiles returns a slice of File objects and nested File objects that are in a Folder or nested Folder from a DocumentRequest.
// Ensure DocumentRequest mode is GET_FILES (3).
// Find out more about DocumentRequests under request.go.
func (req DocumentRequest) getRootFiles() ([]File, error) {
	files := []File{}
	if req.Mode != GET_FILES {
		return files, errors.New("mode mismatched: ensure DocumentRequest mode is GET_FILES (3)")
	}

	if !req.Folder.Downloadable {
		return files, nil
	}

	if req.Folder.HasSubFolder {
		subFolderReq, subFolderReqBuildErr := BuildDocumentRequest(req.Folder, GET_FOLDERS)
		if subFolderReqBuildErr != nil {
			return files, subFolderReqBuildErr
		}

		subFolders, err := subFolderReq.GetAllFolders()
		if err != nil {
			return files, err
		}

		for _, subFolder := range subFolders {
			subFolder.Ancestors = append(subFolder.Ancestors, req.Folder.Ancestors...)
			subFolder.Ancestors = append(subFolder.Ancestors, strings.TrimSpace(subFolder.Name))
			rootFilesReq, rootFilesBuildErr := BuildDocumentRequest(subFolder, GET_FILES)
			if rootFilesBuildErr != nil {
				return files, rootFilesBuildErr
			}

			subFiles, err := rootFilesReq.getRootFiles()
			if err != nil {
				return files, err
			}

			files = append(files, subFiles...)
		}
	}

	rawResponse := RawResponse{}
	err := req.Request.GetRawResponse(&rawResponse)
	if err != nil {
		return files, err
	}

	for _, content := range rawResponse.Data {
		lastUpdated, timeParseErr := time.Parse(time.RFC3339, content["lastUpdatedDate"].(string))

		if timeParseErr != nil {
			return files, timeParseErr
		}
		newFile := File{
			Id:          content["id"].(string),
			Name:        content["name"].(string),
			LastUpdated: lastUpdated,
			Ancestors:   req.Folder.Ancestors,
		}
		files = append(files, newFile)
	}

	return files, nil
}

// Download downloads the specified file in a DocumentRequest into local storage.
// Ensure DocumentRequest mode is DOWNLOAD_FILE (2).
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
