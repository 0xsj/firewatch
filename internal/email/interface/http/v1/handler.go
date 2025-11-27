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

// CreateTemplate godoc
// @Summary      Create email template
// @Description  Creates a new email template with subject, body, and variable definitions
// @Tags         email
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateTemplateRequest true "Create template request"
// @Success      201 {object} dto.CreateTemplateResponse "Template created successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      409 {object} ErrorResponse "Template with slug already exists"
// @Router       /api/v1/email/templates [post]
// @Security     BearerAuth
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

// UpdateTemplate godoc
// @Summary      Update email template
// @Description  Updates an existing email template
// @Tags         email
// @Accept       json
// @Produce      json
// @Param        id path string true "Template ID" format(uuid)
// @Param        request body dto.UpdateTemplateRequest true "Update template request"
// @Success      200 {object} dto.UpdateTemplateResponse "Template updated successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body or template ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Template not found"
// @Router       /api/v1/email/templates/{id} [put]
// @Security     BearerAuth
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

// ActivateTemplate godoc
// @Summary      Activate email template
// @Description  Activates a draft or archived template, making it available for use
// @Tags         email
// @Produce      json
// @Param        id path string true "Template ID" format(uuid)
// @Success      200 {object} dto.ActivateTemplateResponse "Template activated successfully"
// @Failure      400 {object} ErrorResponse "Invalid template ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Template not found"
// @Failure      422 {object} ErrorResponse "Invalid status transition"
// @Router       /api/v1/email/templates/{id}/activate [post]
// @Security     BearerAuth
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

// ArchiveTemplate godoc
// @Summary      Archive email template
// @Description  Archives an active template, removing it from active use
// @Tags         email
// @Produce      json
// @Param        id path string true "Template ID" format(uuid)
// @Success      200 {object} dto.ArchiveTemplateResponse "Template archived successfully"
// @Failure      400 {object} ErrorResponse "Invalid template ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Template not found"
// @Failure      422 {object} ErrorResponse "Invalid status transition"
// @Router       /api/v1/email/templates/{id}/archive [post]
// @Security     BearerAuth
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

// DeleteTemplate godoc
// @Summary      Delete email template
// @Description  Soft-deletes an email template
// @Tags         email
// @Produce      json
// @Param        id path string true "Template ID" format(uuid)
// @Success      200 {object} dto.DeleteTemplateResponse "Template deleted successfully"
// @Failure      400 {object} ErrorResponse "Invalid template ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Template not found"
// @Router       /api/v1/email/templates/{id} [delete]
// @Security     BearerAuth
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

// GetTemplate godoc
// @Summary      Get email template by ID
// @Description  Retrieves an email template by its unique identifier
// @Tags         email
// @Produce      json
// @Param        id path string true "Template ID" format(uuid)
// @Success      200 {object} dto.GetTemplateResponse "Template found"
// @Failure      400 {object} ErrorResponse "Invalid template ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Template not found"
// @Router       /api/v1/email/templates/{id} [get]
// @Security     BearerAuth
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

// GetTemplateBySlug godoc
// @Summary      Get email template by slug
// @Description  Retrieves an email template by its slug and locale
// @Tags         email
// @Produce      json
// @Param        slug query string true "Template slug" example(welcome-email)
// @Param        locale query string false "Locale code" default(en) example(en)
// @Param        tenant_id query string false "Tenant ID" format(uuid)
// @Success      200 {object} dto.GetTemplateResponse "Template found"
// @Failure      400 {object} ErrorResponse "Slug is required"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Template not found"
// @Router       /api/v1/email/templates/by-slug [get]
// @Security     BearerAuth
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

// ListTemplates godoc
// @Summary      List email templates
// @Description  Retrieves a paginated list of email templates with optional filters
// @Tags         email
// @Produce      json
// @Param        tenant_id query string false "Filter by tenant ID" format(uuid)
// @Param        status query string false "Filter by status" Enums(draft, active, archived, deleted)
// @Param        locale query string false "Filter by locale" example(en)
// @Param        slug_contains query string false "Filter by slug containing text"
// @Param        name_contains query string false "Filter by name containing text"
// @Param        include_system query bool false "Include system templates" default(true)
// @Param        offset query int false "Pagination offset" default(0) minimum(0)
// @Param        limit query int false "Pagination limit" default(20) minimum(1) maximum(100)
// @Param        sort_by query string false "Sort field" Enums(created_at, updated_at, name, slug)
// @Param        sort_order query string false "Sort order" Enums(asc, desc) default(desc)
// @Success      200 {object} dto.ListTemplatesResponse "List of templates"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/v1/email/templates [get]
// @Security     BearerAuth
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

// PreviewTemplate godoc
// @Summary      Preview email template
// @Description  Renders an email template with provided data for preview
// @Tags         email
// @Accept       json
// @Produce      json
// @Param        id path string true "Template ID" format(uuid)
// @Param        request body map[string]interface{} true "Template variables data"
// @Success      200 {object} dto.PreviewTemplateResponse "Rendered template preview"
// @Failure      400 {object} ErrorResponse "Invalid request body or template ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Template not found"
// @Router       /api/v1/email/templates/{id}/preview [post]
// @Security     BearerAuth
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

// PreviewTemplateBySlug godoc
// @Summary      Preview email template by slug
// @Description  Renders an email template by slug with provided data for preview
// @Tags         email
// @Accept       json
// @Produce      json
// @Param        request body dto.PreviewTemplateBySlugRequest true "Preview request with slug and data"
// @Success      200 {object} dto.PreviewTemplateResponse "Rendered template preview"
// @Failure      400 {object} ErrorResponse "Invalid request body or missing slug"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Template not found"
// @Router       /api/v1/email/templates/preview-by-slug [post]
// @Security     BearerAuth
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

// ============================================================================
// Swagger Models
// ============================================================================

// ErrorResponse represents an error response body.
// @Description Error response returned by the API
type ErrorResponse struct {
	Code    string         `json:"code" example:"TEMPLATE_NOT_FOUND"`
	Message string         `json:"message" example:"template not found"`
	Meta    map[string]any `json:"meta,omitempty"`
}
