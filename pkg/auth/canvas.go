package auth

import (
	"errors"
	"net/http"

	"github.com/beebeeoii/lominus/pkg/constants"
)

// TODO Documentation
type CanvasTokenData struct {
	CanvasApiToken string
}

type CanvasCredentials struct {
	CanvasApiToken string
}

// TODO Documentation
func (canvasTokenData CanvasTokenData) Save(tokensPath string) error {
	return saveTokenData(tokensPath, TokensData{
		CanvasToken: canvasTokenData,
	})
}

// TODO Documentation
func (credentials CanvasCredentials) Save(credentialsPath string) error {
	return saveCredentialsData(credentialsPath, CredentialsData{
		CanvasCredentials: credentials,
	})
}

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
