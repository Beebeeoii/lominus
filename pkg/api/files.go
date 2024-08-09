// Package api provides functions that link up and communicate with LMS servers,
// such as Canvas.
package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	appFile "github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/pkg/constants"
	"github.com/beebeeoii/lominus/pkg/interfaces"
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

func (moduleFolderRequest ModuleFolderRequest) GetModuleFolder() (Folder, error) {
	folder := Folder{}

	if moduleFolderRequest.Request.Token == "" {
		return folder, nil
	}

	response := []interfaces.CanvasFolderObject{}
	reqErr := moduleFolderRequest.Request.Send(&response)
	if reqErr != nil {
		return folder, reqErr
	}

	folder.Id = fmt.Sprint(response[0].Id)
	folder.Name = moduleFolderRequest.Module.ModuleCode
	folder.Downloadable = !response[0].HiddenForUser
	folder.IsRootFolder = true
	folder.HasSubFolder = response[0].FoldersCount > 0

	return folder, nil
}

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

		if len(response) == 10 {
			url, _ := url.Parse(foldersRequest.Request.Url.Url)

			currPage, _ := strconv.Atoi(url.Query().Get("page"))
			if currPage == 0 {
				currPage = 2
			} else {
				currPage += 1
			}

			q := url.Query()
			if q.Has("page") {
				q.Set("page", strconv.Itoa(currPage))
			} else {
				q.Add("page", strconv.Itoa(currPage))
			}
			url.RawQuery = q.Encode()
			foldersRequest.Request.Url.Url = url.String()
			nextFolders, nextFoldersErr := foldersRequest.GetFolders()

			if nextFoldersErr != nil {
				return folders, nextFoldersErr
			}

			folders = append(folders, nextFolders...)
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
			ancestors = []string{filesRequest.Folder.Name}
		}

		response := []interfaces.CanvasFileObject{}
		reqErr := filesRequest.Request.Send(&response)
		if reqErr != nil {
			return files, reqErr
		}

		for _, fileObject := range response {
			lastUpdated, err := time.Parse(time.RFC3339, fileObject.LastUpdated)
			tz, _ := time.LoadLocation("Asia/Singapore")
			lastUpdated = lastUpdated.In(tz)

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

		if len(response) == 10 {
			url, _ := url.Parse(filesRequest.Request.Url.Url)

			currPage, _ := strconv.Atoi(url.Query().Get("page"))
			if currPage == 0 {
				currPage = 2
			} else {
				currPage += 1
			}

			q := url.Query()
			if q.Has("page") {
				q.Set("page", strconv.Itoa(currPage))
			} else {
				q.Add("page", strconv.Itoa(currPage))
			}
			url.RawQuery = q.Encode()
			filesRequest.Request.Url.Url = url.String()
			nextFiles, nextFilesErr := filesRequest.GetFiles()

			if nextFilesErr != nil {
				return files, nextFilesErr
			}

			files = append(files, nextFiles...)
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
	if appFile.Exists(filePath) {
		renameErr := appFile.AutoRename(filePath)

		if renameErr != nil {
			return renameErr
		}
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
