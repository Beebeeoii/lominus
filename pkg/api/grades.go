// Package api provides functions that link up and communicate with Luminus servers.
package api

import (
	"time"
)

// Grade struct is the datapack for containing details about every Grade in a module.
type Grade struct {
	Name        string
	Marks       float64
	MaxMarks    float64
	Comments    string
	LastUpdated int64
}

const GRADE_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/gradebook/?populate=scores&ParentID=%s"

// GetGrades retrieves all grades for a particular module represented by moduleCode specified in GradeRequest.
// Find out more about GradeRequests under request.go.
func (req GradeRequest) GetGrades() ([]Grade, error) {
	var grades []Grade

	rawResponse := RawResponse{}
	err := req.Request.GetRawResponse(&rawResponse)
	if err != nil {
		return grades, err
	}

	for _, content := range rawResponse.Data {
		if _, exists := content["access"]; exists { // only grades that can be accessed will be placed in grades slice
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

			grade := Grade{
				testName,
				mark,
				maxMark,
				remark,
				lastUpdated,
			}
			grades = append(grades, grade)
		}
	}
	return grades, nil
}
