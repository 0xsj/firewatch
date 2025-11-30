package demo

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/pkg/flags"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// Handler demonstrates feature flag usage.
type Handler struct {
	flagsClient flags.Client
	logger      logger.Logger
}

// NewHandler creates a new demo handler.
func NewHandler(flagsClient flags.Client, logger logger.Logger) *Handler {
	return &Handler{
		flagsClient: flagsClient,
		logger:      logger,
	}
}

// GetUsers demonstrates v1/v2 route switching based on feature flag.
//
// GET /demo/users
//
// If flag "v2_test" is enabled: returns v2 response format
// If flag "v2_test" is disabled: returns v1 response format
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check feature flag
	if h.flagsClient.IsEnabled(ctx, "v2_test") {
		h.logger.Info("serving v2 response (flag enabled)")
		h.getUsersV2(w, r)
		return
	}

	h.logger.Info("serving v1 response (flag disabled)")
	h.getUsersV1(w, r)
}

// getUsersV1 returns the v1 response format.
func (h *Handler) getUsersV1(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"version": "v1",
		"data": []map[string]any{
			{
				"id":    "user-1",
				"name":  "Alice",
				"email": "alice@example.com",
			},
			{
				"id":    "user-2",
				"name":  "Bob",
				"email": "bob@example.com",
			},
		},
	}

	response.JSON(w, http.StatusOK, resp)
}

// getUsersV2 returns the v2 response format (enhanced).
func (h *Handler) getUsersV2(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"version": "v2",
		"meta": map[string]any{
			"total":  2,
			"limit":  10,
			"offset": 0,
		},
		"data": []map[string]any{
			{
				"id":         "user-1",
				"name":       "Alice",
				"email":      "alice@example.com",
				"created_at": "2024-01-15T10:30:00Z",
				"role":       "admin",
				"avatar_url": "https://example.com/avatars/alice.png",
			},
			{
				"id":         "user-2",
				"name":       "Bob",
				"email":      "bob@example.com",
				"created_at": "2024-02-20T14:45:00Z",
				"role":       "user",
				"avatar_url": "https://example.com/avatars/bob.png",
			},
		},
	}

	response.JSON(w, http.StatusOK, resp)
}

// GetVariant demonstrates getting a specific variant value.
//
// GET /demo/variant
func (h *Handler) GetVariant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	flagKey := r.URL.Query().Get("flag")
	if flagKey == "" {
		flagKey = "v2_test"
	}

	result := h.flagsClient.Evaluate(ctx, flagKey)

	resp := map[string]any{
		"flag":    flagKey,
		"enabled": result.Enabled,
		"variant": result.Variant,
		"reason":  result.Reason,
	}

	response.JSON(w, http.StatusOK, resp)
}

// CheckFlag demonstrates a simple flag check.
//
// GET /demo/check?flag=my_flag
func (h *Handler) CheckFlag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	flagKey := r.URL.Query().Get("flag")
	if flagKey == "" {
		response.BadRequest(w, "flag query parameter is required")
		return
	}

	enabled := h.flagsClient.IsEnabled(ctx, flagKey)

	resp := map[string]any{
		"flag":    flagKey,
		"enabled": enabled,
	}

	response.JSON(w, http.StatusOK, resp)
}
