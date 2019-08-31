package handler

import "encoding/json"

type errorResponse struct {
	Description string `json:"message"`
}

func NewErrorResponse(description string) string {
	err := errorResponse{Description: description}
	b, _ := json.Marshal(err)
	return string(b)
}
