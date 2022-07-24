package auth

// TODO Documentation
type CanvasTokenData struct {
	CanvasApiToken string
}

// TODO Documentation
func (canvasTokenData CanvasTokenData) Save(tokensPath string) error {
	return saveTokenData(tokensPath, TokensData{
		CanvasToken: canvasTokenData,
	})
}
