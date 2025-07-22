package utils

import (
	"encoding/json"
	"net/http"
)

func JSONResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func ErrorResponse(w http.ResponseWriter, message string, status int) {
	response := struct {
		Code	int    `json:"code"`   // Use the HTTP status code
		Error   string `json:"error"`
		Message string `json:"message"` // Often good to include a user-friendly message
	}{
		Code:   status, // Use the HTTP status code
		Error:   http.StatusText(status), // Get standard HTTP status text
		Message: message,
	}
	JSONResponse(w, response, status)
}