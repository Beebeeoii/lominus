// Package api provides functions that link up and communicate with LMS servers,
// such as Canvas.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	appFile "github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/pkg/constants"
	"github.com/beebeeoii/lominus/pkg/interfaces"
)

// Request struct is the datapack for containing details about a HTTP request.
type Request struct {
	Method    string
	Token     string
	Url       interfaces.Url
	UserAgent string
}

// ModulesRequest struct is the datapack for containing details about a specific
// HTTP request used for retrieving all the modules taken by the user.
type ModulesRequest struct {
	Request Request
}

// FoldersRequest struct is the datapack for containing details about a specific
// HTTP request used for retrieving folders in a module's uploaded files.
type FoldersRequest struct {
	Request Request
	Builder interface{}
}

// FilesRequest struct is the datapack for containing details about a specific
// HTTP request used for retrieving files in a module's uploaded files.
type FilesRequest struct {
	Request Request
	Folder  Folder
}

type ModuleFolderRequest struct {
	Request Request
	Module  Module
}

const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0"
const POST = "POST"
const GET_METHOD = "GET"
const CONTENT_TYPE_FORM = "application/x-www-form-urlencoded"
const CONTENT_TYPE_JSON = "application/json; charset=UTF-8"

// BuildModulesRequest builds and returns a ModulesRequest that can be used to retrieve
// all modules taken by a user.
func BuildModulesRequest(token string, platform constants.Platform) (ModulesRequest, error) {
	var url string

	switch p := platform; p {
	case constants.Canvas:
		url = constants.CANVAS_MODULES_ENDPOINT
	default:
		return ModulesRequest{}, errors.New("invalid platform provided")
	}

	return ModulesRequest{
		Request: Request{
			Method: GET_METHOD,
			Token:  token,
			Url: interfaces.Url{
				Url:      url,
				Platform: platform,
			},
			UserAgent: USER_AGENT,
		},
	}, nil
}

// BuildFoldersRequest builds and returns a FoldersRequest that can be used for Folder related
// operations such as retrieving folders of a module.
func BuildFoldersRequest(token string, platform constants.Platform, builder interface{}) (FoldersRequest, error) {
	var url string

	switch b := builder.(type) {
	case Module:
		switch p := platform; p {
		case constants.Canvas:
			url = fmt.Sprintf(constants.CANVAS_MODULE_FOLDERS_ENDPOINT, b.Id)
			folderRequest := FoldersRequest{
				Request: Request{
					Method: GET_METHOD,
					Token:  token,
					Url: interfaces.Url{
						Url:      url,
						Platform: platform,
					},
					UserAgent: USER_AGENT,
				},
				Builder: b,
			}

			folders, foldersErr := folderRequest.GetFolders()
			if foldersErr != nil {
				return folderRequest, foldersErr
			}

			var rootFolderId string
			for _, folder := range folders {
				if folder.Name == "course files" {
					rootFolderId = folder.Id
					break
				}
			}

			if rootFolderId == "" {
				return folderRequest, foldersErr
			}

			url = fmt.Sprintf(constants.CANVAS_FOLDERS_ENDPOINT, b.Id)

			builder = Folder{
				Id:           rootFolderId,
				Name:         appFile.CleanseFolderFileName(b.ModuleCode),
				Downloadable: b.IsAccessible,
				HasSubFolder: true,
				Ancestors:    []string{},
			}
		default:
			return FoldersRequest{}, errors.New("invalid platform provided")
		}
	case Folder:
		switch p := platform; p {
		case constants.Canvas:
			url = fmt.Sprintf(constants.CANVAS_FOLDERS_ENDPOINT, b.Id)
		default:
			return FoldersRequest{}, errors.New("invalid platform provided")
		}
	default:
		return FoldersRequest{}, errors.New(
			"invalid mode: FoldersRequest must be built using Module or Folder",
		)
	}

	return FoldersRequest{
		Request: Request{
			Method: GET_METHOD,
			Token:  token,
			Url: interfaces.Url{
				Url:      url,
				Platform: platform,
			},
			UserAgent: USER_AGENT,
		},
		Builder: builder,
	}, nil
}

// BuildFilesRequest builds and returns a FilesRequest that can be used for File related operations
// such as retrieving files of a module.
func BuildFilesRequest(token string, platform constants.Platform, folder Folder) (FilesRequest, error) {
	var url string

	switch p := platform; p {
	case constants.Canvas:
		url = fmt.Sprintf(constants.CANVAS_FILES_ENDPOINT, folder.Id)
	default:
		return FilesRequest{}, errors.New("invalid platform provided")
	}

	return FilesRequest{
		Request: Request{
			Method: GET_METHOD,
			Token:  token,
			Url: interfaces.Url{
				Url:      url,
				Platform: platform,
			},
			UserAgent: USER_AGENT,
		},
		Folder: folder,
	}, nil
}

func BuildModuleFolderRequest(token string, module Module) (ModuleFolderRequest, error) {
	url := fmt.Sprintf(constants.CANVAS_MODULE_FOLDER_ENDPOINT, module.Id)

	return ModuleFolderRequest{
		Request: Request{
			Method: GET_METHOD,
			Token:  token,
			Url: interfaces.Url{
				Url:      url,
				Platform: constants.Canvas,
			},
			UserAgent: USER_AGENT,
		},
		Module: module,
	}, nil
}

// Send takes a Request that encapsulates a HTTP request and sends it. The response body is then
// unmarshalled into the interface{} argument provided.
// Note that the argument parsed must be a pointer.
func (req Request) Send(res interface{}) error {
	request, err := http.NewRequest(req.Method, req.Url.Url, nil)
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", "Bearer "+req.Token)

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, res)
	if err != nil {
		return err
	}

	return err
}
