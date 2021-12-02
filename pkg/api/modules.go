package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Module struct is the datapack for containing details about every module
type module struct {
	Name         string
	ModuleCode   string
	Id           string
	CreatorName  string
	CreatorEmail string
	period       string
}

// raw struct is the datapack for containing api raw data
type raw struct {
	Status string                   `json:"status"`
	Code   int                      `json:"code"`
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Data   []map[string]interface{} `json:"data"`
}

const MODULE_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/module/?populate=Creator%2CtermDetail%2CisMandatory"
const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0"

func GetModules(token string) ([]module, error) {
	var modules []module //Initialise slice to capture of modules

	req := &Request{MODULE_URL_ENDPOINT, token, USER_AGENT}
	request, err := http.NewRequest("GET", req.Url, nil)

	if err != nil {
		return modules, err
	}

	request.Header.Add("Authorization", "Bearer "+req.JwtToken)

	cl := &http.Client{}
	response, err := cl.Do(request)

	if err != nil {
		return modules, err
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return modules, err
	}

	var obj raw                                        //variable which holds the raw data
	json.Unmarshal([]byte(string([]byte(body))), &obj) //Converting from byte to struct

	for _, content := range obj.Data {

		if _, ok := content["access"]; ok { // only modules that can be accessed will be placed in modules slice

			termDetail := content["termDetail"].(map[string]interface{}) //getting inner map
			newStruct := module{
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
