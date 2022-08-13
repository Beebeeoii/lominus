package interfaces

// CanvasFileObject depicts the actual object return from Canvas.
// There are more fields being returned by Canvas, but these are just the
// relevant ones as of now.
type CanvasFileObject struct {
	Id   int    `json:"id"`
	Name string `json:"filename"`
	// DisplayName is the name that is seen on Canvas Web. It can differ
	// from the actual name of the file being uploaded.
	// Using DisplayName would be more accurate as Professors might
	// set the DisplayName to contain more information for the File.
	// For eg. filename can be "Tutorial1.pdf" but DisplayName can be
	// "Tutorial1_HW1.pdf" to show that Tutorial 1 is to be submitted as
	// a graded Homework.
	DisplayName   string `json:"display_name"`
	UUID          string `json:"uuid"`
	Url           string `json:"url"`
	HiddenForUser bool   `json:"hidden_for_user"`
	LastUpdated   string `json:"modified_at"`
}

// LuminusFileObject depicts the actual object return from Luminus.
// There are more fields being returned by Luminus, but these are just the
// relevant ones as of now.
type LuminusFileObject struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	LastUpdated string `json:"lastUpdatedDate" mapstructure:"lastUpdatedDate"`
}
