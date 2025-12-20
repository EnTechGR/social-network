package utils

import (
	"encoding/json"
	"forum/models"
	"net/http"
	"strings"
)

func JSONResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func ErrorResponse(w http.ResponseWriter, message string, status int) {
	// Use the named model type
	response := models.ErrorResponse{
		Code:    status,                  // Use the HTTP status code
		Error:   http.StatusText(status), // Get standard HTTP status text
		Message: message,
	}
	// Assuming JSONResponse is in utils and accepts models.ErrorResponse
	JSONResponse(w, response, status)
}

// GetLastPathParam extracts the last segment of the URL path (e.g., for /posts/123 it returns "123")
func GetLastPathParam(r *http.Request) string {
	path := r.URL.Path
	if path == "" {
		return ""
	}
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) == 0 {
		return ""
	}
	return segments[len(segments)-1]
}

// DerefString returns the value of a *string or an empty string if nil
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
