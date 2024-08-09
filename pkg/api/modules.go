// Package api provides functions that link up and communicate with LMS servers,
// such as Canvas.
package api

import (
	"errors"
	"strconv"
	"strings"

	"github.com/beebeeoii/lominus/pkg/constants"
	"github.com/beebeeoii/lominus/pkg/interfaces"
)

// Module struct is the datapack for containing details about every module
type Module struct {
	Id           string
	Name         string
	ModuleCode   string
	IsAccessible bool
}

// GetModules retrieves all the modules being taken by the user on the specified LMS
// via a ModulesRequest.
func (modulesRequest ModulesRequest) GetModules() ([]Module, error) {
	modules := []Module{}
	if modulesRequest.Request.Token == "" {
		return modules, nil
	}

	switch moduleDataType := modulesRequest.Request.Url.Platform; moduleDataType {
	case constants.Canvas:
		response := []interfaces.CanvasModuleObject{}
		reqErr := modulesRequest.Request.Send(&response)

		if reqErr != nil {
			return modules, reqErr
		}

		for _, moduleObject := range response {
			modules = append(modules, Module{
				Id:           strconv.Itoa(moduleObject.Id),
				Name:         moduleObject.Name,
				ModuleCode:   cleanseModuleCode(moduleObject.ModuleCode),
				IsAccessible: !moduleObject.IsAccessRestrictedByDate,
			})
		}
	default:
		return modules, errors.New("modulesRequest.Request.Url.Platform is not available")
	}

	return modules, nil
}

// cleanseModuleCode is a helper function that replaces all instances of "/" with "-".
// This is necessary for multi-coded modules like ST2131/MA2216.
func cleanseModuleCode(code string) string {
	return strings.Replace(code, "/", "-", -1)
}
