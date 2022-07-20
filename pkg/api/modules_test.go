package api

import (
	"testing"

	"github.com/beebeeoii/lominus/pkg/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestGetCanvasModules(t *testing.T) {
	mockModuleResponse := CanvasModuleResponse{
		Data: []interfaces.CanvasModuleObject{
			{
				Id:                     12345,
				UUID:                   "",
				Name:                   "",
				ModuleCode:             "",
				AccessRestrictedByDate: true,
			},
			{
				Id:                     45678,
				UUID:                   "F9SdBDH1NTUJKB0lH2aKsABzlbAbLWtpLIfkAe93",
				Name:                   "CP3107 Computing for Voluntary Welfare Organisations [2130]",
				ModuleCode:             "CP3107",
				AccessRestrictedByDate: false,
			},
		},
	}

	modules, err := mockModuleResponse.GetCanvasModules()

	assert.Nil(t, err)
	assert.Equal(t, modules[0].Id, "12345")
	assert.Equal(t, modules[0].Name, "")
	assert.Equal(t, modules[0].ModuleCode, "")
	assert.Equal(t, modules[1].Id, "45678")
	assert.Equal(t, modules[1].Name, "CP3107 Computing for Voluntary Welfare Organisations [2130]")
	assert.Equal(t, modules[1].ModuleCode, "CP3107")
}
