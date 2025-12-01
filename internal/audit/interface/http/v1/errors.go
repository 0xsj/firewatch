// internal/audit/interface/http/v1/errors.go
package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
)

// HandleError writes an error response with the appropriate HTTP status code.
// The audit domain is primarily read-only, so errors are mostly:
// - Not found (entry doesn't exist)
// - Validation errors (invalid filters)
// - Internal errors (database issues)
func HandleError(w http.ResponseWriter, err error) {
	// Check error kind for generic mapping
	if errors.IsNotFound(err) {
		response.ErrorWithStatus(w, http.StatusNotFound, err)
		return
	}

	if errors.IsValidation(err) {
		response.ErrorWithStatus(w, http.StatusBadRequest, err)
		return
	}

	if errors.IsUnauthorized(err) {
		response.ErrorWithStatus(w, http.StatusUnauthorized, err)
		return
	}

	if errors.IsForbidden(err) {
		response.ErrorWithStatus(w, http.StatusForbidden, err)
		return
	}

	// Fall back to generic error handling
	response.Error(w, err)
}
