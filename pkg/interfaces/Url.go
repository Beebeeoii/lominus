// Package interfaces provide the fundamental blueprint for how each object
// looks like.
package interfaces

import "github.com/beebeeoii/lominus/pkg/constants"

// Url is a struct that encapsulates a Url object.
// It contains the address and the LMS platform which the address belongs to.
type Url struct {
	Url      string
	Platform constants.Platform
}
