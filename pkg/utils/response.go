package utils

import (
	"encoding/json"
	"net/http"
)

type ErrorResponseBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func JSONResponse(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return
	}
}

func ErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(ErrorResponseBody{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
	if err != nil {
		return
	}
}
