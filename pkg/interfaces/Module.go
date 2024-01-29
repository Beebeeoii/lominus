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
