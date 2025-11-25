package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/email/application/dto"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
)

// RespondWithTemplate writes a single template response.
func RespondWithTemplate(w http.ResponseWriter, status int, resp *dto.GetTemplateResponse) {
	response.JSON(w, status, resp)
}

// RespondWithTemplateCreated writes a created template response.
func RespondWithTemplateCreated(w http.ResponseWriter, resp *dto.CreateTemplateResponse) {
	response.JSON(w, http.StatusCreated, resp)
}

// RespondWithTemplateUpdated writes an updated template response.
func RespondWithTemplateUpdated(w http.ResponseWriter, resp *dto.UpdateTemplateResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithTemplateActivated writes an activated template response.
func RespondWithTemplateActivated(w http.ResponseWriter, resp *dto.ActivateTemplateResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithTemplateArchived writes an archived template response.
func RespondWithTemplateArchived(w http.ResponseWriter, resp *dto.ArchiveTemplateResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithTemplateDeleted writes a deleted template response.
func RespondWithTemplateDeleted(w http.ResponseWriter, resp *dto.DeleteTemplateResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithTemplateList writes a list of templates response.
func RespondWithTemplateList(w http.ResponseWriter, resp *dto.ListTemplatesResponse) {
	response.JSON(w, http.StatusOK, resp)
}

// RespondWithTemplatePreview writes a template preview response.
func RespondWithTemplatePreview(w http.ResponseWriter, resp *dto.PreviewTemplateResponse) {
	response.JSON(w, http.StatusOK, resp)
}
