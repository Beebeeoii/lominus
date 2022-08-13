// Package auth provides functions that link up and communicate with LMS (Luminus/Canvas)
// authentication server.
package auth

import (
	"errors"
	"net/http"

	"github.com/beebeeoii/lominus/pkg/constants"
)

// CanvasTokenData is a struct that encapsulates the token required for authentication.
// In this case, it is the CanvasApiToken which is a string.
type CanvasTokenData struct {
	CanvasApiToken string
}

// CanvasCredentials is a struct that encapsulates the credentials required for authentication.
// In this case, it is the CanvasApiToken which is a string.
type CanvasCredentials struct {
	CanvasApiToken string
}

// Save takes in the CanvasTokenData and saves it locally with the path provided as arguments.
func (canvasTokenData CanvasTokenData) Save(tokensPath string) error {
	return saveTokenData(tokensPath, TokensData{
		CanvasToken: canvasTokenData,
	})
}

// Save takes in the CanvasCredentials and saves it locally with the path provided as arguments.
func (credentials CanvasCredentials) Save(credentialsPath string) error {
	return saveCredentialsData(credentialsPath, CredentialsData{
		CanvasCredentials: credentials,
	})
}

// Authenticate checks whether the CanvasCredentials provided is valid.
// This is done by sending a HTTP request to the Canvas server to retrieve information
// on the account using the CanvasCredentials.
// If the credentials is valid, the response status code is expected to be 200.
func (credentials CanvasCredentials) Authenticate() error {
	request, err := http.NewRequest("GET", constants.CANVAS_USER_SELF_ENDPOINT, nil)
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", "Bearer "+credentials.CanvasApiToken)

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return errors.New("invalid Canvas credentials")
	}

	return nil
}
