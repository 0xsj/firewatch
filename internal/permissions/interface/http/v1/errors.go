package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/permissions/domain"
	"github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
)

// HandleError writes an error response with the appropriate HTTP status code.
// It handles both domain-specific errors and generic errors.
func HandleError(w http.ResponseWriter, err error) {
	code := errors.CodeOf(err)

	switch code {
	// Role errors
	case domain.ErrCodeRoleNotFound:
		response.ErrorWithStatus(w, http.StatusNotFound, err)
		return
	case domain.ErrCodeRoleAlreadyExists:
		response.ErrorWithStatus(w, http.StatusConflict, err)
		return
	case domain.ErrCodeRoleIsSystem:
		response.ErrorWithStatus(w, http.StatusForbidden, err)
		return
	case domain.ErrCodeRoleNameInvalid:
		response.ErrorWithStatus(w, http.StatusBadRequest, err)
		return

	// Assignment errors
	case domain.ErrCodeAssignmentNotFound:
		response.ErrorWithStatus(w, http.StatusNotFound, err)
		return
	case domain.ErrCodeAssignmentExists:
		response.ErrorWithStatus(w, http.StatusConflict, err)
		return
	case domain.ErrCodeAssignmentExpired:
		response.ErrorWithStatus(w, http.StatusForbidden, err)
		return

	// Permission errors
	case domain.ErrCodePermissionInvalid:
		response.ErrorWithStatus(w, http.StatusBadRequest, err)
		return
	case domain.ErrCodePermissionDenied:
		response.ErrorWithStatus(w, http.StatusForbidden, err)
		return

	// Action/Resource errors
	case domain.ErrCodeActionInvalid:
		response.ErrorWithStatus(w, http.StatusBadRequest, err)
		return
	case domain.ErrCodeResourceInvalid:
		response.ErrorWithStatus(w, http.StatusBadRequest, err)
		return
	}

	// Fall back to generic error kind mapping
	response.Error(w, err)
}
