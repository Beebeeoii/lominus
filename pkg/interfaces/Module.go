// Package interfaces provide the fundamental blueprint for how each object
// looks like.
package interfaces

// CanvasModuleObject depicts the actual object return from Canvas.
// There are more fields being returned by Canvas, but these are just the
// relevant ones as of now.
type CanvasModuleObject struct {
	Id                       int    `json:"id"`
	Name                     string `json:"originalName"`
	ModuleCode               string `json:"courseCode"`
	IsAccessRestrictedByDate bool   // Automatically false since is still accessible by the Canvas dashboard
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
