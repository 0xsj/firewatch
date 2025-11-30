package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/permissions/application/command"
	"github.com/0xsj/hexagonal-go/internal/permissions/application/query"
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Handler handles HTTP requests for the Permissions domain (v1 API).
type Handler struct {
	createRoleCmd           *command.CreateRoleCommand
	updateRoleCmd           *command.UpdateRoleCommand
	deleteRoleCmd           *command.DeleteRoleCommand
	assignRoleCmd           *command.AssignRoleCommand
	revokeRoleCmd           *command.RevokeRoleCommand
	getRoleQuery            *query.GetRoleQuery
	listRolesQuery          *query.ListRolesQuery
	getUserPermissionsQuery *query.GetUserPermissionsQuery
	checkPermissionQuery    *query.CheckPermissionQuery
	logger                  logger.Logger
}

// NewHandler creates a new v1 permissions HTTP handler.
func NewHandler(
	createRoleCmd *command.CreateRoleCommand,
	updateRoleCmd *command.UpdateRoleCommand,
	deleteRoleCmd *command.DeleteRoleCommand,
	assignRoleCmd *command.AssignRoleCommand,
	revokeRoleCmd *command.RevokeRoleCommand,
	getRoleQuery *query.GetRoleQuery,
	listRolesQuery *query.ListRolesQuery,
	getUserPermissionsQuery *query.GetUserPermissionsQuery,
	checkPermissionQuery *query.CheckPermissionQuery,
	log logger.Logger,
) *Handler {
	return &Handler{
		createRoleCmd:           createRoleCmd,
		updateRoleCmd:           updateRoleCmd,
		deleteRoleCmd:           deleteRoleCmd,
		assignRoleCmd:           assignRoleCmd,
		revokeRoleCmd:           revokeRoleCmd,
		getRoleQuery:            getRoleQuery,
		listRolesQuery:          listRolesQuery,
		getUserPermissionsQuery: getUserPermissionsQuery,
		checkPermissionQuery:    checkPermissionQuery,
		logger:                  log,
	}
}

// ============================================================================
// Role Endpoints
// ============================================================================

// CreateRole godoc
// @Summary      Create a new role
// @Description  Creates a new role with specified permissions
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        request body CreateRoleRequest true "Role creation request"
// @Success      201 {object} dto.RoleDTO "Role created successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      409 {object} ErrorResponse "Role already exists"
// @Router       /api/v1/roles [post]
// @Security     BearerAuth
func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	adminID := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())

	req, err := ParseCreateRoleRequest(r)
	if err != nil {
		h.logger.Warn("invalid create role request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	if req.TenantID == "" {
		req.TenantID = tenantID
	}

	cmdReq := command.CreateRoleRequest{
		TenantID:    req.TenantID,
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		CreatedBy:   adminID,
	}

	resp, err := h.createRoleCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("create role failed", logger.String("name", req.Name), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("role created", logger.String("role_id", resp.Role.ID), logger.String("name", resp.Role.Name))
	response.JSON(w, http.StatusCreated, resp.Role)
}

// GetRole godoc
// @Summary      Get role by ID
// @Description  Retrieves a role by its unique identifier
// @Tags         permissions
// @Produce      json
// @Param        id path string true "Role ID" format(uuid)
// @Success      200 {object} dto.RoleDTO "Role found"
// @Failure      400 {object} ErrorResponse "Invalid role ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Role not found"
// @Router       /api/v1/roles/{id} [get]
// @Security     BearerAuth
func (h *Handler) GetRole(w http.ResponseWriter, r *http.Request) {
	roleID, err := ParseRoleID(r)
	if err != nil {
		h.logger.Warn("invalid role ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	req := query.GetRoleByIDRequest{ID: roleID}
	resp, err := h.getRoleQuery.HandleByID(r.Context(), req)
	if err != nil {
		h.logger.Warn("get role failed", logger.String("role_id", roleID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("role retrieved", logger.String("role_id", resp.Role.ID))
	response.JSON(w, http.StatusOK, resp.Role)
}

// ListRoles godoc
// @Summary      List roles
// @Description  Retrieves a paginated list of roles
// @Tags         permissions
// @Produce      json
// @Param        include_system query bool false "Include system roles"
// @Param        search query string false "Search in name and description"
// @Param        limit query int false "Pagination limit" default(50)
// @Param        offset query int false "Pagination offset" default(0)
// @Success      200 {object} ListRolesResponse "List of roles"
// @Failure      400 {object} ErrorResponse "Invalid query parameters"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/v1/roles [get]
// @Security     BearerAuth
func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	req, err := ParseListRolesRequest(r)
	if err != nil {
		h.logger.Warn("invalid list roles request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	req.TenantID = tenantID

	resp, err := h.listRolesQuery.Handle(r.Context(), query.ListRolesRequest{
		TenantID:      req.TenantID,
		IncludeSystem: req.IncludeSystem,
		Search:        req.Search,
		Limit:         req.Limit,
		Offset:        req.Offset,
	})
	if err != nil {
		h.logger.Error("list roles failed", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("roles listed", logger.Int("count", len(resp.Roles)))
	response.JSON(w, http.StatusOK, resp)
}

// UpdateRole godoc
// @Summary      Update role
// @Description  Updates a role's name and description
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        id path string true "Role ID" format(uuid)
// @Param        request body UpdateRoleRequest true "Role update request"
// @Success      200 {object} dto.RoleDTO "Role updated successfully"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Cannot modify system role"
// @Failure      404 {object} ErrorResponse "Role not found"
// @Router       /api/v1/roles/{id} [put]
// @Security     BearerAuth
func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	adminID := middleware.GetUserID(r.Context())

	roleID, err := ParseRoleID(r)
	if err != nil {
		h.logger.Warn("invalid role ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	req, err := ParseUpdateRoleRequest(r)
	if err != nil {
		h.logger.Warn("invalid update role request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	cmdReq := command.UpdateRoleRequest{
		ID:          roleID,
		Name:        req.Name,
		Description: req.Description,
		UpdatedBy:   adminID,
	}

	resp, err := h.updateRoleCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("update role failed", logger.String("role_id", roleID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("role updated", logger.String("role_id", resp.Role.ID))
	response.JSON(w, http.StatusOK, resp.Role)
}

// DeleteRole godoc
// @Summary      Delete role
// @Description  Deletes a role and all its assignments
// @Tags         permissions
// @Produce      json
// @Param        id path string true "Role ID" format(uuid)
// @Success      204 "Role deleted successfully"
// @Failure      400 {object} ErrorResponse "Invalid role ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Cannot delete system role"
// @Failure      404 {object} ErrorResponse "Role not found"
// @Router       /api/v1/roles/{id} [delete]
// @Security     BearerAuth
func (h *Handler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	adminID := middleware.GetUserID(r.Context())

	roleID, err := ParseRoleID(r)
	if err != nil {
		h.logger.Warn("invalid role ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	cmdReq := command.DeleteRoleRequest{
		ID:        roleID,
		DeletedBy: adminID,
	}

	_, err = h.deleteRoleCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("delete role failed", logger.String("role_id", roleID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("role deleted", logger.String("role_id", roleID.String()))
	response.NoContent(w)
}

// ============================================================================
// Assignment Endpoints
// ============================================================================

// AssignRole godoc
// @Summary      Assign role to user
// @Description  Assigns a role to a user within a tenant
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        request body AssignRoleRequest true "Role assignment request"
// @Success      201 {object} dto.AssignmentDTO "Role assigned successfully"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Role not found"
// @Failure      409 {object} ErrorResponse "Assignment already exists"
// @Router       /api/v1/assignments [post]
// @Security     BearerAuth
func (h *Handler) AssignRole(w http.ResponseWriter, r *http.Request) {
	adminID := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())

	req, err := ParseAssignRoleRequest(r)
	if err != nil {
		h.logger.Warn("invalid assign role request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	userID, err := types.ParseID(req.UserID)
	if err != nil {
		h.logger.Warn("invalid user ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	roleID, err := types.ParseID(req.RoleID)
	if err != nil {
		h.logger.Warn("invalid role ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	if req.TenantID == "" {
		req.TenantID = tenantID
	}

	cmdReq := command.AssignRoleRequest{
		UserID:     userID,
		RoleID:     roleID,
		TenantID:   req.TenantID,
		ResourceID: req.ResourceID,
		ExpiresAt:  req.ExpiresAt,
		AssignedBy: adminID,
	}

	resp, err := h.assignRoleCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("assign role failed", logger.String("user_id", req.UserID), logger.String("role_id", req.RoleID), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("role assigned", logger.String("assignment_id", resp.Assignment.ID), logger.String("user_id", req.UserID), logger.String("role_id", req.RoleID))
	response.JSON(w, http.StatusCreated, resp.Assignment)
}

// RevokeRole godoc
// @Summary      Revoke role from user
// @Description  Revokes a role assignment from a user
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        request body RevokeRoleRequest true "Role revocation request"
// @Success      200 {object} MessageResponse "Role revoked successfully"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Assignment not found"
// @Router       /api/v1/assignments/revoke [post]
// @Security     BearerAuth
func (h *Handler) RevokeRole(w http.ResponseWriter, r *http.Request) {
	adminID := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())

	req, err := ParseRevokeRoleRequest(r)
	if err != nil {
		h.logger.Warn("invalid revoke role request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	userID, err := types.ParseID(req.UserID)
	if err != nil {
		h.logger.Warn("invalid user ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	roleID, err := types.ParseID(req.RoleID)
	if err != nil {
		h.logger.Warn("invalid role ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	if req.TenantID == "" {
		req.TenantID = tenantID
	}

	cmdReq := command.RevokeRoleRequest{
		UserID:     userID,
		RoleID:     roleID,
		TenantID:   req.TenantID,
		ResourceID: req.ResourceID,
		RevokedBy:  adminID,
		Reason:     req.Reason,
	}

	_, err = h.revokeRoleCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("revoke role failed", logger.String("user_id", req.UserID), logger.String("role_id", req.RoleID), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("role revoked", logger.String("user_id", req.UserID), logger.String("role_id", req.RoleID))
	RespondWithMessage(w, "Role revoked successfully")
}

// ============================================================================
// Permission Endpoints
// ============================================================================

// GetMyPermissions godoc
// @Summary      Get current user's permissions
// @Description  Retrieves all effective permissions for the current user
// @Tags         permissions
// @Produce      json
// @Success      200 {object} dto.UserPermissionsDTO "User permissions"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/v1/permissions/me [get]
// @Security     BearerAuth
func (h *Handler) GetMyPermissions(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())
	identityRole := middleware.GetRole(r.Context())

	userID, err := types.ParseID(userIDStr)
	if err != nil {
		h.logger.Error("invalid user_id", logger.String("user_id", userIDStr), logger.Err(err))
		response.BadRequest(w, "invalid user")
		return
	}

	req := query.GetUserPermissionsRequest{
		UserID:       userID,
		TenantID:     tenantID,
		IdentityRole: identityRole,
	}

	resp, err := h.getUserPermissionsQuery.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("get permissions failed", logger.String("user_id", userIDStr), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("permissions retrieved", logger.String("user_id", userIDStr), logger.Int("count", len(resp.Permissions.Permissions)))
	response.JSON(w, http.StatusOK, resp.Permissions)
}

// Health godoc
// @Summary      Permissions health check
// @Description  Returns OK if the permissions service is healthy
// @Tags         permissions
// @Produce      json
// @Success      200 {object} MessageResponse "Service is healthy"
// @Router       /api/v1/permissions/health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	RespondWithMessage(w, "permissions OK")
}

// CheckPermission godoc
// @Summary      Check if user has permission
// @Description  Checks if the current user has a specific permission
// @Tags         permissions
// @Produce      json
// @Param        action query string true "Action (create, read, update, delete, manage)"
// @Param        resource query string true "Resource (users, tenants, flags, roles, etc.)"
// @Param        resource_id query string false "Specific resource ID"
// @Success      200 {object} dto.PermissionCheckDTO "Permission check result"
// @Failure      400 {object} ErrorResponse "Invalid parameters"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/v1/permissions/check [get]
// @Security     BearerAuth
func (h *Handler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())
	identityRole := middleware.GetRole(r.Context())

	userID, err := types.ParseID(userIDStr)
	if err != nil {
		h.logger.Error("invalid user_id", logger.String("user_id", userIDStr), logger.Err(err))
		response.BadRequest(w, "invalid user")
		return
	}

	action := r.URL.Query().Get("action")
	resource := r.URL.Query().Get("resource")
	resourceID := r.URL.Query().Get("resource_id")

	if action == "" || resource == "" {
		response.BadRequest(w, "action and resource are required")
		return
	}

	req := query.CheckPermissionRequest{
		UserID:       userID,
		TenantID:     tenantID,
		IdentityRole: identityRole,
		Action:       action,
		Resource:     resource,
		ResourceID:   resourceID,
	}

	resp, err := h.checkPermissionQuery.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("check permission failed", logger.String("user_id", userIDStr), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("permission checked",
		logger.String("user_id", userIDStr),
		logger.String("permission", action+":"+resource),
		logger.Bool("allowed", resp.Result.Allowed),
	)
	response.JSON(w, http.StatusOK, resp.Result)
}

// ============================================================================
// Swagger Models
// ============================================================================

// ErrorResponse represents an error response body.
type ErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// MessageResponse represents a simple message response.
type MessageResponse struct {
	Message string `json:"message"`
}

// ListRolesResponse represents the list roles response.
type ListRolesResponse struct {
	Roles []RoleSummary `json:"roles"`
	Total int           `json:"total"`
}

// RoleSummary represents a role summary.
type RoleSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
	Permissions int    `json:"permissions_count"`
}
