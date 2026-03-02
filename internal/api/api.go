package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/0xsj/firewatch/internal/storage"
)

// Handler serves the Firewatch query API.
type Handler struct {
	store  storage.Store
	logger *slog.Logger
	mux    *http.ServeMux
}

// New creates a Handler with all routes registered.
func New(store storage.Store, logger *slog.Logger) *Handler {
	h := &Handler{
		store:  store,
		logger: logger,
		mux:    http.NewServeMux(),
	}

	h.mux.HandleFunc("GET /api/v1/health", h.handleHealth)
	h.mux.HandleFunc("GET /api/v1/events", h.handleEvents)
	h.mux.HandleFunc("GET /api/v1/attackers", h.handleAttackers)
	h.mux.HandleFunc("GET /api/v1/campaigns", h.handleCampaigns)
	h.mux.HandleFunc("GET /api/v1/iocs", h.handleIOCs)
	h.mux.HandleFunc("GET /api/v1/tokens", h.handleTokens)
	h.mux.HandleFunc("GET /api/v1/stats", h.handleStats)

	return h
}

// ServeHTTP delegates to the internal mux.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// parseSince parses a duration string (e.g., "1h", "24h", "7d") or
// an RFC3339 timestamp into a time.Time.
func parseSince(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	// Try duration-like strings with "d" suffix.
	if strings.HasSuffix(s, "d") {
		days := s[:len(s)-1]
		var n int
		if _, err := fmt.Sscanf(days, "%d", &n); err == nil {
			return time.Now().UTC().Add(-time.Duration(n) * 24 * time.Hour), nil
		}
	}

	// Try standard Go duration.
	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().UTC().Add(-d), nil
	}

	// Try RFC3339.
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// Try date-only.
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unrecognized time format %q (use 1h, 24h, 7d, or RFC3339)", s)
}

func parseIntParam(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return def
	}
	return n
}
