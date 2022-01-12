// Package api provides functions that link up and communicate with Luminus servers.
package api

import "strings"

// Module struct is the datapack for containing details about every module
type Module struct {
	Id           string
	Name         string
	ModuleCode   string
	CreatorName  string
	CreatorEmail string
	Period       string
}

const MODULE_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/module/?populate=Creator,termDetail,isMandatory"

// GetModules retrieves all modules that are taken by the user using a ModuleRequest.
// Find out more about ModuleRequests under request.go.
func (req ModuleRequest) GetModules() ([]Module, error) {
	var modules []Module

	rawResponse := RawResponse{}
	err := req.Request.GetRawResponse(&rawResponse)
	if err != nil {
		return modules, err
	}

	for _, content := range rawResponse.Data {
		_, accessible := content["access"]
		if accessible {
			termDetail := content["termDetail"].(map[string]interface{})
			module := Module{
				Id:           content["id"].(string),
				Name:         content["courseName"].(string),
				ModuleCode:   strings.Replace(content["name"].(string), "/", "-", -1), // for multi-coded modules that uses '/' as a separator
				CreatorName:  content["creatorName"].(string),
				CreatorEmail: content["creatorEmail"].(string),
				Period:       termDetail["description"].(string),
			}
			modules = append(modules, module)
		}
	}

	return modules, nil
}
