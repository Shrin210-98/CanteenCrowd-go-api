package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// JsonResponse sends a JSON response with the given status code and data
func JsonResponse(w http.ResponseWriter, statusCode int, data map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// ErrorResponse sends an error response with the given status code and message
func ErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	JsonResponse(w, statusCode, map[string]any{
		"message": message,
	})
}

// SuccessResponse sends a success response with data and optional message
// If message is empty, it will be omitted from the response
func SuccessResponse(w http.ResponseWriter, statusCode int, data any, message string) {
	response := make(map[string]any)

	// Only include data if it's not nil
	if data != nil {
		response["data"] = data
	}

	// Only include message if it's not empty
	if message != "" {
		response["message"] = message
	}

	JsonResponse(w, statusCode, response)
}

// SuccessResponseWithMeta sends a success response with data, message, and metadata
// Useful for pagination or additional information
func SuccessResponseWithMeta(w http.ResponseWriter, statusCode int, data any, message string, meta map[string]any) {
	response := make(map[string]any)

	if data != nil {
		response["data"] = data
	}

	if message != "" {
		response["message"] = message
	}

	if meta != nil {
		response["meta"] = meta
	}

	JsonResponse(w, statusCode, response)
}

// ValidationErrorResponse sends a validation error response with field-specific errors
func ValidationErrorResponse(w http.ResponseWriter, errors map[string]string) {
	JsonResponse(w, http.StatusBadRequest, map[string]any{
		"message": "Validation failed",
		"errors":  errors,
	})
}

// PaginatedResponse sends a paginated response with data and pagination metadata
func PaginatedResponse(w http.ResponseWriter, statusCode int, data any, message string, page, pageSize, total int) {
	response := map[string]any{
		"data": data,
		"meta": map[string]any{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": (total + pageSize - 1) / pageSize, // Ceiling division
		},
	}

	if message != "" {
		response["message"] = message
	}

	JsonResponse(w, statusCode, response)
}
