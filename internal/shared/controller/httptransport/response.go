package httptransport

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Status  int
	Code    string
	Message string
}

func WriteJSON(responseWriter http.ResponseWriter, status int, body any) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(status)
	_ = json.NewEncoder(responseWriter).Encode(body)
}

func WriteError(responseWriter http.ResponseWriter, response ErrorResponse) {
	WriteJSON(responseWriter, response.Status, errorResponse{Error: errorBody{Code: response.Code, Message: response.Message}})
}
