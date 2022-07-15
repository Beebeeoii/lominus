// Package api provides functions that link up and communicate with Luminus servers.
package api

import (
	"strconv"
	"strings"
)

// Module struct is the datapack for containing details about every module
type Module struct {
	Id         string
	Name       string
	ModuleCode string
}

const MODULE_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/module/?populate=Creator,termDetail,isMandatory"

// getModuleFieldsRequired is a helper function that returns a constant array with fields that a Module response
// returned by Luminus needs.
func getModuleFieldsRequired() []string {
	return []string{"access", "termDetail", "courseName", "name", "creatorName", "creatorEmail"}
}

// getTermDetailFieldsRequired is a helper function that returns a constant array with fields that a Module["termDetail"]
// returned by Luminus needs.
func getTermDetailFieldsRequired() []string {
	return []string{"description"}
}

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
		if !IsResponseValid(getModuleFieldsRequired(), content) {
			continue
		}

		_, accessible := content["access"]
		if accessible {
			termDetail := content["termDetail"].(map[string]interface{})
			if !IsResponseValid(getTermDetailFieldsRequired(), termDetail) {
				continue
			}

			module := Module{
				Id:         content["id"].(string),
				Name:       content["courseName"].(string),
				ModuleCode: strings.Replace(content["name"].(string), "/", "-", -1), // for multi-coded modules that uses '/' as a separator
			}
			modules = append(modules, module)
		}
	}

	return modules, nil
}

func (moduleResponse CanvasModuleResponse) GetCanvasModules() ([]Module, error) {
	var modules []Module

	for _, module := range moduleResponse.Data {
		module := Module{
			Id:         strconv.Itoa(module.Id),
			Name:       module.Name,
			ModuleCode: strings.Replace(module.ModuleCode, "/", "-", -1), // for multi-coded modules that uses '/' as a separator
		}
		modules = append(modules, module)
	}

	return modules, nil
}
