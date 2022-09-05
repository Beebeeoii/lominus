// Package api provides functions that link up and communicate with LMS servers,
// such as Canvas and Luminus (probably removed in near future).
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

	appFile "github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/pkg/constants"
	"github.com/beebeeoii/lominus/pkg/interfaces"
	"github.com/mitchellh/mapstructure"
)

// Folder struct is the datapack for containing details about a Folder.
// Ancestors describe the relative folders that precedes the current folder, exclusing itself.
// Eg. Ancestors for a folder with the path: /MA2001/Lectures/Week1 is ['MA2001', 'Lectures'].
// IsRootFolder = true if the folder is the root for a module. Root folders are expected to
// be named the module code.
type Folder struct {
	Id           string
	Name         string
	Downloadable bool
	HasSubFolder bool
	Ancestors    []string
	IsRootFolder bool
}

// File struct is the datapack for containing details about a File.
// Ancestors describe the relative folders that precedes the current file, excluding itself.
// Eg. Ancestors for a file with the path: /MA2001/Lectures/Lecture1.pdf is ['MA2001', 'Lectures'].
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

// GetFolders returns a slice of Folder objects from a given FoldersRequest.
// Only the folders in the current Folder/Module (via the builder) provided
// in the FoldersRequest will be returned. In other words, nested folders will not be included.
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
			// We do not want to download that folder as we just want the folders and files in that
			// folder.
			//
			// The "course files" folder resembles the 'home directory' of a module.
			downloadable := !folderObject.HiddenForUser

			if folderObject.FullName == "course files" {
				downloadable = false
			}

			folders = append(folders, Folder{
				Id:           strconv.Itoa(folderObject.Id),
				Name:         appFile.CleanseFolderFileName(folderObject.Name),
				Downloadable: downloadable,
				HasSubFolder: folderObject.FoldersCount > 0,
				Ancestors:    ancestors,
				IsRootFolder: folderObject.ParentFolderId == 0 &&
					folderObject.FullName == "course files",
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
				IsRootFolder: false,
			})
		}
	default:
		return folders, errors.New("foldersRequest.Request.Url.Platform is not available")
	}

	return folders, nil
}

// GetRootFiles is a recursive function that returns a slice of File objects and nested
// File objects that are in a Folder.
// Note that it will traverse all nested folders and return all nested files.
func (foldersRequest FoldersRequest) GetRootFiles() ([]File, error) {
	files := []File{}

	if foldersRequest.Request.Token == "" {
		return files, nil
	}

	switch builder := foldersRequest.Builder.(type) {
	case Module:
		// Module exists but its contents are restricted to be downloaded.
		if !builder.IsAccessible {
			return files, nil
		}

		// Retrieving of folders in main folder is only required for Luminus
		// as Canvas already returns its
		if foldersRequest.Request.Url.Platform != constants.Luminus {
			break
		}

		moduleMainFolder := Folder{
			Id:           builder.Id,
			Name:         builder.Name,
			Downloadable: true,
			HasSubFolder: true,       // doesn't matter
			Ancestors:    []string{}, // main folder does not have any ancestors
		}
		subFilesReq, subFilesReqErr := BuildFilesRequest(
			foldersRequest.Request.Token,
			foldersRequest.Request.Url.Platform,
			moduleMainFolder,
		)
		if subFilesReqErr != nil {
			return files, subFilesReqErr
		}

		subFiles, subFilesErr := subFilesReq.GetFiles()
		if subFilesErr != nil {
			return files, subFilesErr
		}

		files = append(files, subFiles...)
	case Folder:
		// Folder exists but its contents are restricted to be downloaded.
		if !builder.Downloadable {
			return files, nil
		}

		subFilesReq, subFilesReqErr := BuildFilesRequest(
			foldersRequest.Request.Token,
			foldersRequest.Request.Url.Platform,
			builder,
		)
		if subFilesReqErr != nil {
			return files, subFilesReqErr
		}

		subFiles, subFilesErr := subFilesReq.GetFiles()
		if subFilesErr != nil {
			return files, subFilesErr
		}

		files = append(files, subFiles...)

		if !builder.HasSubFolder {
			break
		}
	}

	subFoldersReq, subFoldersReqErr := BuildFoldersRequest(
		foldersRequest.Request.Token,
		foldersRequest.Request.Url.Platform,
		foldersRequest.Builder,
	)
	if subFoldersReqErr != nil {
		return files, subFoldersReqErr
	}

	subFolders, subFoldersErr := subFoldersReq.GetFolders()
	if subFoldersErr != nil {
		return files, subFoldersErr
	}

	for _, subFolder := range subFolders {
		nestedFoldersReq, nestedFoldersReqErr := BuildFoldersRequest(
			foldersRequest.Request.Token,
			foldersRequest.Request.Url.Platform,
			subFolder,
		)
		if nestedFoldersReqErr != nil {
			return files, nestedFoldersReqErr
		}

		nestedFiles, nestedFilesErr := nestedFoldersReq.GetRootFiles()
		if nestedFilesErr != nil {
			return files, nestedFilesErr
		}

		files = append(files, nestedFiles...)
	}

	return files, nil
}

// GetFiles returns a slice of File objects from a given FilesRequest.
// Only the files in the current Folder provided in the FilesRequest will be returned.
// In other words, nested files will not be included.
func (filesRequest FilesRequest) GetFiles() ([]File, error) {
	files := []File{}

	if !filesRequest.Folder.Downloadable || filesRequest.Request.Token == "" {
		return files, nil
	}

	ancestors := append(filesRequest.Folder.Ancestors, filesRequest.Folder.Name)

	switch folderDataType := filesRequest.Request.Url.Platform; folderDataType {
	case constants.Canvas:
		// All the folders and files of a module are stored under the "course files" folder.
		// We do not want to get that folder as we just want the folders and files in that
		// folder.
		//
		// The "course files" folder resembles the 'home directory' of a module.
		if filesRequest.Folder.IsRootFolder {
			ancestors = filesRequest.Folder.Ancestors
		}

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
				Name:        appFile.CleanseFolderFileName(fileObject.DisplayName),
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
				Name:        appFile.CleanseFolderFileName(fileObject.Name),
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

// Download downloads the given file via the DownloadUrl of the File object.
// The downloaded file will be placed in the folderPath specified in the parameter.
func (file File) Download(folderPath string) error {
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

	filePath := filepath.Join(folderPath, file.Name)

	// This checks if there already exists the specified file
	// to prevent overwritting of files.
	// If file exists, new file will have a new name appended with [vX],
	// where X is an integer.
	if appFile.Exists(filePath) {
		filePath = filepath.Join(folderPath, appFile.AutoRename(filePath))
	}

	f, err := os.Create(filePath)
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
