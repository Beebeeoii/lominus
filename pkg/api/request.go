// Package api provides functions that link up and communicate with Luminus servers.
package api

import (
	"errors"
	"fmt"

	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
	"github.com/beebeeoii/lominus/pkg/auth"
)

// Request struct is the datapack for containing details about a HTTP request.
type Request struct {
	Url       string
	JwtToken  string
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

const (
	GET_ALL_FOLDERS = 0
	GET_ALL_FILES   = 1
	DOWNLOAD_FILE   = 2
)

const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0"

// BuildModuleRequest builds and returns a ModuleRequest that can be used for Module related operations
// such as retrieving all modules.
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
			Url:       fmt.Sprintf(GRADE_URL_ENDPOINT, module.Id),
			JwtToken:  jwtToken,
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
		urlEndpoint = FOLDER_URL_ENDPOINT
	case GET_ALL_FILES:
		urlEndpoint = FILE_URL_ENDPOINT
	case DOWNLOAD_FILE:
		_, isFile := builder.(File)
		if !isFile {
			return DocumentRequest{}, errors.New("invalid arguments: DocumentRequest must be built using File to download")
		}
		urlEndpoint = DOWNLOAD_URL_ENDPOINT
	default:
		return DocumentRequest{}, errors.New("invalid arguments: mode provided is not a valid mode")
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

// retrieveJwtToken is a util function that loads user's JWT data to be used to communicate with Luminus servers.
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
