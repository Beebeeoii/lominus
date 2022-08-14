// Package interfaces provide the fundamental blueprint for how each object
// looks like.
package interfaces

// LuminusRawResponse depicts the actual object return from Luminus.
// There are more fields being returned by Luminus, but these are just the
// relevant ones as of now.
type LuminusRawResponse struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Total  int         `json:"total"`
	Offset int         `json:"offset"`
	Data   interface{} `json:"data"`
}
