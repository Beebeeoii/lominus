package interfaces

// TODO Documentation
type CanvasModuleObject struct {
	Id                     int    `json:"id"`
	UUID                   string `json:"uuid"`
	Name                   string `json:"name"`
	ModuleCode             string `json:"course_code"`
	AccessRestrictedByDate bool   `json:"access_restricted_by_date"`
}
