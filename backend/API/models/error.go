package models


// ErrorResponse represents a standardized error message returned by the API.
// swagger:model ErrorResponse
type ErrorResponse struct {
	// The HTTP status code of the response.
	// example: 400
	Code int `json:"code"`
	// The standard HTTP status text for the code.
	// example: Bad Request
	Error string `json:"error"`
	// A user-friendly message describing the specific issue.
	// example: Invalid request body or validation failed.
	Message string `json:"message"`
}