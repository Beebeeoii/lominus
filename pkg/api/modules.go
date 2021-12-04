package api

// Module struct is the datapack for containing details about every module
type Module struct {
	Id           string
	Name         string
	ModuleCode   string
	CreatorName  string
	CreatorEmail string
	Period       string
}

const MODULE_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/module/?populate=Creator%2CtermDetail%2CisMandatory"

func (req Request) GetModules() ([]Module, error) {
	var modules []Module
	rawResponse := RawResponse{}
	err := req.GetRawResponse(&rawResponse)
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
				ModuleCode:   content["name"].(string),
				CreatorName:  content["creatorName"].(string),
				CreatorEmail: content["creatorEmail"].(string),
				Period:       termDetail["description"].(string),
			}
			modules = append(modules, module)
		}
	}

	return modules, nil
}
