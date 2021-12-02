package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Grade struct is the datapack for containing details about every Grade in a module
type Grade struct {
	Name        string
	Marks       float64
	MaxMarks    float64
	Comments    string
	LastUpdated int64
}

const GRADE_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/gradebook/?populate=scores&ParentID=%s"

// Retrieves all grades for a particular module represented by moduleCode.
func (req Request) GetGrades(moduleCode string) ([]Grade, error) {
	var grades []Grade

	request, err := http.NewRequest("GET", fmt.Sprintf(req.Url, moduleCode), nil)

	if err != nil {
		return grades, err
	}

	request.Header.Add("Authorization", "Bearer "+req.JwtToken)

	cl := &http.Client{}
	response, err := cl.Do(request)

	if err != nil {
		return grades, err
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return grades, err
	}

	var obj RawResponse                                //variable which holds the raw data
	json.Unmarshal([]byte(string([]byte(body))), &obj) //Converting from byte to struct

	for _, content := range obj.Data {

		if _, ok := content["access"]; ok { // only grades that can be accessed will be placed in grades slice
			scoreDetail := make(map[string]interface{})
			if len(content["scores"].([]interface{})) > 0 {
				scoreDetail = (content["scores"].([]interface{})[0]).(map[string]interface{})
			}
			testName := content["name"].(string)
			mark := -1.0
			if _, exists := scoreDetail["finalMark"]; exists {
				mark = scoreDetail["finalMark"].(float64)
			}
			maxMark := content["maxMark"].(float64)
			remark := ""
			if _, exists := scoreDetail["remark"]; exists {
				remark = scoreDetail["remark"].(string)
			}
			lastUpdated := int64(-1)
			if _, exists := scoreDetail["lastUpdatedDate"]; exists {
				lastUpdatedTime, err := time.Parse(time.RFC3339, scoreDetail["lastUpdatedDate"].(string))
				if err != nil {
					return grades, err
				}
				lastUpdated = lastUpdatedTime.Unix()
			}

			newStruct := Grade{
				testName,
				mark,
				maxMark,
				remark,
				lastUpdated,
			}
			grades = append(grades, newStruct)
		}
	}
	return grades, nil
}
