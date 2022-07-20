package interfaces

// TODO Documentation
type CanvasFileObject struct {
	Id            int    `json:"id"`
	Name          string `json:"filename"`
	UUID          string `json:"uuid"`
	Url           string `json:"url"`
	HiddenForUser bool   `json:"hidden_for_user"`
	LastUpdated   string `json:"updated_at"`
}
