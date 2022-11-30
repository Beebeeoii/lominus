// Package auth provides functions that link up and communicate with LMS (Luminus/Canvas)
// authentication server.
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

// CredentialsData struct is the datapack that contains all the different Credentials for
// available LMS (Luminus, Canvas etc.).
type CredentialsData struct {
	CanvasCredentials  CanvasCredentials
	LuminusCredentials LuminusCredentials
}

// TokensData struct is the datapack that contains all the different Tokens for
// available LMS (Luminus, Canvas etc.).
type TokensData struct {
	CanvasToken  CanvasTokenData
	LuminusToken LuminusTokenData
}

const CREDENTIALS_FILE_NAME = appConstants.CREDENTIALS_FILE_NAME

const CONTENT_TYPE = "application/x-www-form-urlencoded"
const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0"
const POST = "POST"
const AUTH_METHOD = "FormsAuthentication"

// saveTokenData saves the user's Tokens data to local storage for future use.
func saveTokenData(tokensPath string, tokensData TokensData) error {
	if file.Exists(tokensPath) {
		localTokensData, err := LoadTokensData(tokensPath, false)
		if err != nil {
			return err
		}

		tokensData.Merge(localTokensData)
	}

	return file.EncodeStructToFile(tokensPath, tokensData)
}

// LoadTokensData loads the user's Tokens data from local storage.
func LoadTokensData(tokensPath string, autoRenew bool) (TokensData, error) {
	tokensData := TokensData{}
	var err error
	if !file.Exists(tokensPath) {
		return tokensData, &file.FileNotFoundError{FileName: tokensPath}
	}

	err = file.DecodeStructFromFile(tokensPath, &tokensData)

	if !autoRenew || !tokensData.LuminusToken.IsExpired() {
		return tokensData, err
	}

	credentialsPath := "appAuth.GetCredentialsPath()"
	// if getCredentialsPathErr != nil {
	// 	return tokensData, getCredentialsPathErr
	// }

	credentials, credentialsErr := LoadCredentialsData(credentialsPath)
	if credentialsErr != nil {
		return tokensData, credentialsErr
	}

	_, retrieveErr := RetrieveJwtToken(credentials.LuminusCredentials, true)
	if retrieveErr != nil {
		return tokensData, retrieveErr
	}

	err = file.DecodeStructFromFile(tokensPath, &tokensData)

	return tokensData, err
}

// saveCredentialsData saves the user's Credentials data to local storage for future use.
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

// Merge takes n individual Token data encapsulated in TokensData and merge/combine them
// into a TokensData that contains the individual Token data.
// Eg. a := TokensData{CanvasToken}
//
//	a.merge(TokensData{LuminusToken}) // a == TokensData{CanvasToken, LuminusToken}
func (t *TokensData) Merge(t2 TokensData) {
	if t.CanvasToken == (CanvasTokenData{}) {
		t.CanvasToken = t2.CanvasToken
	}

	if t.LuminusToken == (LuminusTokenData{}) {
		t.LuminusToken = t2.LuminusToken
	}
}

// Merge takes n individual Credentials data encapsulated in CredentialsData and merge/combine them
// into a CredentialsData that contains the individual Credentials data.
// Eg. a := CredentialsData{CanvasCredentials}
//
//	a.merge(CredentialsData{LuminusCredentials}) // a == CredentialsData{CanvasCredentials, LuminusCredentials}
func (t *CredentialsData) Merge(t2 CredentialsData) {
	if t.CanvasCredentials == (CanvasCredentials{}) {
		t.CanvasCredentials = t2.CanvasCredentials
	}

	if t.LuminusCredentials == (LuminusCredentials{}) {
		t.LuminusCredentials = t2.LuminusCredentials
	}
}

// JwtExpiredError struct contains the JwtExpiredError which will be thrown when the JWT data has expired.
type JwtExpiredError struct{}

// JwtExpiredError error to be thrown when the JWT data has expired.
func (e *JwtExpiredError) Error() string {
	return "JwtExpiredError: JWT token has expired."
}
