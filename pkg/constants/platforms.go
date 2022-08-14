// Package constants provides constants such as web endpoints.
package constants

type Platform int

// This is an enum.
// Eg. Canvas = 0, Luminus = 1, ...
const (
	Canvas Platform = iota
	Luminus
)

// Platforms is a list of available LMS platforms supported by Lominus.
// It is used to differentiate the various LMS due to the possible different ways
// each platform works.
var Platforms = []Platform{
	Canvas,
	Luminus,
}
