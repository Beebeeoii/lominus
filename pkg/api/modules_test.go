package api

import (
	"testing"

	"github.com/beebeeoii/lominus/pkg/interfaces"
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
	if err != nil {
		t.Fatal(err)
	}

	if modules[0].Id != "12345" {
		t.Fatalf("Result: %s | Expected: %s", modules[0].Id, "12345")
	}

	if modules[0].Name != "" {
		t.Fatalf("Result: %s | Expected: %s", modules[0].Name, "")
	}

	if modules[0].ModuleCode != "" {
		t.Fatalf("Result: %s | Expected: %s", modules[0].ModuleCode, "")
	}

	if modules[1].Id != "45678" {
		t.Fatalf("Result: %s | Expected: %s", modules[1].Id, "45678")
	}

	if modules[1].Name != "CP3107 Computing for Voluntary Welfare Organisations [2130]" {
		t.Fatalf("Result: %s | Expected: %s", modules[1].Name,
			"CP3107 Computing for Voluntary Welfare Organisations [2130]")
	}

	if modules[1].ModuleCode != "CP3107" {
		t.Fatalf("Result: %s | Expected: %s", modules[1].ModuleCode, "CP3107")
	}
}
