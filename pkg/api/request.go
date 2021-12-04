package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Request struct {
	Url       string
	JwtToken  string
	UserAgent string
}

const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0"

func (req Request) GetRawResponse() (RawResponse, error) {
	var raw RawResponse //variable which holds the raw data

	request, err := http.NewRequest("GET", req.Url, nil)

	if err != nil {
		return raw, err
	}

	request.Header.Add("Authorization", "Bearer "+req.JwtToken)

	cl := &http.Client{}
	response, err := cl.Do(request)

	if err != nil {
		return raw, err
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return raw, err
	}
	json.Unmarshal([]byte(string([]byte(body))), &raw) //Converting from byte to struct
	return raw, err
}
