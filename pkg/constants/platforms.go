// Package constants provides constants such as web endpoints.
package constants

type Platform int

const (
	Canvas Platform = iota
)

// Platforms is a list of available LMS platforms supported by Lominus.
var Platforms = []Platform{
	Canvas,
}
