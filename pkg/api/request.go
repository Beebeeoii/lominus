// Package api provides functions that link up and communicate with LMS servers,
// such as Canvas and Luminus (probably removed in near future).
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
	appFile "github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/pkg/auth"
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

// GradeRequest struct is the datapack for containing details about a specific
// HTTP request used for grades (Luminus Gradebook only).
type GradeRequest struct {
	Module  Module
	Request Request
}

// ModulesRequest struct is the datapack for containing details about a specific
// HTTP request used for retrieving all the modules taken by the user.
type ModulesRequest struct {
	Request Request
}

// FoldersRequest struct is the datapack for containing details about a specific
// HTTP request used for retrieving folders in a module's uploaded files on Luminus/Canvas.
type FoldersRequest struct {
	Request Request
	Builder interface{}
}

// FilesRequest struct is the datapack for containing details about a specific
// HTTP request used for retrieving files in a module's uploaded files on Luminus/Canvas.
type FilesRequest struct {
	Request Request
	Folder  Folder
}

// MultimediaChannelRequest struct is the datapack for containing details about a
// specific HTTP request used for multimedia channels (Luminus Multimedia).
type MultimediaChannelRequest struct {
	Module  Module
	Request Request
}

// MultimediaVideoRequest struct is the datapack for containing details about a
// specific HTTP request used for multimedia video (Luminus Multimedia).
type MultimediaVideoRequest struct {
	MultimediaChannel MultimediaChannel
	Request           Request
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
	case constants.Luminus:
		url = MODULE_URL_ENDPOINT
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
		case constants.Luminus:
			url = fmt.Sprintf(FOLDER_URL_ENDPOINT, b.Id)
		default:
			return FoldersRequest{}, errors.New("invalid platform provided")
		}
	case Folder:
		switch p := platform; p {
		case constants.Canvas:
			url = fmt.Sprintf(constants.CANVAS_FOLDERS_ENDPOINT, b.Id)
		case constants.Luminus:
			url = fmt.Sprintf(FOLDER_URL_ENDPOINT, b.Id)
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
	case constants.Luminus:
		url = fmt.Sprintf(FILE_URL_ENDPOINT, folder.Id)
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

// BuildMultimediaChannelRequest builds and returns a MultimediaChannelRequuest that can be used for Multimedia
// channel related operations such as retrieving all channels of a module.
// A Module is required to build a BuildMultimediaChannelRequest as it is module specific.
func BuildMultimediaChannelRequest(module Module) (MultimediaChannelRequest, error) {
	jwtToken, jwtTokenErr := retrieveJwtToken()
	if jwtTokenErr != nil {
		return MultimediaChannelRequest{}, jwtTokenErr
	}

	return MultimediaChannelRequest{
		Module: module,
		Request: Request{
			Url: interfaces.Url{
				Url:      fmt.Sprintf(MULTIMEMDIA_CHANNEL_URL_ENDPOINT, module.Id),
				Platform: constants.Luminus,
			},
			Token:     jwtToken,
			UserAgent: USER_AGENT,
		},
	}, nil
}

// BuildMultimediaChannelRequest builds and returns a MultimediaChannelRequuest that can be used for Multimedia
// channel related operations such as retrieving all channels of a module.
// A Module is required to build a BuildMultimediaChannelRequest as it is module specific.
func BuildMultimediaVideoRequest(multimediaChannel MultimediaChannel) (MultimediaVideoRequest, error) {
	jwtToken, jwtTokenErr := retrieveJwtToken()
	if jwtTokenErr != nil {
		return MultimediaVideoRequest{}, jwtTokenErr
	}

	return MultimediaVideoRequest{
		MultimediaChannel: multimediaChannel,
		Request: Request{
			Url: interfaces.Url{
				Url:      fmt.Sprintf(LTI_DATA_URL_ENDPOINT, multimediaChannel.Id),
				Platform: constants.Luminus,
			},
			Token:     jwtToken,
			UserAgent: USER_AGENT,
		},
	}, nil
}

// retrieveJwtToken is a util function that loads user's JWT data to be used to communicate with Luminus servers.
func retrieveJwtToken() (string, error) {
	jwtPath, getJwtPathErr := appAuth.GetTokensPath()
	if getJwtPathErr != nil {
		return jwtPath, getJwtPathErr
	}

	tokensData, tokensErr := auth.LoadTokensData(jwtPath, true)
	if tokensErr != nil {
		return tokensData.LuminusToken.JwtToken, tokensErr
	}

	return tokensData.LuminusToken.JwtToken, tokensErr
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

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, res)
	if err != nil {
		return err
	}

	return err
}
