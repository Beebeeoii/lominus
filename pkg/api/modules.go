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
const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0"

func GetModules(token string) (string, error) {
	var modules string

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

	return string([]byte(body)), nil
}
