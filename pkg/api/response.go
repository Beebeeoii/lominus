package api

// raw struct is the datapack for containing api raw data
type RawResponse struct {
	Status string                   `json:"status"`
	Code   int                      `json:"code"`
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Data   []map[string]interface{} `json:"data"`
}
