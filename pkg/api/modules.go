// Package api provides functions that link up and communicate with Luminus servers.
package api

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/beebeeoii/lominus/pkg/constants"
	"github.com/beebeeoii/lominus/pkg/interfaces"
	"github.com/mitchellh/mapstructure"
)

// Module struct is the datapack for containing details about every module
type Module struct {
	Id           string
	Name         string
	ModuleCode   string
	IsAccessible bool
}

const MODULE_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/module/?populate=Creator,termDetail,isMandatory"

// TODO Documentation
func (modulesRequest ModulesRequest) GetModules() ([]Module, error) {
	modules := []Module{}

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
	case constants.Luminus:
		modulesData := []interfaces.LuminusModuleObject{}

		response := interfaces.LuminusRawResponse{}
		reqErr := modulesRequest.Request.Send(&response)

		if reqErr != nil {
			return modules, reqErr
		}

		data := reflect.ValueOf(response.Data)
		if data.Kind() == reflect.Slice {
			for i := 0; i < data.Len(); i++ {
				moduleData := interfaces.LuminusModuleObject{}
				decodeErr := mapstructure.Decode(data.Index(i).Interface(), &moduleData)
				if decodeErr != nil {
					return modules, decodeErr
				}
				modulesData = append(modulesData, moduleData)
			}
		}

		for _, moduleObject := range modulesData {
			modules = append(modules, Module{
				Id:           moduleObject.Id,
				Name:         moduleObject.Name,
				ModuleCode:   cleanseModuleCode(moduleObject.ModuleCode),
				IsAccessible: moduleObject.IsCourseSearchable && moduleObject.IsPublished,
			})
		}
	default:
		return modules, errors.New("modulesRequest.Request.Url.Platform is not available")
	}

	return modules, nil
}

// TODO Documentation
func cleanseModuleCode(code string) string {
	return strings.Replace(code, "/", "-", -1)
}
