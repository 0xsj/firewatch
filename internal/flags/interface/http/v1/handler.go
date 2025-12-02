package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/flags/application/command"
	_ "github.com/0xsj/hexagonal-go/internal/flags/application/dto"
	"github.com/0xsj/hexagonal-go/internal/flags/application/query"
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for feature flags.
type Handler struct {
	createCmd         *command.CreateFlagCommand
	updateCmd         *command.UpdateFlagCommand
	deleteCmd         *command.DeleteFlagCommand
	enableCmd         *command.EnableFlagCommand
	disableCmd        *command.DisableFlagCommand
	addRuleCmd        *command.AddRuleCommand
	removeRuleCmd     *command.RemoveRuleCommand
	setOverrideCmd    *command.SetOverrideCommand
	removeOverrideCmd *command.RemoveOverrideCommand
	getFlagQuery      *query.GetFlagQuery
	listFlagsQuery    *query.ListFlagsQuery
	evaluateFlagQuery *query.EvaluateFlagQuery
	logger            logger.Logger
}

// NewHandler creates a new feature flags handler.
func NewHandler(
	createCmd *command.CreateFlagCommand,
	updateCmd *command.UpdateFlagCommand,
	deleteCmd *command.DeleteFlagCommand,
	enableCmd *command.EnableFlagCommand,
	disableCmd *command.DisableFlagCommand,
	addRuleCmd *command.AddRuleCommand,
	removeRuleCmd *command.RemoveRuleCommand,
	setOverrideCmd *command.SetOverrideCommand,
	removeOverrideCmd *command.RemoveOverrideCommand,
	getFlagQuery *query.GetFlagQuery,
	listFlagsQuery *query.ListFlagsQuery,
	evaluateFlagQuery *query.EvaluateFlagQuery,
	logger logger.Logger,
) *Handler {
	return &Handler{
		createCmd:         createCmd,
		updateCmd:         updateCmd,
		deleteCmd:         deleteCmd,
		enableCmd:         enableCmd,
		disableCmd:        disableCmd,
		addRuleCmd:        addRuleCmd,
		removeRuleCmd:     removeRuleCmd,
		setOverrideCmd:    setOverrideCmd,
		removeOverrideCmd: removeOverrideCmd,
		getFlagQuery:      getFlagQuery,
		listFlagsQuery:    listFlagsQuery,
		evaluateFlagQuery: evaluateFlagQuery,
		logger:            logger,
	}
}

// CreateFlag godoc
// @Summary      Create feature flag
// @Description  Creates a new feature flag
// @Tags         flags
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateFlagRequest true "Create flag request"
// @Success      201 {object} dto.CreateFlagResponse "Flag created successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      409 {object} ErrorResponse "Flag with key already exists"
// @Router       /api/v1/flags [post]
// @Security     BearerAuth
func (h *Handler) CreateFlag(w http.ResponseWriter, r *http.Request) {
	req, err := ParseCreateFlagRequest(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	userID := middleware.GetUserID(r.Context())

	cmdReq := command.CreateFlagRequest{
		TenantID:    tenantID,
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		CreatedBy:   userID,
	}

	resp, err := h.createCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("failed to create flag", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("flag created",
		logger.String("flag_id", resp.Flag.ID),
		logger.String("flag_key", resp.Flag.Key),
	)
	RespondWithFlagCreated(w, resp)
}

// UpdateFlag godoc
// @Summary      Update feature flag
// @Description  Updates an existing feature flag
// @Tags         flags
// @Accept       json
// @Produce      json
// @Param        id path string true "Flag ID" format(uuid)
// @Param        request body dto.UpdateFlagRequest true "Update flag request"
// @Success      200 {object} dto.UpdateFlagResponse "Flag updated successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body or flag ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Flag not found"
// @Router       /api/v1/flags/{id} [put]
// @Security     BearerAuth
func (h *Handler) UpdateFlag(w http.ResponseWriter, r *http.Request) {
	flagID, err := ParseFlagID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	req, err := ParseUpdateFlagRequest(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	userID := middleware.GetUserID(r.Context())

	cmdReq := command.UpdateFlagRequest{
		ID:          flagID,
		Name:        req.Name,
		Description: req.Description,
		UpdatedBy:   userID,
	}

	resp, err := h.updateCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("failed to update flag", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("flag updated", logger.String("flag_id", flagID.String()))
	RespondWithFlagUpdated(w, resp)
}

// DeleteFlag godoc
// @Summary      Delete feature flag
// @Description  Deletes a feature flag
// @Tags         flags
// @Produce      json
// @Param        id path string true "Flag ID" format(uuid)
// @Success      200 {object} dto.DeleteFlagResponse "Flag deleted successfully"
// @Failure      400 {object} ErrorResponse "Invalid flag ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Flag not found"
// @Router       /api/v1/flags/{id} [delete]
// @Security     BearerAuth
func (h *Handler) DeleteFlag(w http.ResponseWriter, r *http.Request) {
	flagID, err := ParseFlagID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	userID := middleware.GetUserID(r.Context())

	cmdReq := command.DeleteFlagRequest{
		ID:        flagID,
		DeletedBy: userID,
	}

	resp, err := h.deleteCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("failed to delete flag", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("flag deleted", logger.String("flag_id", flagID.String()))
	RespondWithFlagDeleted(w, resp)
}

// EnableFlag godoc
// @Summary      Enable feature flag
// @Description  Enables a feature flag
// @Tags         flags
// @Produce      json
// @Param        id path string true "Flag ID" format(uuid)
// @Success      200 {object} dto.EnableFlagResponse "Flag enabled successfully"
// @Failure      400 {object} ErrorResponse "Invalid flag ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Flag not found"
// @Router       /api/v1/flags/{id}/enable [post]
// @Security     BearerAuth
func (h *Handler) EnableFlag(w http.ResponseWriter, r *http.Request) {
	flagID, err := ParseFlagID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	userID := middleware.GetUserID(r.Context())

	cmdReq := command.EnableFlagRequest{
		ID:        flagID,
		EnabledBy: userID,
	}

	resp, err := h.enableCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("failed to enable flag", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("flag enabled", logger.String("flag_id", flagID.String()))
	RespondWithFlagEnabled(w, resp)
}

// DisableFlag godoc
// @Summary      Disable feature flag
// @Description  Disables a feature flag
// @Tags         flags
// @Produce      json
// @Param        id path string true "Flag ID" format(uuid)
// @Success      200 {object} dto.DisableFlagResponse "Flag disabled successfully"
// @Failure      400 {object} ErrorResponse "Invalid flag ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Flag not found"
// @Router       /api/v1/flags/{id}/disable [post]
// @Security     BearerAuth
func (h *Handler) DisableFlag(w http.ResponseWriter, r *http.Request) {
	flagID, err := ParseFlagID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	userID := middleware.GetUserID(r.Context())

	cmdReq := command.DisableFlagRequest{
		ID:         flagID,
		DisabledBy: userID,
	}

	resp, err := h.disableCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("failed to disable flag", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("flag disabled", logger.String("flag_id", flagID.String()))
	RespondWithFlagDisabled(w, resp)
}

// GetFlag godoc
// @Summary      Get feature flag by ID
// @Description  Retrieves a feature flag by its unique identifier
// @Tags         flags
// @Produce      json
// @Param        id path string true "Flag ID" format(uuid)
// @Success      200 {object} dto.GetFlagResponse "Flag found"
// @Failure      400 {object} ErrorResponse "Invalid flag ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Flag not found"
// @Router       /api/v1/flags/{id} [get]
// @Security     BearerAuth
func (h *Handler) GetFlag(w http.ResponseWriter, r *http.Request) {
	flagID, err := ParseFlagID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	req := query.GetFlagByIDRequest{
		ID: flagID,
	}

	resp, err := h.getFlagQuery.HandleByID(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to get flag", logger.Err(err))
		HandleError(w, err)
		return
	}

	RespondWithFlag(w, http.StatusOK, resp)
}

// GetFlagByKey godoc
// @Summary      Get feature flag by key
// @Description  Retrieves a feature flag by its key
// @Tags         flags
// @Produce      json
// @Param        key query string true "Flag key"
// @Success      200 {object} dto.GetFlagResponse "Flag found"
// @Failure      400 {object} ErrorResponse "Key is required"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Flag not found"
// @Router       /api/v1/flags/by-key [get]
// @Security     BearerAuth
func (h *Handler) GetFlagByKey(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		response.BadRequest(w, "key is required")
		return
	}

	tenantID := middleware.GetTenantID(r.Context())

	req := query.GetFlagByKeyRequest{
		TenantID: tenantID,
		Key:      key,
	}

	resp, err := h.getFlagQuery.HandleByKey(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to get flag by key", logger.Err(err))
		HandleError(w, err)
		return
	}

	RespondWithFlag(w, http.StatusOK, resp)
}

// ListFlags godoc
// @Summary      List feature flags
// @Description  Retrieves a paginated list of feature flags
// @Tags         flags
// @Produce      json
// @Param        enabled query bool false "Filter by enabled status"
// @Param        search query string false "Search by key, name, or description"
// @Param        limit query int false "Pagination limit" default(50)
// @Param        offset query int false "Pagination offset" default(0)
// @Success      200 {object} dto.ListFlagsResponse "List of flags"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/v1/flags [get]
// @Security     BearerAuth
func (h *Handler) ListFlags(w http.ResponseWriter, r *http.Request) {
	req := ParseListFlagsRequest(r)
	tenantID := middleware.GetTenantID(r.Context())

	queryReq := query.ListFlagsRequest{
		TenantID: tenantID,
		Enabled:  req.Enabled,
		Search:   req.Search,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}

	resp, err := h.listFlagsQuery.Handle(r.Context(), queryReq)
	if err != nil {
		h.logger.Error("failed to list flags", logger.Err(err))
		HandleError(w, err)
		return
	}

	RespondWithFlagList(w, resp)
}

// AddRule godoc
// @Summary      Add targeting rule
// @Description  Adds a targeting rule to a feature flag
// @Tags         flags
// @Accept       json
// @Produce      json
// @Param        id path string true "Flag ID" format(uuid)
// @Param        request body dto.AddRuleRequest true "Add rule request"
// @Success      201 {object} dto.AddRuleResponse "Rule added successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body or flag ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Flag not found"
// @Router       /api/v1/flags/{id}/rules [post]
// @Security     BearerAuth
func (h *Handler) AddRule(w http.ResponseWriter, r *http.Request) {
	flagID, err := ParseFlagID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	req, err := ParseAddRuleRequest(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	userID := middleware.GetUserID(r.Context())

	cmdReq := command.AddRuleRequest{
		FlagID:     flagID,
		Type:       req.Type,
		Attribute:  req.Attribute,
		Operator:   req.Operator,
		Values:     req.Values,
		Percentage: req.Percentage,
		VariantKey: req.VariantKey,
		Priority:   req.Priority,
		AddedBy:    userID,
	}

	resp, err := h.addRuleCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("failed to add rule", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("rule added",
		logger.String("flag_id", flagID.String()),
		logger.String("rule_id", resp.RuleID),
	)
	RespondWithRuleAdded(w, resp)
}

// RemoveRule godoc
// @Summary      Remove targeting rule
// @Description  Removes a targeting rule from a feature flag
// @Tags         flags
// @Produce      json
// @Param        id path string true "Flag ID" format(uuid)
// @Param        ruleId path string true "Rule ID" format(uuid)
// @Success      200 {object} dto.RemoveRuleResponse "Rule removed successfully"
// @Failure      400 {object} ErrorResponse "Invalid flag ID or rule ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Flag or rule not found"
// @Router       /api/v1/flags/{id}/rules/{ruleId} [delete]
// @Security     BearerAuth
func (h *Handler) RemoveRule(w http.ResponseWriter, r *http.Request) {
	flagID, err := ParseFlagID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	ruleID, err := ParseRuleID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	userID := middleware.GetUserID(r.Context())

	cmdReq := command.RemoveRuleRequest{
		FlagID:    flagID,
		RuleID:    ruleID,
		RemovedBy: userID,
	}

	resp, err := h.removeRuleCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("failed to remove rule", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("rule removed",
		logger.String("flag_id", flagID.String()),
		logger.String("rule_id", ruleID.String()),
	)
	RespondWithRuleRemoved(w, resp)
}

// SetOverride godoc
// @Summary      Set override
// @Description  Sets an override for a specific target (tenant or user)
// @Tags         flags
// @Accept       json
// @Produce      json
// @Param        id path string true "Flag ID" format(uuid)
// @Param        request body dto.SetOverrideRequest true "Set override request"
// @Success      200 {object} dto.SetOverrideResponse "Override set successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body or flag ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Flag not found"
// @Router       /api/v1/flags/{id}/overrides [post]
// @Security     BearerAuth
func (h *Handler) SetOverride(w http.ResponseWriter, r *http.Request) {
	flagID, err := ParseFlagID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	req, err := ParseSetOverrideRequest(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	userID := middleware.GetUserID(r.Context())

	cmdReq := command.SetOverrideRequest{
		FlagID:     flagID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		VariantKey: req.VariantKey,
		SetBy:      userID,
	}

	resp, err := h.setOverrideCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("failed to set override", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("override set",
		logger.String("flag_id", flagID.String()),
		logger.String("target_type", req.TargetType),
		logger.String("target_id", req.TargetID),
	)
	RespondWithOverrideSet(w, resp)
}

// RemoveOverride godoc
// @Summary      Remove override
// @Description  Removes an override for a specific target
// @Tags         flags
// @Accept       json
// @Produce      json
// @Param        id path string true "Flag ID" format(uuid)
// @Param        request body dto.RemoveOverrideRequest true "Remove override request"
// @Success      200 {object} dto.RemoveOverrideResponse "Override removed successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body or flag ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Flag or override not found"
// @Router       /api/v1/flags/{id}/overrides [delete]
// @Security     BearerAuth
func (h *Handler) RemoveOverride(w http.ResponseWriter, r *http.Request) {
	flagID, err := ParseFlagID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	req, err := ParseRemoveOverrideRequest(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	userID := middleware.GetUserID(r.Context())

	cmdReq := command.RemoveOverrideRequest{
		FlagID:     flagID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		RemovedBy:  userID,
	}

	resp, err := h.removeOverrideCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("failed to remove override", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("override removed",
		logger.String("flag_id", flagID.String()),
		logger.String("target_type", req.TargetType),
		logger.String("target_id", req.TargetID),
	)
	RespondWithOverrideRemoved(w, resp)
}

// EvaluateFlag godoc
// @Summary      Evaluate feature flag
// @Description  Evaluates a feature flag for the given context
// @Tags         flags
// @Accept       json
// @Produce      json
// @Param        key path string true "Flag key"
// @Param        request body dto.EvaluateFlagRequest true "Evaluation context"
// @Success      200 {object} dto.EvaluateFlagResponse "Evaluation result"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Flag not found"
// @Router       /api/v1/flags/{key}/evaluate [post]
// @Security     BearerAuth
func (h *Handler) EvaluateFlag(w http.ResponseWriter, r *http.Request) {
	key := getFlagKeyFromPath(r)
	if key == "" {
		response.BadRequest(w, "flag key is required")
		return
	}

	req, err := ParseEvaluateFlagRequest(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	if req.TenantID == "" {
		req.TenantID = tenantID
	}

	userID := middleware.GetUserID(r.Context())
	if req.UserID == "" {
		req.UserID = userID
	}

	queryReq := query.EvaluateFlagRequest{
		TenantID:   req.TenantID,
		FlagKey:    key,
		UserID:     req.UserID,
		Attributes: req.Attributes,
	}

	resp, err := h.evaluateFlagQuery.Handle(r.Context(), queryReq)
	if err != nil {
		h.logger.Error("failed to evaluate flag", logger.Err(err))
		HandleError(w, err)
		return
	}

	RespondWithEvaluation(w, resp)
}

// ============================================================================
// Helpers
// ============================================================================

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

func getFlagKeyFromPath(r *http.Request) string {
	return chi.URLParam(r, "key")
}

// ============================================================================
// Swagger Models
// ============================================================================

// ErrorResponse represents an error response body.
// @Description Error response returned by the API
type ErrorResponse struct {
	Code    string         `json:"code" example:"FLAG_NOT_FOUND"`
	Message string         `json:"message" example:"flag not found"`
	Meta    map[string]any `json:"meta,omitempty"`
}
