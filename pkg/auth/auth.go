// Package auth provides functions that link up and communicate with Luminus authentication server.
package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	appAuth "github.com/beebeeoii/lominus/internal/app/auth"
	file "github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/internal/lominus"
)

// JsonResponse struct is the datapack for containing API authentication response raw data.
type JsonResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// JwtData struct is the datapack that describes the user's JWT data.
type JwtData struct {
	JwtToken string
	Expiry   int64
}

// Credentials struct is the datapack that describes the user's credentials.
type Credentials struct {
	Username string
	Password string
}

const CREDENTIALS_FILE_NAME = lominus.CREDENTIALS_FILE_NAME

const CODE_URL = "https://vafs.nus.edu.sg/adfs/oauth2/authorize?response_type=code&client_id=E10493A3B1024F14BDC7D0D8B9F649E9-234390&state=V6E9kYSq3DDQ72fSZZYFzLNKFT9dz38vpoR93IL8&redirect_uri=https://luminus.nus.edu.sg/auth/callback&scope=&resource=sg_edu_nus_oauth&nonce=V6E9kYSq3DDQ72fSZZYFzLNKFT9dz38vpoR93IL8"

const JWT_URL = "https://luminus.nus.edu.sg/v2/api/login/adfstoken"
const REDIRECT_URI = "https://luminus.nus.edu.sg/auth/callback"
const RESOURCE = "sg_edu_nus_oauth"
const CLIENT_ID = "E10493A3B1024F14BDC7D0D8B9F649E9-234390"
const GRANT_TYPE = "authorization_code"

const CONTENT_TYPE = "application/x-www-form-urlencoded"
const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0"
const POST = "POST"
const AUTH_METHOD = "FormsAuthentication"

const EXPIRY_HOURS = 1

// RetrieveJwtToken takes in the user's Credentials and return a JWT token issued by Luminus authentication server.
func RetrieveJwtToken(credentials Credentials, save bool) (string, error) {
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
	codeReq, codeReqErr := http.NewRequest(POST, CODE_URL, strings.NewReader(codeBody.Encode()))

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
	jwtBody.Set("redirect_uri", REDIRECT_URI)
	jwtBody.Set("code", code)
	jwtBody.Set("resource", RESOURCE)
	jwtBody.Set("client_id", CLIENT_ID)
	jwtBody.Set("grant_type", GRANT_TYPE)
	jwtReq, jwtReqErr := http.NewRequest(POST, JWT_URL, strings.NewReader(jwtBody.Encode()))
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
		return jwtToken, saveJwtData(appAuth.GetJwtPath(), jwtToken)
	}

	return jwtToken, nil
}

// saveJwtData saves the user's JWT data to local storage for future use.
func saveJwtData(jwtPath string, jwtToken string) error {
	jwtData := JwtData{jwtToken, time.Now().Add(time.Hour * 24).Unix()}
	return file.EncodeStructToFile(jwtPath, jwtData)
}

// LoadJwtData loads the user's JWT data from local storage.
func LoadJwtData(jwtPath string) (JwtData, error) {
	jwtData := JwtData{}
	if !file.Exists(jwtPath) {
		return jwtData, &file.FileNotFoundError{FileName: jwtPath}
	}
	err := file.DecodeStructFromFile(jwtPath, &jwtData)

	return jwtData, err
}

// SaveCredentials saves the user's Credentials for future authentications or renewals of JWT data.
func SaveCredentials(credentialsPath string, credentials Credentials) error {
	return file.EncodeStructToFile(credentialsPath, credentials)
}

// LoadCredentials loads the user's Credentials data from local storage.
func LoadCredentials(credentialsPath string) (Credentials, error) {
	credentials := Credentials{}
	if !file.Exists(credentialsPath) {
		return credentials, &file.FileNotFoundError{FileName: credentialsPath}
	}
	err := file.DecodeStructFromFile(credentialsPath, &credentials)

	return credentials, err
}

// IsExpired is a util function that checks if the user's JWT data has expired.
func (jwtData JwtData) IsExpired() bool {
	expiry := time.Unix(jwtData.Expiry, 0)
	return time.Until(expiry).Hours() <= EXPIRY_HOURS
}

// JwtExpiredError struct contains the JwtExpiredError which will be thrown when the JWT data has expired.
type JwtExpiredError struct{}

// JwtExpiredError error to be thrown when the JWT data has expired.
func (e *JwtExpiredError) Error() string {
	return "JwtExpiredError: JWT token has expired."
}
