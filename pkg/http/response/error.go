package response

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/pkg/errors"
)

// ErrorResponse represents a standard error response structure.
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error writes an error response with appropriate HTTP status code.
// Automatically maps domain error kinds to HTTP status codes.
//
// Example:
//
//	if err != nil {
//	    response.Error(w, err)
//	    return
//	}
func Error(w http.ResponseWriter, err error) {
	statusCode := MapErrorToStatusCode(err)
	ErrorWithStatus(w, statusCode, err)
}

// ErrorWithStatus writes an error response with explicit status code.
//
// Example:
//
//	response.ErrorWithStatus(w, http.StatusBadRequest, err)
func ErrorWithStatus(w http.ResponseWriter, statusCode int, err error) {
	// Build error response
	errorResp := ErrorResponse{
		Error: err.Error(),
	}

	// Try to extract code and metadata from our error type
	if code := errors.CodeOf(err); code != "" {
		errorResp.Code = string(code)
	}

	if metadata := errors.MetaOf(err); len(metadata) > 0 {
		errorResp.Details = metadata
	}

	JSON(w, statusCode, errorResp)
}

// BadRequest writes a 400 Bad Request error response.
//
// Example:
//
//	response.BadRequest(w, "invalid request body")
func BadRequest(w http.ResponseWriter, message string) {
	err := errors.Validation("http", message)
	ErrorWithStatus(w, http.StatusBadRequest, err)
}

// Unauthorized writes a 401 Unauthorized error response.
//
// Example:
//
//	response.Unauthorized(w, "invalid credentials")
func Unauthorized(w http.ResponseWriter, message string) {
	err := errors.Unauthorized("http", message)
	ErrorWithStatus(w, http.StatusUnauthorized, err)
}

// Forbidden writes a 403 Forbidden error response.
//
// Example:
//
//	response.Forbidden(w, "insufficient permissions")
func Forbidden(w http.ResponseWriter, message string) {
	err := errors.Forbidden("http", message)
	ErrorWithStatus(w, http.StatusForbidden, err)
}

// NotFound writes a 404 Not Found error response.
//
// Example:
//
//	response.NotFound(w, "user not found")
func NotFound(w http.ResponseWriter, message string) {
	err := errors.NotFound("http", message)
	ErrorWithStatus(w, http.StatusNotFound, err)
}

// Conflict writes a 409 Conflict error response.
//
// Example:
//
//	response.Conflict(w, "email already exists")
func Conflict(w http.ResponseWriter, message string) {
	err := errors.Conflict("http", message)
	ErrorWithStatus(w, http.StatusConflict, err)
}

// InternalServerError writes a 500 Internal Server Error response.
//
// Example:
//
//	response.InternalServerError(w, "unexpected error occurred")
func InternalServerError(w http.ResponseWriter, message string) {
	err := errors.Internal("http", nil)
	if message != "" {
		err = err.WithMessage(message)
	}
	ErrorWithStatus(w, http.StatusInternalServerError, err)
}

// ValidationError writes a 400 Bad Request with field-level validation errors.
//
// Example:
//
//	response.ValidationError(w, map[string]string{
//	    "email": "must be a valid email address",
//	    "password": "must be at least 8 characters",
//	})
func ValidationError(w http.ResponseWriter, fields map[string]string) {
	JSON(w, http.StatusBadRequest, map[string]interface{}{
		"error":  "Validation failed",
		"code":   "VALIDATION_ERROR",
		"fields": fields,
	})
}

// MapErrorToStatusCode maps domain error kinds to HTTP status codes.
func MapErrorToStatusCode(err error) int {
	// Get error kind using helper function
	kind := errors.KindOf(err)

	// Map error kind to HTTP status code
	switch kind {
	case errors.KindValidation:
		return http.StatusBadRequest // 400

	case errors.KindUnauthorized:
		return http.StatusUnauthorized // 401

	case errors.KindForbidden:
		return http.StatusForbidden // 403

	case errors.KindNotFound:
		return http.StatusNotFound // 404

	case errors.KindConflict:
		return http.StatusConflict // 409

	case errors.KindRateLimit:
		return http.StatusTooManyRequests // 429

	case errors.KindInternal:
		return http.StatusInternalServerError // 500

	case errors.KindInfrastructure:
		return http.StatusServiceUnavailable // 503

	case errors.KindTimeout:
		return http.StatusGatewayTimeout // 504

	default:
		return http.StatusInternalServerError // 500
	}
}
