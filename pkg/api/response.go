// Package api provides functions that link up and communicate with Luminus servers.
package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// RawResponse struct is the datapack for containing API response raw data.
type RawResponse struct {
	Status string                   `json:"status"`
	Code   int                      `json:"code"`
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Data   []map[string]interface{} `json:"data"`
}

// DownloadResponse struct is the datapack for containing API download response raw data.
type DownloadResponse struct {
	Code        int    `json:"code"`
	Status      string `json:"status"`
	DownloadUrl string `json:"data"`
}

// GetRawResponse sends the HTTP request and marshals it into the pointer provided.
// Argument provided must be a pointer.
func (req Request) GetRawResponse(res interface{}) error {

	request, err := http.NewRequest("GET", req.Url, nil)
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", "Bearer "+req.JwtToken)

	cl := &http.Client{}

	response, err := cl.Do(request)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	json.Unmarshal(body, res)

	return err
}

// isResponseValid is a helper function that checks if a Folder/File response is valid.
// It checks if the response contains the required fields required.
func IsResponseValid(fieldsRequired []string, response map[string]interface{}) bool {
	isValid := true
	for _, field := range fieldsRequired {
		_, exists := response[field]

		if !exists {
			isValid = false
			break
		}
	}

	return isValid
}
