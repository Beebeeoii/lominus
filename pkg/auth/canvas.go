package auth

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
