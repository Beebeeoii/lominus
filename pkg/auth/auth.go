// Package auth provides functions that link up and communicate with the LMS.
package auth

import (
	file "github.com/beebeeoii/lominus/internal/file"
)

// CredentialsData struct is the datapack that contains all the credentials for
// available LMS (Canvas etc.).
//
// Note: Credentials refer to things like username/password or access tokens.
type CredentialsData struct {
	CanvasCredentials CanvasCredentials
}

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
