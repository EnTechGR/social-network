package utils

import (
	"encoding/json"
	"net/http"
)

// ============================================================================
// HTTP RESPONSE UTILITIES
// ============================================================================

// ErrorResponse represents an error response
type ErrorResp struct {
	Error string `json:"error"`
}

// JSONResponse sends a JSON response with the given status code
func JSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, send a plain text error
		ErrorResponse(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// ErrorResponse sends an error response in JSON format
func ErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResp{
		Error: message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If encoding fails, send a plain text error
		ErrorResponse(w, message, statusCode)
	}
}

// SuccessResponse sends a success message in JSON format
func SuccessResponse(w http.ResponseWriter, message string) {
	JSONResponse(w, map[string]string{
		"message": message,
	}, http.StatusOK)
}
