package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
)

// HandleError writes an error response with the appropriate HTTP status code.
// It handles both domain-specific errors and generic errors.
//
// Domain errors are mapped to specific HTTP status codes based on error codes.
// Generic errors use the standard error kind → status code mapping.
func HandleError(w http.ResponseWriter, err error) {
	// Check for specific domain error codes first
	code := errors.CodeOf(err)

	switch code {
	// Email errors
	case user.ErrCodeEmailAlreadyTaken:
		response.ErrorWithStatus(w, http.StatusConflict, err)
		return

	case user.ErrCodeEmailNotVerified:
		response.ErrorWithStatus(w, http.StatusForbidden, err)
		return

	case user.ErrCodeEmailInvalid:
		response.ErrorWithStatus(w, http.StatusBadRequest, err)
		return

	// Password errors
	case user.ErrCodePasswordTooWeak,
		user.ErrCodePasswordInvalid:
		response.ErrorWithStatus(w, http.StatusBadRequest, err)
		return

	// Account status errors
	case user.ErrCodeAccountLocked:
		response.ErrorWithStatus(w, http.StatusForbidden, err)
		return

	case user.ErrCodeAccountSuspended:
		response.ErrorWithStatus(w, http.StatusForbidden, err)
		return

	case user.ErrCodeAccountDeleted:
		response.ErrorWithStatus(w, http.StatusGone, err)
		return

	// Authentication errors
	case user.ErrCodeInvalidCredentials:
		response.ErrorWithStatus(w, http.StatusUnauthorized, err)
		return

	case user.ErrCodeTooManyAttempts:
		response.ErrorWithStatus(w, http.StatusTooManyRequests, err)
		return

	// User not found
	case user.ErrCodeUserNotFound:
		response.ErrorWithStatus(w, http.StatusNotFound, err)
		return
	}

	// Fall back to generic error kind mapping
	response.Error(w, err)
}
