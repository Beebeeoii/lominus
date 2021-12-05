package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// raw struct is the datapack for containing api raw data
type RawResponse struct {
	Status string                   `json:"status"`
	Code   int                      `json:"code"`
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Data   []map[string]interface{} `json:"data"`
}

type DownloadResponse struct {
	Code   int    `json:"code"`
	Data   string `json:"data"`
	Status string `json:"status"`
}

func (req Request) GetRawResponse(res interface{}) error {

	request, err := http.NewRequest("GET", req.Url, nil)
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", "Bearer "+req.JwtToken)

	cl := &http.Client{}

	response, err := cl.Do(request)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	json.Unmarshal(body, res)

	return err
}
