package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/tenant/application/command"
	"github.com/0xsj/hexagonal-go/internal/tenant/application/query"
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// Handler handles HTTP requests for the Tenant domain (v1 API).
type Handler struct {
	createTenantCmd      *command.CreateTenantCommand
	updateTenantCmd      *command.UpdateTenantCommand
	updateSettingsCmd    *command.UpdateSettingsCommand
	suspendTenantCmd     *command.SuspendTenantCommand
	reactivateTenantCmd  *command.ReactivateTenantCommand
	changePlanCmd        *command.ChangePlanCommand
	deleteTenantCmd      *command.DeleteTenantCommand
	getTenantQuery       *query.GetTenantQuery
	getTenantBySlugQuery *query.GetTenantBySlugQuery
	listTenantsQuery     *query.ListTenantsQuery
	logger               logger.Logger
}

// NewHandler creates a new v1 tenant HTTP handler.
func NewHandler(
	createTenantCmd *command.CreateTenantCommand,
	updateTenantCmd *command.UpdateTenantCommand,
	updateSettingsCmd *command.UpdateSettingsCommand,
	suspendTenantCmd *command.SuspendTenantCommand,
	reactivateTenantCmd *command.ReactivateTenantCommand,
	changePlanCmd *command.ChangePlanCommand,
	deleteTenantCmd *command.DeleteTenantCommand,
	getTenantQuery *query.GetTenantQuery,
	getTenantBySlugQuery *query.GetTenantBySlugQuery,
	listTenantsQuery *query.ListTenantsQuery,
	log logger.Logger,
) *Handler {
	return &Handler{
		createTenantCmd:      createTenantCmd,
		updateTenantCmd:      updateTenantCmd,
		updateSettingsCmd:    updateSettingsCmd,
		suspendTenantCmd:     suspendTenantCmd,
		reactivateTenantCmd:  reactivateTenantCmd,
		changePlanCmd:        changePlanCmd,
		deleteTenantCmd:      deleteTenantCmd,
		getTenantQuery:       getTenantQuery,
		getTenantBySlugQuery: getTenantBySlugQuery,
		listTenantsQuery:     listTenantsQuery,
		logger:               log,
	}
}

// CreateTenant godoc
// @Summary      Create a new tenant
// @Description  Creates a new tenant with the specified details. Requires admin privileges.
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateTenantRequest true "Create tenant request"
// @Success      201 {object} dto.TenantDTO "Tenant created successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body or validation error"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure      409 {object} ErrorResponse "Slug already taken"
// @Router       /api/v1/tenants [post]
// @Security     BearerAuth
func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	dtoReq, err := ParseCreateTenantRequest(r)
	if err != nil {
		h.logger.Warn("invalid create tenant request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	createdBy := getActorFromContext(r)

	cmdReq := command.CreateTenantRequest{
		Slug:      dtoReq.Slug,
		Name:      dtoReq.Name,
		Plan:      dtoReq.Plan,
		OwnerID:   dtoReq.OwnerID,
		CreatedBy: createdBy,
	}

	tenantDTO, err := h.createTenantCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("tenant creation failed", logger.String("slug", cmdReq.Slug), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("tenant created", logger.String("tenant_id", tenantDTO.ID), logger.String("slug", tenantDTO.Slug))
	RespondCreated(w, tenantDTO)
}

// GetTenant godoc
// @Summary      Get tenant by ID
// @Description  Retrieves a tenant by its unique identifier
// @Tags         tenants
// @Produce      json
// @Param        id path string true "Tenant ID" format(uuid)
// @Success      200 {object} dto.TenantDTO "Tenant found"
// @Failure      400 {object} ErrorResponse "Invalid tenant ID format"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Tenant not found"
// @Router       /api/v1/tenants/{id} [get]
// @Security     BearerAuth
func (h *Handler) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, err := ParseTenantID(r)
	if err != nil {
		h.logger.Warn("invalid tenant ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	tenantDTO, err := h.getTenantQuery.Handle(r.Context(), query.GetTenantRequest{
		TenantID: tenantID.String(),
	})
	if err != nil {
		h.logger.Warn("get tenant failed", logger.String("tenant_id", tenantID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Debug("tenant retrieved", logger.String("tenant_id", tenantDTO.ID))
	RespondWithTenant(w, http.StatusOK, tenantDTO)
}

// GetTenantBySlug godoc
// @Summary      Get tenant by slug
// @Description  Retrieves a tenant by its URL-friendly slug identifier
// @Tags         tenants
// @Produce      json
// @Param        slug path string true "Tenant slug" example(acme-corp)
// @Success      200 {object} dto.TenantDTO "Tenant found"
// @Failure      400 {object} ErrorResponse "Invalid slug format"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Tenant not found"
// @Router       /api/v1/tenants/slug/{slug} [get]
// @Security     BearerAuth
func (h *Handler) GetTenantBySlug(w http.ResponseWriter, r *http.Request) {
	slug, err := ParseTenantSlug(r)
	if err != nil {
		h.logger.Warn("invalid tenant slug", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	tenantDTO, err := h.getTenantBySlugQuery.Handle(r.Context(), query.GetTenantBySlugRequest{
		Slug: slug,
	})
	if err != nil {
		h.logger.Warn("get tenant by slug failed", logger.String("slug", slug), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Debug("tenant retrieved by slug", logger.String("tenant_id", tenantDTO.ID), logger.String("slug", slug))
	RespondWithTenant(w, http.StatusOK, tenantDTO)
}

// ListTenants godoc
// @Summary      List tenants
// @Description  Retrieves a paginated list of tenants with optional filters
// @Tags         tenants
// @Produce      json
// @Param        status query string false "Filter by status" Enums(trialing, active, suspended, cancelled, deleted)
// @Param        plan query string false "Filter by plan" Enums(free, starter, pro, enterprise)
// @Param        owner_id query string false "Filter by owner user ID" format(uuid)
// @Param        search query string false "Search by name or slug"
// @Param        offset query int false "Pagination offset" default(0) minimum(0)
// @Param        limit query int false "Pagination limit" default(20) minimum(1) maximum(100)
// @Success      200 {object} dto.TenantListDTO "List of tenants"
// @Failure      400 {object} ErrorResponse "Invalid query parameters"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/v1/tenants [get]
// @Security     BearerAuth
func (h *Handler) ListTenants(w http.ResponseWriter, r *http.Request) {
	dtoReq, err := ParseListTenantsRequest(r)
	if err != nil {
		h.logger.Warn("invalid list tenants request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	listResp, err := h.listTenantsQuery.Handle(r.Context(), query.ListTenantsRequest{
		Status:  dtoReq.Status,
		Plan:    dtoReq.Plan,
		OwnerID: dtoReq.OwnerID,
		Search:  dtoReq.Search,
		Offset:  dtoReq.Offset,
		Limit:   dtoReq.Limit,
	})
	if err != nil {
		h.logger.Error("list tenants failed", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Debug("tenants listed", logger.Int("count", len(listResp.Tenants)), logger.Int64("total", listResp.Total))
	RespondWithTenantList(w, listResp)
}

// UpdateTenant godoc
// @Summary      Update tenant
// @Description  Updates tenant details. Requires admin privileges.
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        id path string true "Tenant ID" format(uuid)
// @Param        request body dto.UpdateTenantRequest true "Update tenant request"
// @Success      200 {object} dto.TenantDTO "Tenant updated successfully"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure      404 {object} ErrorResponse "Tenant not found"
// @Router       /api/v1/tenants/{id} [patch]
// @Security     BearerAuth
func (h *Handler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, err := ParseTenantID(r)
	if err != nil {
		h.logger.Warn("invalid tenant ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	dtoReq, err := ParseUpdateTenantRequest(r)
	if err != nil {
		h.logger.Warn("invalid update tenant request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	updatedBy := getActorFromContext(r)

	cmdReq := command.UpdateTenantRequest{
		TenantID:  tenantID.String(),
		Name:      dtoReq.Name,
		UpdatedBy: updatedBy,
	}

	tenantDTO, err := h.updateTenantCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("tenant update failed", logger.String("tenant_id", tenantID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("tenant updated", logger.String("tenant_id", tenantDTO.ID), logger.String("updated_by", updatedBy))
	RespondWithTenant(w, http.StatusOK, tenantDTO)
}

// UpdateSettings godoc
// @Summary      Update tenant settings
// @Description  Updates tenant-specific configuration settings. Requires admin privileges.
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        id path string true "Tenant ID" format(uuid)
// @Param        request body dto.UpdateSettingsRequest true "Update settings request"
// @Success      200 {object} dto.TenantSettingsDTO "Settings updated successfully"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure      404 {object} ErrorResponse "Tenant not found"
// @Router       /api/v1/tenants/{id}/settings [put]
// @Security     BearerAuth
func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	tenantID, err := ParseTenantID(r)
	if err != nil {
		h.logger.Warn("invalid tenant ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	dtoReq, err := ParseUpdateSettingsRequest(r)
	if err != nil {
		h.logger.Warn("invalid update settings request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	updatedBy := getActorFromContext(r)

	cmdReq := command.UpdateSettingsRequest{
		TenantID:  tenantID.String(),
		Settings:  dtoReq.Settings,
		UpdatedBy: updatedBy,
	}

	settingsDTO, err := h.updateSettingsCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("settings update failed", logger.String("tenant_id", tenantID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("tenant settings updated", logger.String("tenant_id", tenantID.String()), logger.String("updated_by", updatedBy))
	RespondWithSettings(w, settingsDTO)
}

// ChangePlan godoc
// @Summary      Change tenant plan
// @Description  Changes the tenant's subscription plan. Requires admin privileges.
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        id path string true "Tenant ID" format(uuid)
// @Param        request body dto.ChangePlanRequest true "Change plan request"
// @Success      200 {object} dto.TenantDTO "Plan changed successfully"
// @Failure      400 {object} ErrorResponse "Invalid plan"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure      404 {object} ErrorResponse "Tenant not found"
// @Failure      422 {object} ErrorResponse "Plan change not allowed"
// @Router       /api/v1/tenants/{id}/plan [post]
// @Security     BearerAuth
func (h *Handler) ChangePlan(w http.ResponseWriter, r *http.Request) {
	tenantID, err := ParseTenantID(r)
	if err != nil {
		h.logger.Warn("invalid tenant ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	dtoReq, err := ParseChangePlanRequest(r)
	if err != nil {
		h.logger.Warn("invalid change plan request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	changedBy := getActorFromContext(r)

	cmdReq := command.ChangePlanRequest{
		TenantID:  tenantID.String(),
		Plan:      dtoReq.Plan,
		Reason:    dtoReq.Reason,
		ChangedBy: changedBy,
	}

	tenantDTO, err := h.changePlanCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("plan change failed", logger.String("tenant_id", tenantID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("tenant plan changed", logger.String("tenant_id", tenantDTO.ID), logger.String("new_plan", tenantDTO.Plan), logger.String("changed_by", changedBy))
	RespondWithTenant(w, http.StatusOK, tenantDTO)
}

// SuspendTenant godoc
// @Summary      Suspend tenant
// @Description  Suspends a tenant, preventing access to their resources. Requires admin privileges.
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        id path string true "Tenant ID" format(uuid)
// @Param        request body dto.SuspendTenantRequest true "Suspend tenant request"
// @Success      200 {object} dto.TenantDTO "Tenant suspended successfully"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure      404 {object} ErrorResponse "Tenant not found"
// @Failure      422 {object} ErrorResponse "Invalid status transition"
// @Router       /api/v1/tenants/{id}/suspend [post]
// @Security     BearerAuth
func (h *Handler) SuspendTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, err := ParseTenantID(r)
	if err != nil {
		h.logger.Warn("invalid tenant ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	dtoReq, err := ParseSuspendTenantRequest(r)
	if err != nil {
		h.logger.Warn("invalid suspend tenant request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	suspendedBy := getActorFromContext(r)

	cmdReq := command.SuspendTenantRequest{
		TenantID:    tenantID.String(),
		Reason:      dtoReq.Reason,
		SuspendedBy: suspendedBy,
	}

	tenantDTO, err := h.suspendTenantCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("tenant suspension failed", logger.String("tenant_id", tenantID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("tenant suspended", logger.String("tenant_id", tenantDTO.ID), logger.String("suspended_by", suspendedBy))
	RespondWithTenant(w, http.StatusOK, tenantDTO)
}

// ReactivateTenant godoc
// @Summary      Reactivate tenant
// @Description  Reactivates a suspended tenant. Requires admin privileges.
// @Tags         tenants
// @Produce      json
// @Param        id path string true "Tenant ID" format(uuid)
// @Success      200 {object} dto.TenantDTO "Tenant reactivated successfully"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure      404 {object} ErrorResponse "Tenant not found"
// @Failure      422 {object} ErrorResponse "Invalid status transition"
// @Router       /api/v1/tenants/{id}/reactivate [post]
// @Security     BearerAuth
func (h *Handler) ReactivateTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, err := ParseTenantID(r)
	if err != nil {
		h.logger.Warn("invalid tenant ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	reactivatedBy := getActorFromContext(r)

	cmdReq := command.ReactivateTenantRequest{
		TenantID:      tenantID.String(),
		ReactivatedBy: reactivatedBy,
	}

	tenantDTO, err := h.reactivateTenantCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("tenant reactivation failed", logger.String("tenant_id", tenantID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("tenant reactivated", logger.String("tenant_id", tenantDTO.ID), logger.String("reactivated_by", reactivatedBy))
	RespondWithTenant(w, http.StatusOK, tenantDTO)
}

// DeleteTenant godoc
// @Summary      Delete tenant
// @Description  Soft-deletes a tenant. Requires admin privileges.
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        id path string true "Tenant ID" format(uuid)
// @Param        request body dto.DeleteTenantRequest false "Delete tenant request"
// @Success      204 "Tenant deleted successfully"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure      404 {object} ErrorResponse "Tenant not found"
// @Failure      422 {object} ErrorResponse "Invalid status transition"
// @Router       /api/v1/tenants/{id} [delete]
// @Security     BearerAuth
func (h *Handler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, err := ParseTenantID(r)
	if err != nil {
		h.logger.Warn("invalid tenant ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	dtoReq, err := ParseDeleteTenantRequest(r)
	if err != nil {
		h.logger.Warn("invalid delete tenant request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	deletedBy := getActorFromContext(r)

	cmdReq := command.DeleteTenantRequest{
		TenantID:  tenantID.String(),
		Reason:    dtoReq.Reason,
		DeletedBy: deletedBy,
	}

	_, err = h.deleteTenantCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("tenant deletion failed", logger.String("tenant_id", tenantID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("tenant deleted", logger.String("tenant_id", tenantID.String()), logger.String("deleted_by", deletedBy))
	RespondNoContent(w)
}

// Health godoc
// @Summary      Health check
// @Description  Returns OK if the service is healthy
// @Tags         system
// @Produce      json
// @Success      200 {object} map[string]string "Service is healthy"
// @Router       /health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	RespondWithMessage(w, "OK")
}

// ============================================================================
// Helper Functions
// ============================================================================

// getActorFromContext extracts the actor ID from the request context.
// Falls back to "system" if not found.
func getActorFromContext(r *http.Request) string {
	if userID := middleware.GetUserID(r.Context()); userID != "" {
		return userID
	}
	return "system"
}

// ============================================================================
// Swagger Models
// ============================================================================

// ErrorResponse represents an error response body.
// @Description Error response returned by the API
type ErrorResponse struct {
	Code    string         `json:"code" example:"TENANT_NOT_FOUND"`
	Message string         `json:"message" example:"tenant not found"`
	Meta    map[string]any `json:"meta,omitempty"`
}
