package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/flags/domain"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
)

// HandleError maps domain errors to HTTP responses.
func HandleError(w http.ResponseWriter, err error) {
	// Check for specific error codes
	code := pkgerrors.CodeOf(err)

	switch code {
	case domain.ErrCodeFlagNotFound:
		response.NotFound(w, err.Error())
		return
	case domain.ErrCodeFlagAlreadyExists:
		response.Conflict(w, err.Error())
		return
	case domain.ErrCodeFlagDisabled,
		domain.ErrCodeFlagKeyInvalid,
		domain.ErrCodeVariantKeyInvalid,
		domain.ErrCodeRuleInvalid:
		response.BadRequest(w, err.Error())
		return
	case domain.ErrCodeVariantNotFound,
		domain.ErrCodeRuleNotFound,
		domain.ErrCodeOverrideNotFound:
		response.NotFound(w, err.Error())
		return
	case domain.ErrCodeEvaluationFailed:
		response.InternalServerError(w, err.Error())
		return
	}

	// Fall back to generic error handling
	response.Error(w, err)
}
