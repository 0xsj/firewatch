package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/flags/application/dto"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
)

// RespondWithFlag writes a single flag response.
func RespondWithFlag(w http.ResponseWriter, status int, resp *dto.GetFlagResponse) {
	response.JSON(w, status, resp)
}

// RespondWithFlagCreated writes a created flag response.
func RespondWithFlagCreated(w http.ResponseWriter, resp *dto.CreateFlagResponse) {
	response.JSON(w, http.StatusCreated, resp)
}

// RespondWithFlagUpdated writes an updated flag response.
func RespondWithFlagUpdated(w http.ResponseWriter, resp *dto.UpdateFlagResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithFlagDeleted writes a deleted flag response.
func RespondWithFlagDeleted(w http.ResponseWriter, resp *dto.DeleteFlagResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithFlagEnabled writes an enabled flag response.
func RespondWithFlagEnabled(w http.ResponseWriter, resp *dto.EnableFlagResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithFlagDisabled writes a disabled flag response.
func RespondWithFlagDisabled(w http.ResponseWriter, resp *dto.DisableFlagResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithFlagList writes a list of flags response.
func RespondWithFlagList(w http.ResponseWriter, resp *dto.ListFlagsResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithVariantAdded writes a variant added response.
func RespondWithVariantAdded(w http.ResponseWriter, resp *dto.AddVariantResponse) {
	response.JSON(w, http.StatusCreated, resp)
}

// RespondWithVariantRemoved writes a variant removed response.
func RespondWithVariantRemoved(w http.ResponseWriter, resp *dto.RemoveVariantResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithRuleAdded writes a rule added response.
func RespondWithRuleAdded(w http.ResponseWriter, resp *dto.AddRuleResponse) {
	response.JSON(w, http.StatusCreated, resp)
}

// RespondWithRuleRemoved writes a rule removed response.
func RespondWithRuleRemoved(w http.ResponseWriter, resp *dto.RemoveRuleResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithOverrideSet writes an override set response.
func RespondWithOverrideSet(w http.ResponseWriter, resp *dto.SetOverrideResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithOverrideRemoved writes an override removed response.
func RespondWithOverrideRemoved(w http.ResponseWriter, resp *dto.RemoveOverrideResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithEvaluation writes an evaluation response.
func RespondWithEvaluation(w http.ResponseWriter, resp *dto.EvaluateFlagResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithBulkEvaluation writes a bulk evaluation response.
func RespondWithBulkEvaluation(w http.ResponseWriter, resp *dto.EvaluateFlagsResponse) {
	response.JSON(w, http.StatusOK, resp)
}
