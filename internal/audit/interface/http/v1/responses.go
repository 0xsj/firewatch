// internal/audit/interface/http/v1/responses.go
package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/audit/application/dto"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
)

// RespondWithEntry writes an audit entry DTO as JSON response.
func RespondWithEntry(w http.ResponseWriter, resp *dto.GetEntryResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithEntryList writes a paginated list of audit entries.
func RespondWithEntryList(w http.ResponseWriter, resp *dto.ListEntriesResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithResourceTrail writes a resource audit trail response.
func RespondWithResourceTrail(w http.ResponseWriter, resp *dto.GetResourceTrailResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithActorActivity writes an actor activity response.
func RespondWithActorActivity(w http.ResponseWriter, resp *dto.GetActorActivityResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithMessage writes a simple success message.
func RespondWithMessage(w http.ResponseWriter, message string) {
	response.JSON(w, http.StatusOK, map[string]string{
		"message": message,
	})
}

// RespondValidationError writes a validation error response.
func RespondValidationError(w http.ResponseWriter, err error) {
	response.ValidationError(w, map[string]string{
		"error": err.Error(),
	})
}
