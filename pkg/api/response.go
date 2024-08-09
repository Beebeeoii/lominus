// Package api provides functions that link up and communicate with LMS servers,
// such as Canvas.
package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// GetRawResponse sends the HTTP request and marshals it into the pointer provided.
// Argument provided must be a pointer.
func (req Request) GetRawResponse(res interface{}) error {
	request, err := http.NewRequest("GET", req.Url.Url, nil)
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", "Bearer "+req.Token)

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
