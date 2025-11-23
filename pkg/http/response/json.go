package response

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON response with the given status code and data.
// Automatically sets Content-Type header to application/json.
//
// Example:
//
//	response.JSON(w, http.StatusOK, user)
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// If encoding fails, we've already written headers, so log the error
			// In production, you'd log this with your logger
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

// Created writes a 201 Created response with the given data.
//
// Example:
//
//	response.Created(w, newUser)
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// OK writes a 200 OK response with the given data.
//
// Example:
//
//	response.OK(w, users)
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// NoContent writes a 204 No Content response (no body).
// Used for successful operations that don't return data (e.g., DELETE).
//
// Example:
//
//	response.NoContent(w)
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Accepted writes a 202 Accepted response.
// Used for async operations that were accepted but not yet completed.
//
// Example:
//
//	response.Accepted(w, map[string]string{"job_id": jobID})
func Accepted(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusAccepted, data)
}
