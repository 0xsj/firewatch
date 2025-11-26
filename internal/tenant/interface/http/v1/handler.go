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

// CreateTenant handles tenant creation.
// POST /api/v1/tenants
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

// GetTenant retrieves a tenant by ID.
// GET /api/v1/tenants/{id}
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

// GetTenantBySlug retrieves a tenant by slug.
// GET /api/v1/tenants/slug/{slug}
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

// ListTenants retrieves a paginated list of tenants.
// GET /api/v1/tenants
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

// UpdateTenant updates tenant details.
// PATCH /api/v1/tenants/{id}
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

// UpdateSettings updates tenant settings.
// PUT /api/v1/tenants/{id}/settings
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

// ChangePlan changes the tenant's subscription plan.
// POST /api/v1/tenants/{id}/plan
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

// SuspendTenant suspends a tenant.
// POST /api/v1/tenants/{id}/suspend
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

// ReactivateTenant reactivates a suspended tenant.
// POST /api/v1/tenants/{id}/reactivate
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

// DeleteTenant soft-deletes a tenant.
// DELETE /api/v1/tenants/{id}
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

// Health handles health check requests.
// GET /health
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
