package constants

type Platform int

type HasPlatform interface {
	GetPlatform()
}

const (
	Canvas Platform = iota
	Luminus
)

var Platforms = []Platform{
	Canvas,
	Luminus,
}
