// Package api provides functions that link up and communicate with LMS servers,
// such as Canvas.
package api

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
