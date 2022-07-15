package api

import (
	"testing"
)

func TestGetCanvasModules(t *testing.T) {
	moduleRequest, modReqErr := BuildCanvasModuleRequest()
	if modReqErr != nil {
		t.Fatal(modReqErr)
	}
	modules, err := moduleRequest.GetCanvasModules()
	if err != nil {
		t.Fatal(err)
	}

	for _, module := range modules {
		if module.Id == "" {
			t.Fatal("Module Id is empty.")
		}

		if module.Name == "" {
			t.Fatal("Module Name is empty.")
		}

		if module.ModuleCode == "" {
			t.Fatal("Module ModuleCode is empty.")
		}
	}
}
