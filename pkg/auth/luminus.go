// Package auth provides functions that link up and communicate with LMS (Luminus/Canvas)
// authentication server.
package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
	"github.com/beebeeoii/lominus/pkg/constants"
)

const (
	RESOURCE     = "sg_edu_nus_oauth"
	CLIENT_ID    = "E10493A3B1024F14BDC7D0D8B9F649E9-234390"
	GRANT_TYPE   = "authorization_code"
	EXPIRY_HOURS = 1
)

// LuminusTokenData is a struct that encapsulates the token required for authentication.
// In this case, it is the JwtToken which is a string, and the corresponding expiry timestamp
// in milliseconds.
type LuminusTokenData struct {
	JwtToken  string
	JwtExpiry int64
}

// LuminusCredentials is a struct that encapsulates the credentials required for authentication.
// In this case, it is the username and password.
type LuminusCredentials struct {
	Username string
	Password string
}

// RetrieveJwtToken takes in the user's Credentials and return a JWT token issued by
// Luminus authentication server.
// TODO Refactor to use api.Send()
func RetrieveJwtToken(credentials LuminusCredentials, save bool) (string, error) {
	var jwtToken string
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	codeBody := url.Values{}
	codeBody.Set("UserName", credentials.Username)
	codeBody.Set("Password", credentials.Password)
	codeBody.Set("AuthMethod", AUTH_METHOD)
	codeReq, codeReqErr := http.NewRequest(
		POST,
		constants.LUMINUS_AUTH_CODE_ENDPOINT,
		strings.NewReader(codeBody.Encode()),
	)

	if codeReqErr != nil {
		return jwtToken, codeReqErr
	}

	codeReq.Header.Add("Content-Type", CONTENT_TYPE)
	codeReq.Header.Add("User-Agent", USER_AGENT)

	codeRes1, codeRes1Err := client.Do(codeReq)
	if codeRes1Err != nil {
		return jwtToken, codeRes1Err
	}

	for _, cookie := range codeRes1.Cookies() {
		codeReq.AddCookie(cookie)
	}

	codeRes2, codeRes2Err := client.Do(codeReq)
	if codeRes2Err != nil {
		return jwtToken, codeRes2Err
	}

	indexStart := strings.Index(codeRes2.Header.Get("Location"), "code=") + 5
	indexEnd := strings.Index(codeRes2.Header.Get("Location"), "&state=")
	code := codeRes2.Header.Get("Location")[indexStart:indexEnd]

	jwtBody := url.Values{}
	jwtBody.Set("redirect_uri", constants.LUMINUS_AUTH_REDIRECT_ENDPOINT)
	jwtBody.Set("code", code)
	jwtBody.Set("resource", RESOURCE)
	jwtBody.Set("client_id", CLIENT_ID)
	jwtBody.Set("grant_type", GRANT_TYPE)
	jwtReq, jwtReqErr := http.NewRequest(
		POST,
		constants.LUMINUS_AUTH_JWT_ENDPOINT,
		strings.NewReader(jwtBody.Encode()),
	)
	if jwtReqErr != nil {
		return jwtToken, jwtReqErr
	}
	jwtReq.Header.Add("Content-Type", CONTENT_TYPE)
	jwtReq.Header.Add("User-Agent", USER_AGENT)

	jwtRes, jwtResErr := client.Do(jwtReq)
	if jwtResErr != nil {
		return jwtToken, jwtResErr
	}

	body, err := ioutil.ReadAll(jwtRes.Body)
	if err != nil {
		return jwtToken, err
	}

	var jsonResponse JsonResponse
	toJsonErr := json.Unmarshal(body, &jsonResponse)
	if toJsonErr != nil {
		return jwtToken, toJsonErr
	} else {
		jwtToken = jsonResponse.AccessToken
	}

	if save {
		jwtPath, getJwtPathErr := appAuth.GetTokensPath()
		if getJwtPathErr != nil {
			return jwtToken, getJwtPathErr
		}

		return jwtToken, LuminusTokenData{
			JwtToken:  jwtToken,
			JwtExpiry: time.Now().Add(time.Hour * 24).Unix(),
		}.Save(jwtPath)
	}

	return jwtToken, nil
}

// Save takes in the LuminusTokenData and saves it locally with the path provided as arguments.
func (luminusTokenData LuminusTokenData) Save(tokensPath string) error {
	return saveTokenData(tokensPath, TokensData{
		LuminusToken: luminusTokenData,
	})
}

// Save takes in the LuminusCredentials and saves it locally with the path provided as arguments.
func (credentials LuminusCredentials) Save(credentialsPath string) error {
	return saveCredentialsData(credentialsPath, CredentialsData{
		LuminusCredentials: credentials,
	})
}

// IsExpired is a helper function that checks if the user's JWT data has expired.
func (luminusTokenData LuminusTokenData) IsExpired() bool {
	expiry := time.Unix(luminusTokenData.JwtExpiry, 0)
	return time.Until(expiry).Hours() <= EXPIRY_HOURS
}
