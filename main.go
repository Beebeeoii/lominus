package main

import (
	"fmt"
	"log"

	"github.com/beebeeoii/lominus/pkg/api"
)

func main() {
	/*
		appInitErr := app.Init()
		if appInitErr != nil {
			log.Fatalln(appInitErr)
		}

		lock := fslock.New(appLock.GetLockPath())
		lockErr := lock.TryLock()

		if lockErr != nil {
			log.Fatalln(lockErr)
		}
		defer lock.Unlock()

		uiInitErr := ui.Init()
		if uiInitErr != nil {
			log.Fatalln(uiInitErr)
		}
	*/

	moduleReq, err := api.BuildModuleRequest()

	if err != nil {
		log.Fatalln(err)
	}

	modules, err := moduleReq.GetModules()

	if err != nil {
		log.Fatalln(err)
	}

	for _, module := range modules {
		if module.ModuleCode != "IS4010" {
			continue
		}
		docReq, err := api.BuildDocumentRequest(module, api.GET_ALL_FILES)

		if err != nil {
			log.Fatalln(err)
		}

		file, err := docReq.GetAllFiles()

		if err != nil {
			log.Fatalln(err)
		}

		for _, f := range file {
			for _, fi := range f.Ancestors {
				fmt.Println(fi)
			}
			fmt.Println(f.Name)
			fmt.Println()
		}

	}

}
