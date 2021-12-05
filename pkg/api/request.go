package api

import (
	"errors"
	"fmt"

	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
	"github.com/beebeeoii/lominus/pkg/auth"
)

type Request struct {
	Url       string
	JwtToken  string
	UserAgent string
}

type DocumentRequest struct {
	File    File
	Folder  Folder
	Module  Module
	Request Request
	Mode    int
}

type GradeRequest struct {
	Module  Module
	Request Request
}

type ModuleRequest struct {
	Request Request
}

const (
	GET_FOLDERS   = 0
	GET_ALL_FILES = 1
	DOWNLOAD_FILE = 2
	get_files     = 3
)

const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0"

func BuildModuleRequest() (ModuleRequest, error) {
	jwtToken, jwtTokenErr := retrieveJwtToken()
	if jwtTokenErr != nil {
		return ModuleRequest{}, jwtTokenErr
	}

	return ModuleRequest{
		Request: Request{
			Url:       MODULE_URL_ENDPOINT,
			JwtToken:  jwtToken,
			UserAgent: USER_AGENT,
		},
	}, nil
}

func BuildGradeRequest(module Module) (GradeRequest, error) {
	jwtToken, jwtTokenErr := retrieveJwtToken()
	if jwtTokenErr != nil {
		return GradeRequest{}, jwtTokenErr
	}

	return GradeRequest{
		Module: module,
		Request: Request{
			Url:       fmt.Sprintf(GRADE_URL_ENDPOINT, module.Id),
			JwtToken:  jwtToken,
			UserAgent: USER_AGENT,
		},
	}, nil
}

// DocumentRequests must be built using Module or Folder only.
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
	case GET_FOLDERS:
		urlEndpoint = FOLDER_URL_ENDPOINT
	case GET_ALL_FILES:
		urlEndpoint = FOLDER_URL_ENDPOINT
	case DOWNLOAD_FILE:
		_, isFile := builder.(File)
		if !isFile {
			return DocumentRequest{}, errors.New("invalid arguments: DocumentRequest must be built using File to download")
		}
		urlEndpoint = DOWNLOAD_URL_ENDPOINT
	case get_files:
		urlEndpoint = FILE_URL_ENDPOINT
	default:
		return DocumentRequest{}, errors.New("invalid arguments: mode provided is not a valid mode")
	}

	switch builder := builder.(type) {
	case Module:
		return DocumentRequest{
			Module: builder,
			Request: Request{
				Url:       fmt.Sprintf(urlEndpoint, builder.Id),
				JwtToken:  jwtToken,
				UserAgent: USER_AGENT,
			},
			Mode: mode,
		}, nil
	case Folder:
		return DocumentRequest{
			Folder: builder,
			Request: Request{
				Url:       fmt.Sprintf(urlEndpoint, builder.Id),
				JwtToken:  jwtToken,
				UserAgent: USER_AGENT,
			},
			Mode: mode,
		}, nil
	case File:
		return DocumentRequest{
			File: builder,
			Request: Request{
				Url:       fmt.Sprintf(urlEndpoint, builder.Id),
				JwtToken:  jwtToken,
				UserAgent: USER_AGENT,
			},
			Mode: mode,
		}, nil
	default:
		return DocumentRequest{}, errors.New("invalid arguments: DocumentRequest must be built using Module or Folder only")
	}
}

func retrieveJwtToken() (string, error) {
	jwtData, jwtErr := auth.LoadJwtData(appAuth.GetJwtPath())
	if jwtErr != nil {
		return jwtData.JwtToken, jwtErr
	}

	if !jwtData.IsExpired() {
		return jwtData.JwtToken, nil
	}

	credentials, credentialsErr := auth.LoadCredentials(appAuth.GetCredentialsPath())
	if credentialsErr != nil {
		return jwtData.JwtToken, credentialsErr
	}

	return auth.RetrieveJwtToken(credentials, true)
}
