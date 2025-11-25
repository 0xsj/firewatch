package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/email/domain"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
)

// HandleError maps domain errors to HTTP responses.
func HandleError(w http.ResponseWriter, err error) {
	// Check for specific error codes
	code := pkgerrors.CodeOf(err)

	switch code {
	case domain.ErrCodeTemplateNotFound:
		response.NotFound(w, err.Error())
		return

	case domain.ErrCodeTemplateAlreadyExists:
		response.Conflict(w, err.Error())
		return

	case domain.ErrCodeTemplateCannotActivate,
		domain.ErrCodeTemplateCannotArchive,
		domain.ErrCodeTemplateCannotEdit,
		domain.ErrCodeTemplateCannotDelete,
		domain.ErrCodeTemplateMissingVar,
		domain.ErrCodeTemplateInvalidVar:
		response.BadRequest(w, err.Error())
		return

	case domain.ErrCodeTemplateRenderFailed:
		response.InternalServerError(w, err.Error())
		return
	}

	// Fall back to generic error handling
	response.Error(w, err)
}
