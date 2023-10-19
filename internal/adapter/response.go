package adapter

import "encoding/json"

type ErrorResponse struct {
	Errors string `json:"errors"`
}

func (e *ErrorResponse) ToJson() []byte {
	result, _ := json.Marshal(e)
	return result
}
