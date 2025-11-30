package admin

import (
	"embed"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/hexagonal-go/internal/flags/application/command"
	"github.com/0xsj/hexagonal-go/internal/flags/application/query"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/templates"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

//go:embed templates/*.html
var templateFS embed.FS

// Handler handles HTTP requests for the flags admin dashboard.
type Handler struct {
	createCmd      *command.CreateFlagCommand
	updateCmd      *command.UpdateFlagCommand
	deleteCmd      *command.DeleteFlagCommand
	enableCmd      *command.EnableFlagCommand
	disableCmd     *command.DisableFlagCommand
	getFlagQuery   *query.GetFlagQuery
	listFlagsQuery *query.ListFlagsQuery
	renderer       *templates.Renderer
	logger         logger.Logger
}

// NewHandler creates a new admin dashboard handler.
func NewHandler(
	createCmd *command.CreateFlagCommand,
	updateCmd *command.UpdateFlagCommand,
	deleteCmd *command.DeleteFlagCommand,
	enableCmd *command.EnableFlagCommand,
	disableCmd *command.DisableFlagCommand,
	getFlagQuery *query.GetFlagQuery,
	listFlagsQuery *query.ListFlagsQuery,
	log logger.Logger,
) *Handler {
	renderer := templates.NewRenderer()
	if err := renderer.LoadFS(templateFS, "templates/layout.html", "templates/*.html"); err != nil {
		log.Error("failed to load admin templates", logger.Err(err))
	}

	return &Handler{
		createCmd:      createCmd,
		updateCmd:      updateCmd,
		deleteCmd:      deleteCmd,
		enableCmd:      enableCmd,
		disableCmd:     disableCmd,
		getFlagQuery:   getFlagQuery,
		listFlagsQuery: listFlagsQuery,
		renderer:       renderer,
		logger:         log,
	}
}

// PageData holds common data for all pages.
type PageData struct {
	Title   string
	Content any
	Error   string
	Success string
}

// ============================================================================
// Page Handlers
// ============================================================================

// Index renders the flags list page.
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	req := query.ListFlagsRequest{
		TenantID: tenantID,
		Limit:    100,
		Offset:   0,
	}

	resp, err := h.listFlagsQuery.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to list flags", logger.Err(err))
		h.renderError(w, "Failed to load flags", err)
		return
	}

	data := PageData{
		Title: "Feature Flags",
		Content: map[string]any{
			"Flags":    resp.Flags,
			"TenantID": tenantID,
		},
	}

	if err := h.renderer.HTML(w, http.StatusOK, "index", data); err != nil {
		h.logger.Error("failed to render index", logger.Err(err))
	}
}

// New renders the create flag form.
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	data := PageData{
		Title: "Create Flag",
		Content: map[string]any{
			"TenantID": tenantID,
		},
	}

	if err := h.renderer.HTML(w, http.StatusOK, "new", data); err != nil {
		h.logger.Error("failed to render new form", logger.Err(err))
	}
}

// Create handles the create flag form submission.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderError(w, "Invalid form data", err)
		return
	}

	tenantID := r.FormValue("tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	req := command.CreateFlagRequest{
		TenantID:    tenantID,
		Key:         r.FormValue("key"),
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Enabled:     r.FormValue("enabled") == "on",
		CreatedBy:   "admin",
	}

	_, err := h.createCmd.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to create flag", logger.Err(err))
		data := PageData{
			Title: "Create Flag",
			Error: err.Error(),
			Content: map[string]any{
				"TenantID": tenantID,
				"Form":     req,
			},
		}
		h.renderer.HTML(w, http.StatusBadRequest, "new", data)
		return
	}

	// Redirect to index with success message
	w.Header().Set("HX-Redirect", "/admin/flags?tenant_id="+tenantID+"&success=Flag+created")
	w.WriteHeader(http.StatusOK)
}

// Show renders the flag detail page.
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	flagID, err := h.parseFlagID(r)
	if err != nil {
		h.renderError(w, "Invalid flag ID", err)
		return
	}

	req := query.GetFlagByIDRequest{ID: flagID}
	resp, err := h.getFlagQuery.HandleByID(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to get flag", logger.Err(err))
		h.renderError(w, "Flag not found", err)
		return
	}

	data := PageData{
		Title:   "Flag: " + resp.Flag.Name,
		Content: resp.Flag,
	}

	if err := h.renderer.HTML(w, http.StatusOK, "show", data); err != nil {
		h.logger.Error("failed to render show", logger.Err(err))
	}
}

// Edit renders the edit flag form.
func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	flagID, err := h.parseFlagID(r)
	if err != nil {
		h.renderError(w, "Invalid flag ID", err)
		return
	}

	req := query.GetFlagByIDRequest{ID: flagID}
	resp, err := h.getFlagQuery.HandleByID(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to get flag", logger.Err(err))
		h.renderError(w, "Flag not found", err)
		return
	}

	data := PageData{
		Title:   "Edit: " + resp.Flag.Name,
		Content: resp.Flag,
	}

	if err := h.renderer.HTML(w, http.StatusOK, "edit", data); err != nil {
		h.logger.Error("failed to render edit", logger.Err(err))
	}
}

// Update handles the update flag form submission.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	flagID, err := h.parseFlagID(r)
	if err != nil {
		h.renderError(w, "Invalid flag ID", err)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderError(w, "Invalid form data", err)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	req := command.UpdateFlagRequest{
		ID:          flagID,
		Name:        &name,
		Description: &description,
		UpdatedBy:   "admin",
	}

	_, err = h.updateCmd.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to update flag", logger.Err(err))
		h.renderError(w, "Failed to update flag", err)
		return
	}

	// Redirect back to show page
	w.Header().Set("HX-Redirect", "/admin/flags/"+flagID.String())
	w.WriteHeader(http.StatusOK)
}

// Delete handles flag deletion.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	flagID, err := h.parseFlagID(r)
	if err != nil {
		h.renderError(w, "Invalid flag ID", err)
		return
	}

	req := command.DeleteFlagRequest{
		ID:        flagID,
		DeletedBy: "admin",
	}

	_, err = h.deleteCmd.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to delete flag", logger.Err(err))
		h.renderError(w, "Failed to delete flag", err)
		return
	}

	// Redirect to index
	w.Header().Set("HX-Redirect", "/admin/flags")
	w.WriteHeader(http.StatusOK)
}

// ============================================================================
// HTMX Partial Handlers
// ============================================================================

// Toggle handles enabling/disabling a flag via HTMX.
func (h *Handler) Toggle(w http.ResponseWriter, r *http.Request) {
	flagID, err := h.parseFlagID(r)
	if err != nil {
		h.htmxError(w, "Invalid flag ID")
		return
	}

	// Get current state
	getReq := query.GetFlagByIDRequest{ID: flagID}
	resp, err := h.getFlagQuery.HandleByID(r.Context(), getReq)
	if err != nil {
		h.htmxError(w, "Flag not found")
		return
	}

	// Toggle state
	if resp.Flag.Enabled {
		disableReq := command.DisableFlagRequest{
			ID:         flagID,
			DisabledBy: "admin",
		}
		_, err = h.disableCmd.Handle(r.Context(), disableReq)
	} else {
		enableReq := command.EnableFlagRequest{
			ID:        flagID,
			EnabledBy: "admin",
		}
		_, err = h.enableCmd.Handle(r.Context(), enableReq)
	}

	if err != nil {
		h.logger.Error("failed to toggle flag", logger.Err(err))
		h.htmxError(w, "Failed to toggle flag")
		return
	}

	// Return updated toggle button
	newState := !resp.Flag.Enabled
	h.renderToggleButton(w, flagID.String(), newState)
}

// FlagRow renders a single flag row for HTMX updates.
func (h *Handler) FlagRow(w http.ResponseWriter, r *http.Request) {
	flagID, err := h.parseFlagID(r)
	if err != nil {
		h.htmxError(w, "Invalid flag ID")
		return
	}

	req := query.GetFlagByIDRequest{ID: flagID}
	resp, err := h.getFlagQuery.HandleByID(r.Context(), req)
	if err != nil {
		h.htmxError(w, "Flag not found")
		return
	}

	if err := h.renderer.HTMLPartial(w, http.StatusOK, "index", "flag-row", resp.Flag); err != nil {
		h.logger.Error("failed to render flag row", logger.Err(err))
	}
}

// ============================================================================
// Helpers
// ============================================================================

func (h *Handler) parseFlagID(r *http.Request) (types.ID, error) {
	idStr := chi.URLParam(r, "id")
	return types.ParseID(idStr)
}

func (h *Handler) renderError(w http.ResponseWriter, message string, err error) {
	data := PageData{
		Title: "Error",
		Error: message + ": " + err.Error(),
	}
	h.renderer.HTML(w, http.StatusInternalServerError, "error", data)
}

func (h *Handler) htmxError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(`<div class="error">` + message + `</div>`))
}

func (h *Handler) renderToggleButton(w http.ResponseWriter, flagID string, enabled bool) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var html string
	if enabled {
		html = `<button 
			class="toggle-btn enabled"
			hx-post="/admin/flags/` + flagID + `/toggle"
			hx-swap="outerHTML"
			title="Click to disable">
			ON
		</button>`
	} else {
		html = `<button 
			class="toggle-btn disabled"
			hx-post="/admin/flags/` + flagID + `/toggle"
			hx-swap="outerHTML"
			title="Click to enable">
			OFF
		</button>`
	}

	w.Write([]byte(html))
}
