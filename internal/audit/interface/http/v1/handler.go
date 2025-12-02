// internal/audit/interface/http/v1/handler.go
package v1

import (
	"net/http"

	_ "github.com/0xsj/hexagonal-go/internal/audit/application/dto"
	"github.com/0xsj/hexagonal-go/internal/audit/application/query"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// Handler handles HTTP requests for the Audit domain (v1 API).
type Handler struct {
	getEntryQuery         *query.GetEntryQuery
	listEntriesQuery      *query.ListEntriesQuery
	getResourceTrailQuery *query.GetResourceTrailQuery
	getActorActivityQuery *query.GetActorActivityQuery
	logger                logger.Logger
}

// NewHandler creates a new v1 audit HTTP handler.
func NewHandler(
	getEntryQuery *query.GetEntryQuery,
	listEntriesQuery *query.ListEntriesQuery,
	getResourceTrailQuery *query.GetResourceTrailQuery,
	getActorActivityQuery *query.GetActorActivityQuery,
	log logger.Logger,
) *Handler {
	return &Handler{
		getEntryQuery:         getEntryQuery,
		listEntriesQuery:      listEntriesQuery,
		getResourceTrailQuery: getResourceTrailQuery,
		getActorActivityQuery: getActorActivityQuery,
		logger:                log,
	}
}

// GetEntry godoc
// @Summary      Get audit entry
// @Description  Retrieves a single audit entry by its ID
// @Tags         audit
// @Produce      json
// @Param        id path string true "Audit entry ID" format(uuid)
// @Success      200 {object} dto.GetEntryResponse "Audit entry found"
// @Failure      400 {object} ErrorResponse "Invalid entry ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Entry not found"
// @Router       /api/v1/audit/entries/{id} [get]
// @Security     BearerAuth
func (h *Handler) GetEntry(w http.ResponseWriter, r *http.Request) {
	req, err := ParseGetEntryRequest(r)
	if err != nil {
		h.logger.Warn("invalid get entry request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	resp, err := h.getEntryQuery.Handle(r.Context(), req)
	if err != nil {
		h.logger.Warn("get audit entry failed", logger.String("id", req.ID), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("audit entry retrieved", logger.String("id", req.ID))
	RespondWithEntry(w, resp)
}

// ListEntries godoc
// @Summary      List audit entries
// @Description  Retrieves audit entries with optional filtering and pagination
// @Tags         audit
// @Produce      json
// @Param        tenant_id query string false "Filter by tenant ID"
// @Param        user_id query string false "Filter by user ID (actor)"
// @Param        event_type query string false "Filter by event type"
// @Param        source query string false "Filter by source domain"
// @Param        correlation_id query string false "Filter by correlation ID"
// @Param        from query string false "Filter from timestamp (RFC3339)"
// @Param        to query string false "Filter to timestamp (RFC3339)"
// @Param        limit query int false "Page size (default 50, max 100)"
// @Param        offset query int false "Page offset (default 0)"
// @Success      200 {object} dto.ListEntriesResponse "List of audit entries"
// @Failure      400 {object} ErrorResponse "Invalid query parameters"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/v1/audit/entries [get]
// @Security     BearerAuth
func (h *Handler) ListEntries(w http.ResponseWriter, r *http.Request) {
	req, err := ParseListEntriesRequest(r)
	if err != nil {
		h.logger.Warn("invalid list entries request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	resp, err := h.listEntriesQuery.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("list audit entries failed", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("audit entries listed",
		logger.Int("count", len(resp.Entries)),
		logger.Int("total", resp.TotalCount),
	)
	RespondWithEntryList(w, resp)
}

// GetResourceTrail godoc
// @Summary      Get resource audit trail
// @Description  Retrieves all audit entries for a specific resource
// @Tags         audit
// @Produce      json
// @Param        type path string true "Resource type (e.g., user, tenant, flag)"
// @Param        id path string true "Resource ID"
// @Param        tenant_id query string false "Filter by tenant ID"
// @Param        from query string false "Filter from timestamp (RFC3339)"
// @Param        to query string false "Filter to timestamp (RFC3339)"
// @Param        limit query int false "Page size (default 50, max 100)"
// @Param        offset query int false "Page offset (default 0)"
// @Success      200 {object} dto.GetResourceTrailResponse "Resource audit trail"
// @Failure      400 {object} ErrorResponse "Invalid parameters"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/v1/audit/resources/{type}/{id} [get]
// @Security     BearerAuth
func (h *Handler) GetResourceTrail(w http.ResponseWriter, r *http.Request) {
	req, err := ParseGetResourceTrailRequest(r)
	if err != nil {
		h.logger.Warn("invalid resource trail request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	resp, err := h.getResourceTrailQuery.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("get resource audit trail failed",
			logger.String("resource_type", req.ResourceType),
			logger.String("resource_id", req.ResourceID),
			logger.Err(err),
		)
		HandleError(w, err)
		return
	}

	h.logger.Info("resource audit trail retrieved",
		logger.String("resource_type", req.ResourceType),
		logger.String("resource_id", req.ResourceID),
		logger.Int("count", len(resp.Entries)),
		logger.Int("total", resp.TotalCount),
	)
	RespondWithResourceTrail(w, resp)
}

// GetActorActivity godoc
// @Summary      Get actor activity
// @Description  Retrieves all audit entries performed by a specific user
// @Tags         audit
// @Produce      json
// @Param        userID path string true "User ID"
// @Param        tenant_id query string false "Filter by tenant ID"
// @Param        event_type query string false "Filter by event type"
// @Param        source query string false "Filter by source domain"
// @Param        from query string false "Filter from timestamp (RFC3339)"
// @Param        to query string false "Filter to timestamp (RFC3339)"
// @Param        limit query int false "Page size (default 50, max 100)"
// @Param        offset query int false "Page offset (default 0)"
// @Success      200 {object} dto.GetActorActivityResponse "Actor activity"
// @Failure      400 {object} ErrorResponse "Invalid parameters"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/v1/audit/actors/{userID} [get]
// @Security     BearerAuth
func (h *Handler) GetActorActivity(w http.ResponseWriter, r *http.Request) {
	req, err := ParseGetActorActivityRequest(r)
	if err != nil {
		h.logger.Warn("invalid actor activity request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	resp, err := h.getActorActivityQuery.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("get actor activity failed",
			logger.String("user_id", req.UserID),
			logger.Err(err),
		)
		HandleError(w, err)
		return
	}

	h.logger.Info("actor activity retrieved",
		logger.String("user_id", req.UserID),
		logger.Int("count", len(resp.Entries)),
		logger.Int("total", resp.TotalCount),
	)
	RespondWithActorActivity(w, resp)
}

// Health godoc
// @Summary      Audit health check
// @Description  Returns OK if the audit service is healthy
// @Tags         audit
// @Produce      json
// @Success      200 {object} MessageResponse "Service is healthy"
// @Router       /api/v1/audit/health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	RespondWithMessage(w, "audit OK")
}

// ============================================================================
// Swagger Models
// ============================================================================

// ErrorResponse represents an error response body.
// @Description Error response returned by the API
type ErrorResponse struct {
	Code    string         `json:"code" example:"ENTRY_NOT_FOUND"`
	Message string         `json:"message" example:"audit entry not found"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// MessageResponse represents a simple message response.
// @Description Simple message response
type MessageResponse struct {
	Message string `json:"message" example:"audit OK"`
}
