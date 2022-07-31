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

// TODO Documentation
func (foldersRequest FoldersRequest) GetFolders() ([]Folder, error) {
	folders := []Folder{}
	ancestors := []string{}

	if foldersRequest.Request.Token == "" {
		return folders, nil
	}

	switch builder := foldersRequest.Builder.(type) {
	case Module:
		// Module exists but its contents are restricted to be downloaded.
		if !builder.IsAccessible {
			return folders, nil
		}
		ancestors = append(ancestors, builder.ModuleCode)
	case Folder:
		// Folder exists but its contents are restricted to be downloaded.
		if !builder.Downloadable {
			return folders, nil
		}
		ancestors = append(builder.Ancestors, builder.Name)
	}

	switch folderDataType := foldersRequest.Request.Url.Platform; folderDataType {
	case constants.Canvas:
		response := []interfaces.CanvasFolderObject{}
		reqErr := foldersRequest.Request.Send(&response)
		if reqErr != nil {
			return folders, reqErr
		}

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
		reqErr := foldersRequest.Request.Send(&response)
		if reqErr != nil {
			return folders, reqErr
		}

		data := reflect.ValueOf(response.Data)
		if data.Kind() == reflect.Slice {
			for i := 0; i < data.Len(); i++ {
				folderData := interfaces.LuminusFolderObject{}
				decodeErr := mapstructure.Decode(data.Index(i).Interface(), &folderData)
				if decodeErr != nil {
					return folders, decodeErr
				}
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

	if !filesRequest.Folder.Downloadable || filesRequest.Request.Token == "" {
		return files, nil
	}

	ancestors := append(filesRequest.Folder.Ancestors, filesRequest.Folder.Name)

	switch folderDataType := filesRequest.Request.Url.Platform; folderDataType {
	case constants.Canvas:
		response := []interfaces.CanvasFileObject{}
		reqErr := filesRequest.Request.Send(&response)
		if reqErr != nil {
			return files, reqErr
		}

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
		reqErr := filesRequest.Request.Send(&response)
		if reqErr != nil {
			return files, reqErr
		}

		data := reflect.ValueOf(response.Data)
		if data.Kind() == reflect.Slice {
			for i := 0; i < data.Len(); i++ {
				fileData := interfaces.LuminusFileObject{}
				decodeErr := mapstructure.Decode(data.Index(i).Interface(), &fileData)
				if decodeErr != nil {
					return files, decodeErr
				}
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
