package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/hexagonal-go/internal/email/application/command"
	"github.com/0xsj/hexagonal-go/internal/email/application/dto"
	"github.com/0xsj/hexagonal-go/internal/email/application/query"
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Handler handles HTTP requests for email templates.
type Handler struct {
	createCmd    *command.CreateTemplateCommand
	updateCmd    *command.UpdateTemplateCommand
	activateCmd  *command.ActivateTemplateCommand
	archiveCmd   *command.ArchiveTemplateCommand
	deleteCmd    *command.DeleteTemplateCommand
	getQuery     *query.GetTemplateQuery
	listQuery    *query.ListTemplatesQuery
	previewQuery *query.PreviewTemplateQuery
	logger       logger.Logger
}

// NewHandler creates a new email template handler.
func NewHandler(
	createCmd *command.CreateTemplateCommand,
	updateCmd *command.UpdateTemplateCommand,
	activateCmd *command.ActivateTemplateCommand,
	archiveCmd *command.ArchiveTemplateCommand,
	deleteCmd *command.DeleteTemplateCommand,
	getQuery *query.GetTemplateQuery,
	listQuery *query.ListTemplatesQuery,
	previewQuery *query.PreviewTemplateQuery,
	logger logger.Logger,
) *Handler {
	return &Handler{
		createCmd:    createCmd,
		updateCmd:    updateCmd,
		activateCmd:  activateCmd,
		archiveCmd:   archiveCmd,
		deleteCmd:    deleteCmd,
		getQuery:     getQuery,
		listQuery:    listQuery,
		previewQuery: previewQuery,
		logger:       logger,
	}
}

// CreateTemplate handles POST /api/v1/email/templates
func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	userID := getUserIDFromContext(r)

	resp, err := h.createCmd.Handle(r.Context(), req, userID)
	if err != nil {
		h.logger.Error("failed to create template", logger.Err(err))
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, resp)
}

// UpdateTemplate handles PUT /api/v1/email/templates/{id}
func (h *Handler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	templateID, err := parseTemplateID(r)
	if err != nil {
		response.BadRequest(w, "invalid template ID")
		return
	}

	var req dto.UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	userID := getUserIDFromContext(r)

	resp, err := h.updateCmd.Handle(r.Context(), templateID, req, userID)
	if err != nil {
		h.logger.Error("failed to update template", logger.Err(err))
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// ActivateTemplate handles POST /api/v1/email/templates/{id}/activate
func (h *Handler) ActivateTemplate(w http.ResponseWriter, r *http.Request) {
	templateID, err := parseTemplateID(r)
	if err != nil {
		response.BadRequest(w, "invalid template ID")
		return
	}

	userID := getUserIDFromContext(r)

	resp, err := h.activateCmd.Handle(r.Context(), templateID, userID)
	if err != nil {
		h.logger.Error("failed to activate template", logger.Err(err))
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// ArchiveTemplate handles POST /api/v1/email/templates/{id}/archive
func (h *Handler) ArchiveTemplate(w http.ResponseWriter, r *http.Request) {
	templateID, err := parseTemplateID(r)
	if err != nil {
		response.BadRequest(w, "invalid template ID")
		return
	}

	userID := getUserIDFromContext(r)

	resp, err := h.archiveCmd.Handle(r.Context(), templateID, userID)
	if err != nil {
		h.logger.Error("failed to archive template", logger.Err(err))
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// DeleteTemplate handles DELETE /api/v1/email/templates/{id}
func (h *Handler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	templateID, err := parseTemplateID(r)
	if err != nil {
		response.BadRequest(w, "invalid template ID")
		return
	}

	userID := getUserIDFromContext(r)

	resp, err := h.deleteCmd.Handle(r.Context(), templateID, userID)
	if err != nil {
		h.logger.Error("failed to delete template", logger.Err(err))
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// GetTemplate handles GET /api/v1/email/templates/{id}
func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	templateID, err := parseTemplateID(r)
	if err != nil {
		response.BadRequest(w, "invalid template ID")
		return
	}

	resp, err := h.getQuery.Handle(r.Context(), templateID)
	if err != nil {
		h.logger.Error("failed to get template", logger.Err(err))
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// GetTemplateBySlug handles GET /api/v1/email/templates/by-slug
func (h *Handler) GetTemplateBySlug(w http.ResponseWriter, r *http.Request) {
	req := dto.GetTemplateBySlugRequest{
		Slug:   r.URL.Query().Get("slug"),
		Locale: r.URL.Query().Get("locale"),
	}

	if tenantID := r.URL.Query().Get("tenant_id"); tenantID != "" {
		req.TenantID = &tenantID
	}

	if req.Slug == "" {
		response.BadRequest(w, "slug is required")
		return
	}
	if req.Locale == "" {
		req.Locale = "en"
	}

	resp, err := h.getQuery.HandleBySlug(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to get template by slug", logger.Err(err))
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// ListTemplates handles GET /api/v1/email/templates
func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	req := parseListTemplatesRequest(r)

	resp, err := h.listQuery.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to list templates", logger.Err(err))
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// PreviewTemplate handles POST /api/v1/email/templates/{id}/preview
func (h *Handler) PreviewTemplate(w http.ResponseWriter, r *http.Request) {
	templateID, err := parseTemplateID(r)
	if err != nil {
		response.BadRequest(w, "invalid template ID")
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	req := dto.PreviewTemplateRequest{
		ID:   templateID.String(),
		Data: data,
	}

	resp, err := h.previewQuery.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to preview template", logger.Err(err))
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// PreviewTemplateBySlug handles POST /api/v1/email/templates/preview-by-slug
func (h *Handler) PreviewTemplateBySlug(w http.ResponseWriter, r *http.Request) {
	var req dto.PreviewTemplateBySlugRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Slug == "" {
		response.BadRequest(w, "slug is required")
		return
	}
	if req.Locale == "" {
		req.Locale = "en"
	}

	resp, err := h.previewQuery.HandleBySlug(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to preview template by slug", logger.Err(err))
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// ============================================================================
// Helpers
// ============================================================================

func parseTemplateID(r *http.Request) (types.ID, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return types.ID{}, fmt.Errorf("template ID is required")
	}
	return types.ParseID(idStr)
}

func getUserIDFromContext(r *http.Request) *types.ID {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		return nil
	}
	userID, err := types.ParseID(userIDStr)
	if err != nil {
		return nil
	}
	return &userID
}

func parseListTemplatesRequest(r *http.Request) dto.ListTemplatesRequest {
	req := dto.ListTemplatesRequest{
		IncludeSystemTemplates: true,
		Limit:                  20,
		Offset:                 0,
	}

	if tenantID := r.URL.Query().Get("tenant_id"); tenantID != "" {
		req.TenantID = &tenantID
	}

	if status := r.URL.Query().Get("status"); status != "" {
		req.Status = &status
	}

	if locale := r.URL.Query().Get("locale"); locale != "" {
		req.Locale = &locale
	}

	if slugContains := r.URL.Query().Get("slug_contains"); slugContains != "" {
		req.SlugContains = slugContains
	}

	if nameContains := r.URL.Query().Get("name_contains"); nameContains != "" {
		req.NameContains = nameContains
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := parseInt(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := parseInt(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		req.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		req.SortOrder = sortOrder
	}

	if includeSystem := r.URL.Query().Get("include_system"); includeSystem == "false" {
		req.IncludeSystemTemplates = false
	}

	return req
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
