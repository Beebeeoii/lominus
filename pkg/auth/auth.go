// Package auth provides functions that link up and communicate with the LMS.
package auth

import (
	appConstants "github.com/beebeeoii/lominus/internal/constants"
	file "github.com/beebeeoii/lominus/internal/file"
)

// JsonResponse struct is the datapack for containing API authentication response raw data.
type JsonResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// CredentialsData struct is the datapack that contains all the credentials for
// available LMS (Canvas etc.).
//
// Note: Credentials refer to things like username/password or access tokens.
type CredentialsData struct {
	CanvasCredentials CanvasCredentials
}

const CREDENTIALS_FILE_NAME = appConstants.CREDENTIALS_FILE_NAME

const CONTENT_TYPE = "application/x-www-form-urlencoded"
const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0"
const POST = "POST"
const AUTH_METHOD = "FormsAuthentication"

// saveCredentialsData saves the user's credentials data to local storage for future use.
func saveCredentialsData(credentialsPath string, credentailsData CredentialsData) error {
	if file.Exists(credentialsPath) {
		localCredentialsData, err := LoadCredentialsData(credentialsPath)
		if err != nil {
			return err
		}

		credentailsData.Merge(localCredentialsData)
	}

	return file.EncodeStructToFile(credentialsPath, credentailsData)
}

// LoadCredentialsData loads the user's Credentials data from local storage.
func LoadCredentialsData(credentialsPath string) (CredentialsData, error) {
	credentialsData := CredentialsData{}
	if !file.Exists(credentialsPath) {
		return credentialsData, &file.FileNotFoundError{FileName: credentialsPath}
	}
	err := file.DecodeStructFromFile(credentialsPath, &credentialsData)

	return credentialsData, err
}

// Merge takes n individual Credentials data encapsulated in CredentialsData and merge/combine them
// into a CredentialsData that contains the individual Credentials data.
func (t *CredentialsData) Merge(t2 CredentialsData) {
	if t.CanvasCredentials == (CanvasCredentials{}) {
		t.CanvasCredentials = t2.CanvasCredentials
	}
}

// JwtExpiredError struct contains the JwtExpiredError which will be thrown when the JWT data has expired.
type JwtExpiredError struct{}

// JwtExpiredError error to be thrown when the JWT data has expired.
func (e *JwtExpiredError) Error() string {
	return "JwtExpiredError: JWT token has expired."
}
