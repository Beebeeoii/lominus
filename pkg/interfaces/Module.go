// Package interfaces provide the fundamental blueprint for how each object
// looks like.
package interfaces

// CanvasModuleObject depicts the actual object return from Canvas.
// There are more fields being returned by Canvas, but these are just the
// relevant ones as of now.
type CanvasModuleObject struct {
	Id                       int    `json:"id"`
	UUID                     string `json:"uuid"`
	Name                     string `json:"name"`
	ModuleCode               string `json:"course_code"`
	IsAccessRestrictedByDate bool   `json:"access_restricted_by_date"`
}

// LuminusModuleObject depicts the actual object return from Luminus.
// There are more fields being returned by Luminus, but these are just the
// relevant ones as of now.
// For more details on what mapstructure means: https://github.com/mitchellh/mapstructure
type LuminusModuleObject struct {
	Id                 string `json:"id"`
	Name               string `json:"courseName" mapstructure:"courseName"`
	ModuleCode         string `json:"name" mapstructure:"name"`
	IsCourseSearchable bool   `json:"courseSearchable" mapstructure:"courseSearchable"`
	IsPublished        bool   `json:"publish" mapstructure:"publish"`
}
