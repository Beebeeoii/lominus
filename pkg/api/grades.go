// Package api provides functions that link up and communicate with LMS servers,
// such as Canvas and Luminus (probably removed in near future).
package api

import (
	"time"
)

// Grade struct is the datapack for containing details about every Grade in a module.
type Grade struct {
	Module      Module
	Name        string
	Marks       float64
	MaxMarks    float64
	Comments    string
	LastUpdated int64
}

const GRADE_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/gradebook/?populate=scores&ParentID=%s"

// getGradeFieldsRequired is a helper function that returns a constant array with fields that a Grade response
// returned by Luminus needs.
func getGradeFieldsRequired() []string {
	return []string{"access", "scores", "name", "maxMark"}
}

// getScoreDetailFieldsRequired is a helper function that returns a constant array with fields that a Grade["scores"] element
// returned by Luminus needs.
func getScoreDetailFieldsRequired() []string {
	return []string{"finalMark", "lastUpdatedDate"}
}

// GetGrades retrieves all grades for a particular module represented by moduleCode specified in GradeRequest.
// Find out more about GradeRequests under request.go.
// Note that Grades API works only for Luminus and not supported for Canvas yet.
func (req GradeRequest) GetGrades() ([]Grade, error) {
	var grades []Grade

	rawResponse := RawResponse{}
	err := req.Request.GetRawResponse(&rawResponse)
	if err != nil {
		return grades, err
	}

	for _, content := range rawResponse.Data {
		if !IsResponseValid(getGradeFieldsRequired(), content) {
			continue
		}

		if _, exists := content["access"]; exists {
			scoreField := content["scores"].([]interface{})
			if len(scoreField) == 0 {
				continue
			}

			scoreDetail := (scoreField[0]).(map[string]interface{})
			if !IsResponseValid(getScoreDetailFieldsRequired(), scoreDetail) {
				continue
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
				req.Module,
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
