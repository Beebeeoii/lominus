// Package api provides functions that link up and communicate with Luminus servers.
package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/pkg/constants"
	"github.com/beebeeoii/lominus/pkg/interfaces"
	"github.com/mitchellh/mapstructure"
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
	DownloadUrl string
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

// type FolderObject interface {
// 	interfaces.CanvasFolderObject | interfaces.LuminusFolderObject
// }

// type retrieveFolderObjects[T interfaces.CanvasFolderObject | interfaces.LuminusFolderObject] func(T) []T

// type FolderInterface interface {
// 	GetId() int
// 	Name() string
// 	HiddenForUser() bool
// 	FoldersCount() int
// 	Ancesters() []string
// 	interfaces.CanvasFolderObject | interfaces.LuminusFolderObject
// }

// func (folder Folder) GetAllFolders(token string, data interface{}) error {
// 	folders := []Folder{}

// 	var folderObjects interface{}

// 	switch folderObject := data.(type) {
// 	case []interfaces.CanvasFolderObject:
// 		foldersReq, _ := BuildCanvasDocumentRequest(token, folder, GET_ALL_FOLDERS)
// 		foldersReq.Request.Send(&folderObject)
// 	case []interfaces.LuminusFolderObject:
// 		folderObjects = []interfaces.LuminusFolderObject(folderObjects.([]interfaces.LuminusFolderObject))
// 	default:
// 		return errors.New(
// 			"invalid builder: DocumentRequest must be built using CanvasFolderObject or LuminusFolderObject only",
// 		)
// 	}

// 	for _, folderObject := range folderObjects.([]interfaces.CanvasFolderObject) {
// 		folders = append(folders, Folder{
// 			Id:           strconv.Itoa(folderObject.GetId()),
// 			Name:         folderObject.Name(),
// 			Downloadable: !folderObject.HiddenForUser(),
// 			HasSubFolder: folderObject.FoldersCount() > 0,
// 			Ancestors:    append(folder.Ancesters(), folderObject.Name()),
// 		})
// 	}

// 	return folders, nil
// }

func (foldersRequest FoldersRequest) GetFolders() ([]Folder, error) {
	folders := []Folder{}
	ancestors := []string{}

	switch builder := foldersRequest.Builder.(type) {
	case Module:
		ancestors = append(ancestors, builder.ModuleCode)
	case Folder:
		ancestors = append(builder.Ancestors, builder.Name)
	}

	switch folderDataType := foldersRequest.Request.Url.Platform; folderDataType {
	case constants.Canvas:
		response := []interfaces.CanvasFolderObject{}
		foldersRequest.Request.Send(&response)

		for _, folderObject := range response {
			// All the folders and files of a module are stored under the "course files" folder.
			// We do not want to get that folder as we just want the folders and files in that
			// folder.
			//
			// The "course files" folder resembles the 'home directory' of a module.
			if folderObject.FullName == "course files" {
				continue
			}

			folders = append(folders, Folder{
				Id:           strconv.Itoa(folderObject.Id),
				Name:         file.CleanseFolderFileName(folderObject.Name),
				Downloadable: !folderObject.HiddenForUser,
				HasSubFolder: folderObject.FoldersCount > 0,
				Ancestors:    ancestors,
			})
		}
	case constants.Luminus:
		foldersData := []interfaces.LuminusFolderObject{}

		response := interfaces.LuminusRawResponse{}
		foldersRequest.Request.Send(&response)

		data := reflect.ValueOf(response.Data)
		if data.Kind() == reflect.Slice {
			for i := 0; i < data.Len(); i++ {
				folderData := interfaces.LuminusFolderObject{}
				mapstructure.Decode(data.Index(i).Interface(), &folderData)
				foldersData = append(foldersData, folderData)
			}
		}

		for _, folderObject := range foldersData {
			// Folder is not available.
			if reflect.ValueOf(folderObject.AccessObject).IsNil() {
				continue
			}

			folders = append(folders, Folder{
				Id:           folderObject.Id,
				Name:         folderObject.Name,
				Downloadable: folderObject.IsActive && !folderObject.AllowUpload,
				HasSubFolder: folderObject.FoldersCount > 0,
				Ancestors:    ancestors,
			})
		}
	default:
		return folders, errors.New("foldersRequest.Request.Url.Platform is not available")
	}

	return folders, nil
}

func (filesRequest FilesRequest) GetFiles() ([]File, error) {
	files := []File{}

	if !filesRequest.Folder.Downloadable {
		return files, nil
	}

	ancestors := append(filesRequest.Folder.Ancestors, filesRequest.Folder.Name)

	switch folderDataType := filesRequest.Request.Url.Platform; folderDataType {
	case constants.Canvas:
		response := []interfaces.CanvasFileObject{}
		filesRequest.Request.Send(&response)

		for _, fileObject := range response {
			lastUpdated, err := time.Parse(time.RFC3339, fileObject.LastUpdated)
			if err != nil {
				return files, err
			}

			files = append(files, File{
				Id:          strconv.Itoa(fileObject.Id),
				Name:        file.CleanseFolderFileName(fileObject.Name),
				LastUpdated: lastUpdated,
				Ancestors:   ancestors,
				DownloadUrl: fileObject.Url,
			})
		}
	case constants.Luminus:
		filesData := []interfaces.LuminusFileObject{}

		response := interfaces.LuminusRawResponse{}
		filesRequest.Request.Send(&response)

		data := reflect.ValueOf(response.Data)
		if data.Kind() == reflect.Slice {
			for i := 0; i < data.Len(); i++ {
				fileData := interfaces.LuminusFileObject{}
				mapstructure.Decode(data.Index(i).Interface(), &fileData)
				filesData = append(filesData, fileData)
			}
		}

		for _, fileObject := range filesData {
			lastUpdated, err := time.Parse(time.RFC3339, fileObject.LastUpdated)
			if err != nil {
				return files, err
			}

			downloadUrlResponse := LuminusDownloadResponse{}
			downloadRequest := Request{
				Method:    GET_METHOD,
				Token:     filesRequest.Request.Token,
				UserAgent: filesRequest.Request.UserAgent,
				Url: interfaces.Url{
					Url:      fmt.Sprintf(DOWNLOAD_URL_ENDPOINT, fileObject.Id),
					Platform: filesRequest.Request.Url.Platform,
				},
			}
			downloadUrlResponseErr := downloadRequest.GetRawResponse(&downloadUrlResponse)
			if downloadUrlResponseErr != nil {
				return files, downloadUrlResponseErr
			}

			files = append(files, File{
				Id:          fileObject.Id,
				Name:        file.CleanseFolderFileName(fileObject.Name),
				LastUpdated: lastUpdated,
				Ancestors:   ancestors,
				DownloadUrl: downloadUrlResponse.DownloadUrl,
			})
		}
	default:
		return files, errors.New("filesRequest.Request.Url.Platform is not available")
	}

	return files, nil
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

	downloadResponse := LuminusDownloadResponse{}
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

func (file File) Download(filePath string) error {
	if file.DownloadUrl == "" {
		return errors.New("file.DownloadUrl is empty")
	}

	response, err := http.Get(file.DownloadUrl)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("received non 200 response code")
	}

	f, err := os.Create(filepath.Join(filePath, file.Name))
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.Copy(f, response.Body)
	if err != nil {
		return err
	}

	return nil
}
