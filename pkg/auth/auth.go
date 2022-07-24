// Package auth provides functions that link up and communicate with Luminus authentication server.
package auth

import (
	file "github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/internal/lominus"
)

// JsonResponse struct is the datapack for containing API authentication response raw data.
type JsonResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// TokenData struct is the datapack that describes the user's tokens data.
type TokensData struct {
	CanvasToken  CanvasTokenData
	LuminusToken LuminusTokenData
}

const CREDENTIALS_FILE_NAME = lominus.CREDENTIALS_FILE_NAME

const CONTENT_TYPE = "application/x-www-form-urlencoded"
const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0"
const POST = "POST"
const AUTH_METHOD = "FormsAuthentication"

// saveTokenData saves the user's Tokens data to local storage for future use.
func saveTokenData(tokensPath string, tokensData TokensData) error {
	if file.Exists(tokensPath) {
		localTokensData, err := LoadTokensData(tokensPath)
		if err != nil {
			return err
		}

		tokensData.Merge(localTokensData)
	}

	return file.EncodeStructToFile(tokensPath, tokensData)
}

// LoadTokensData loads the user's Tokens data from local storage.
func LoadTokensData(tokensPath string) (TokensData, error) {
	tokensData := TokensData{}
	if !file.Exists(tokensPath) {
		return tokensData, &file.FileNotFoundError{FileName: tokensPath}
	}
	err := file.DecodeStructFromFile(tokensPath, &tokensData)

	return tokensData, err
}

// TODO Documentation
func (t *TokensData) Merge(t2 TokensData) {
	if t.CanvasToken == (CanvasTokenData{}) {
		t.CanvasToken = t2.CanvasToken
	}

	if t.LuminusToken == (LuminusTokenData{}) {
		t.LuminusToken = t2.LuminusToken
	}
}

// JwtExpiredError struct contains the JwtExpiredError which will be thrown when the JWT data has expired.
type JwtExpiredError struct{}

// JwtExpiredError error to be thrown when the JWT data has expired.
func (e *JwtExpiredError) Error() string {
	return "JwtExpiredError: JWT token has expired."
}
