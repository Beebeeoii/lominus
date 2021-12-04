package api

// Module struct is the datapack for containing details about every module
type Module struct {
	Name         string
	ModuleCode   string
	Id           string
	CreatorName  string
	CreatorEmail string
	Period       string
}

const MODULE_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/module/?populate=Creator%2CtermDetail%2CisMandatory"

func (req Request) GetModules() ([]Module, error) {

	var modules []Module //Initialise slice to capture of modules

	RawResponse, err := req.GetRawResponse()

	if err != nil {
		return modules, err
	}

	for _, content := range RawResponse.Data {
		_, found1 := content["access"]
		found2, _ := content["usedNusCalendar"].(bool)
		if found1 && !found2 { // only modules that can be accessed will be placed in modules slice

			termDetail := content["termDetail"].(map[string]interface{}) //getting inner map
			newStruct := Module{
				content["courseName"].(string),
				content["name"].(string),
				content["id"].(string),
				content["creatorName"].(string),
				content["creatorEmail"].(string),
				termDetail["description"].(string),
			}
			modules = append(modules, newStruct)
		}
	}
	return modules, nil
}
