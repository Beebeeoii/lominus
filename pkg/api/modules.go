package api

import (
	"io/ioutil"
	"net/http"
)

type module struct {
	Name         string
	ModuleCode   string
	Id           string
	CreatorName  string
	CreatorEmail string
}

const MODULE_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/module/?populate=Creator%2CtermDetail%2CisMandatory"

func (req Request) GetModules() ([]module, error) {
	var modules []module //Initialise slice to capture of modules


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

	return string([]byte(body)), nil
}
