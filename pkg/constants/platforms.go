// TODO Documentations
package constants

type Platform int

const (
	Canvas Platform = iota
	Luminus
)

var Platforms = []Platform{
	Canvas,
	Luminus,
}
