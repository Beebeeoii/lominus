package interfaces

type LuminusRawResponse struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Total  int         `json:"total"`
	Offset int         `json:"offset"`
	Data   interface{} `json:"data"`
}
