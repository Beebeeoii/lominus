// Package api provides functions that link up and communicate with Luminus servers.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
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

// DocumentRequest struct is the datapack for containing details about a specific HTTP request used for documents (Luminus Files).
type DocumentRequest struct {
	File    File
	Folder  Folder
	Module  Module
	Request Request
	Mode    int
}

// GradeRequest struct is the datapack for containing details about a specific HTTP request used for grades (Luminus Gradebook).
type GradeRequest struct {
	Module  Module
	Request Request
}

// ModuleRequest struct is the datapack for containing details about a specific HTTP request used for modules being taken.
type ModuleRequest struct {
	Request Request
}

type ModulesRequest struct {
	Request Request
}

type FoldersRequest struct {
	Request Request
	Builder interface{}
}

type FilesRequest struct {
	Request Request
	Folder  Folder
}

// MultimediaChannelRequest struct is the datapack for containing details about a specific HTTP request used for multimedia channels (Luminus Multimedia).
type MultimediaChannelRequest struct {
	Module  Module
	Request Request
}

// MultimediaVideoRequest struct is the datapack for containing details about a specific HTTP request used for multimedia video (Luminus Multimedia).
type MultimediaVideoRequest struct {
	MultimediaChannel MultimediaChannel
	Request           Request
}

const (
	GET_ALL_FOLDERS = 0
	GET_ALL_FILES   = 1
	DOWNLOAD_FILE   = 2
)

const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0"
const POST = "POST"
const GET_METHOD = "GET"
const CONTENT_TYPE_FORM = "application/x-www-form-urlencoded"
const CONTENT_TYPE_JSON = "application/json; charset=UTF-8"

// BuildModuleRequest builds and returns a ModuleRequest that can be used for Module related operations
// such as retrieving all modules.
func BuildModuleRequest() (ModuleRequest, error) {
	jwtToken, jwtTokenErr := retrieveJwtToken()
	if jwtTokenErr != nil {
		return ModuleRequest{}, jwtTokenErr
	}

	return ModuleRequest{
		Request: Request{
			Url: interfaces.Url{
				Url:      MODULE_URL_ENDPOINT,
				Platform: constants.Luminus,
			},
			Token:     jwtToken,
			UserAgent: USER_AGENT,
		},
	}, nil
}

// TODO Documentations
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

// TODO Documentations
func BuildFoldersRequest(token string, platform constants.Platform, builder interface{}) (FoldersRequest, error) {
	var url string

	switch builder := builder.(type) {
	case Module:
		switch p := platform; p {
		case constants.Canvas:
			url = fmt.Sprintf(constants.CANVAS_MODULE_FOLDERS_ENDPOINT, builder.Id)
		case constants.Luminus:
			url = fmt.Sprintf(FOLDER_URL_ENDPOINT, builder.Id)
		default:
			return FoldersRequest{}, errors.New("invalid platform provided")
		}
	case Folder:
		switch p := platform; p {
		case constants.Canvas:
			url = fmt.Sprintf(constants.CANVAS_FOLDERS_ENDPOINT, builder.Id)
		case constants.Luminus:
			url = fmt.Sprintf(FOLDER_URL_ENDPOINT, builder.Id)
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

// func BuildCanvasDocumentRequest(token string, builder interface{}, mode int) (DocumentRequest, error) {
// 	var urlEndpoint string

// 	switch mode {
// 	case GET_ALL_FOLDERS:
// 		_, isModule := builder.(Module)
// 		_, isFolder := builder.(Folder)
// 		if !isModule && !isFolder {
// 			return DocumentRequest{}, errors.New("invalid mode: DocumentRequest must be built using Module or Folder to have mode=GET_ALL_FOLDERS")
// 		}
// 		urlEndpoint = constants.CANVAS_FOLDERS_ENDPOINT
// 	case GET_ALL_FILES:
// 		_, isModule := builder.(Module)
// 		_, isFolder := builder.(Folder)
// 		if !isModule && !isFolder {
// 			return DocumentRequest{}, errors.New("invalid mode: DocumentRequest must be built using Module or Folder to have mode=GET_ALL_FILES")
// 		}
// 		urlEndpoint = constants.CANVAS_FILES_ENDPOINT
// 	case DOWNLOAD_FILE:
// 		_, isFile := builder.(File)
// 		if !isFile {
// 			return DocumentRequest{}, errors.New("invalid mode: DocumentRequest must be built using File to download")
// 		}
// 		urlEndpoint = constants.CANVAS_FILE_ENDPOINT
// 	default:
// 		return DocumentRequest{}, errors.New("invalid mode: mode provided is invalid. Valid modes are GET_ALL_FOLDERS (0), GET_ALL_FILES (1), DOWNLOAD_FILE (2)")
// 	}

// 	switch builder := builder.(type) {
// 	case Module:
// 		return DocumentRequest{
// 			Folder: Folder{
// 				Id:           builder.Id,
// 				Name:         builder.ModuleCode,
// 				Downloadable: true,
// 				Ancestors:    []string{},
// 				HasSubFolder: true,
// 			},
// 			Request: Request{
// 				Url: interfaces.Url{
// 					Url:      fmt.Sprintf(urlEndpoint, builder.Id),
// 					Platform: constants.Canvas,
// 				},
// 				Token:     token,
// 				UserAgent: USER_AGENT,
// 			},
// 			Mode: mode,
// 		}, nil
// 	case Folder:
// 		return DocumentRequest{
// 			Folder: builder,
// 			Request: Request{
// 				Url: interfaces.Url{
// 					Url:      fmt.Sprintf(urlEndpoint, builder.Id),
// 					Platform: constants.Canvas,
// 				},
// 				Token:     token,
// 				UserAgent: USER_AGENT,
// 			},
// 			Mode: mode,
// 		}, nil
// 	case File:
// 		return DocumentRequest{
// 			File: builder,
// 			Request: Request{
// 				Url: interfaces.Url{
// 					Url:      fmt.Sprintf(urlEndpoint, builder.Id),
// 					Platform: constants.Canvas,
// 				},
// 				Token:     token,
// 				UserAgent: USER_AGENT,
// 			},
// 			Mode: mode,
// 		}, nil
// 	default:
// 		return DocumentRequest{}, errors.New("invalid builder: DocumentRequest must be built using Module, Folder or File only")
// 	}
// }

// BuildGradeRequest builds and returns a GradeRequest that can be used for Grade related operations
// such as retrieving grades of a module.
// A Module is required to build a GradeRequest as it is module specific.
func BuildGradeRequest(module Module) (GradeRequest, error) {
	jwtToken, jwtTokenErr := retrieveJwtToken()
	if jwtTokenErr != nil {
		return GradeRequest{}, jwtTokenErr
	}

	return GradeRequest{
		Module: module,
		Request: Request{
			Url: interfaces.Url{
				Url:      fmt.Sprintf(GRADE_URL_ENDPOINT, module.Id),
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

// BuildDocumentRequest builds and returns a DocumentRequest that can be used for File/Folder related operations
// such as retrieving files/folders of a module.
// DocumentRequests must be built using Module/Folder/File only.
// Building it with Folder enables you to specify the specific folder you are interested in.
// Building it with Module taks you to the root folder of the module files.
// Building it with File enables you to download the file.
//
// Modes available:
// GET_FOLDERS - retrieves folders in a specific folder (root folder if Module was used to build the DocumentRequest). Nested folders are not returned.
// GET_ALL_FILES - retrieves all files in a specific folder (root folder if Module was used to build the DocumentRequest). Nested files are returned.
// DOWNLOAD_FILE - downloads a particular file. DocumentRequest must be built with File.
func BuildDocumentRequest(builder interface{}, mode int) (DocumentRequest, error) {
	jwtToken, jwtTokenErr := retrieveJwtToken()
	if jwtTokenErr != nil {
		return DocumentRequest{}, jwtTokenErr
	}

	var urlEndpoint string

	switch mode {
	case GET_ALL_FOLDERS:
		_, isModule := builder.(Module)
		_, isFolder := builder.(Folder)
		if !isModule && !isFolder {
			return DocumentRequest{}, errors.New("invalid mode: DocumentRequest must be built using Module or Folder to have mode=GET_ALL_FOLDERS")
		}
		urlEndpoint = FOLDER_URL_ENDPOINT
	case GET_ALL_FILES:
		_, isModule := builder.(Module)
		_, isFolder := builder.(Folder)
		if !isModule && !isFolder {
			return DocumentRequest{}, errors.New("invalid mode: DocumentRequest must be built using Module or Folder to have mode=GET_ALL_FILES")
		}
		urlEndpoint = FILE_URL_ENDPOINT
	case DOWNLOAD_FILE:
		_, isFile := builder.(File)
		if !isFile {
			return DocumentRequest{}, errors.New("invalid mode: DocumentRequest must be built using File to download")
		}
		urlEndpoint = DOWNLOAD_URL_ENDPOINT
	default:
		return DocumentRequest{}, errors.New("invalid mode: mode provided is invalid. Valid modes are GET_ALL_FOLDERS (0), GET_ALL_FILES (1), DOWNLOAD_FILE (2)")
	}

	switch builder := builder.(type) {
	case Module:
		return DocumentRequest{
			Folder: Folder{
				Id:           builder.Id,
				Name:         builder.ModuleCode,
				Downloadable: true,
				Ancestors:    []string{},
				HasSubFolder: true,
			},
			Request: Request{
				Url: interfaces.Url{
					Url:      fmt.Sprintf(urlEndpoint, builder.Id),
					Platform: constants.Luminus,
				},
				Token:     jwtToken,
				UserAgent: USER_AGENT,
			},
			Mode: mode,
		}, nil
	case Folder:
		return DocumentRequest{
			Folder: builder,
			Request: Request{
				Url: interfaces.Url{
					Url:      fmt.Sprintf(urlEndpoint, builder.Id),
					Platform: constants.Luminus,
				},
				Token:     jwtToken,
				UserAgent: USER_AGENT,
			},
			Mode: mode,
		}, nil
	case File:
		return DocumentRequest{
			File: builder,
			Request: Request{
				Url: interfaces.Url{
					Url:      fmt.Sprintf(urlEndpoint, builder.Id),
					Platform: constants.Luminus,
				},
				Token:     jwtToken,
				UserAgent: USER_AGENT,
			},
			Mode: mode,
		}, nil
	default:
		return DocumentRequest{}, errors.New("invalid builder: DocumentRequest must be built using Module, Folder or File only")
	}
}

// retrieveJwtToken is a util function that loads user's JWT data to be used to communicate with Luminus servers.
func retrieveJwtToken() (string, error) {
	jwtPath, getJwtPathErr := appAuth.GetJwtPath()
	if getJwtPathErr != nil {
		return jwtPath, getJwtPathErr
	}

	jwtData, jwtErr := auth.LoadJwtData(jwtPath)
	if jwtErr != nil {
		return jwtData.JwtToken, jwtErr
	}

	if !jwtData.IsExpired() {
		return jwtData.JwtToken, nil
	}

	credentialsPath, getCredentialsPathErr := appAuth.GetCredentialsPath()
	if getCredentialsPathErr != nil {
		return jwtData.JwtToken, getCredentialsPathErr
	}

	credentials, credentialsErr := auth.LoadCredentials(credentialsPath)
	if credentialsErr != nil {
		return jwtData.JwtToken, credentialsErr
	}

	return auth.RetrieveJwtToken(credentials, true)
}

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
