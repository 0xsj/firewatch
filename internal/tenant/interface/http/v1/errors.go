package v1

import (
	"net/http"

	tenant "github.com/0xsj/hexagonal-go/internal/tenant/domain"
	"github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
)

// HandleError writes an error response with the appropriate HTTP status code.
// It handles both domain-specific errors and generic errors.
func HandleError(w http.ResponseWriter, err error) {
	// Check for specific domain error codes first
	code := errors.CodeOf(err)

	switch code {
	// Tenant not found
	case tenant.ErrCodeTenantNotFound:
		response.ErrorWithStatus(w, http.StatusNotFound, err)
		return

	// Slug errors
	case tenant.ErrCodeSlugAlreadyTaken:
		response.ErrorWithStatus(w, http.StatusConflict, err)
		return
	case tenant.ErrCodeSlugInvalid,
		tenant.ErrCodeSlugReserved:
		response.ErrorWithStatus(w, http.StatusBadRequest, err)
		return

	// Status errors
	case tenant.ErrCodeTenantSuspended:
		response.ErrorWithStatus(w, http.StatusForbidden, err)
		return
	case tenant.ErrCodeTenantDeleted:
		response.ErrorWithStatus(w, http.StatusGone, err)
		return
	case tenant.ErrCodeInvalidStatusChange:
		response.ErrorWithStatus(w, http.StatusUnprocessableEntity, err)
		return

	// Plan errors
	case tenant.ErrCodePlanInvalid:
		response.ErrorWithStatus(w, http.StatusBadRequest, err)
		return
	case tenant.ErrCodePlanDowngrade:
		response.ErrorWithStatus(w, http.StatusUnprocessableEntity, err)
		return

	// Settings errors
	case tenant.ErrCodeSettingsInvalid:
		response.ErrorWithStatus(w, http.StatusBadRequest, err)
		return

	// Name errors
	case tenant.ErrCodeTenantNameInvalid:
		response.ErrorWithStatus(w, http.StatusBadRequest, err)
		return
	}

	// Fall back to generic error kind mapping
	response.Error(w, err)
}

// RespondValidationError writes a validation error response.
func RespondValidationError(w http.ResponseWriter, err error) {
	response.ValidationError(w, map[string]string{
		"error": err.Error(),
	})
}
