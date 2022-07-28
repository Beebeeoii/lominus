package interfaces

// TODO Documentation
type CanvasModuleObject struct {
	Id                       int    `json:"id"`
	UUID                     string `json:"uuid"`
	Name                     string `json:"name"`
	ModuleCode               string `json:"course_code"`
	IsAccessRestrictedByDate bool   `json:"access_restricted_by_date"`
}

// TODO Documentation
type LuminusModuleObject struct {
	Id                 string `json:"id"`
	Name               string `json:"courseName" mapstructure:"courseName"`
	ModuleCode         string `json:"name" mapstructure:"name"`
	IsCourseSearchable bool   `json:"courseSearchable" mapstructure:"courseSearchable"`
	IsPublished        bool   `json:"publish" mapstructure:"publish"`
}
