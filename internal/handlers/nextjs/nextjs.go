package nextjs

import (
	"log/slog"
	"net/http"

	"github.com/0xsj/firewatch/internal/config"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/internal/storage"
)

const moduleName = "nextjs"

// NextJS is a honeypot module that emulates a Next.js application
// to detect and fingerprint scanners probing for Next.js-specific
// vulnerabilities and endpoints.
type NextJS struct {
	cfg       config.NextJSModuleConfig
	deception config.DeceptionConfig
	store     storage.Store
	logger    *slog.Logger
}

// New creates a NextJS honeypot module.
func New(cfg config.NextJSModuleConfig, deception config.DeceptionConfig, store storage.Store, logger *slog.Logger) *NextJS {
	return &NextJS{
		cfg:       cfg,
		deception: deception,
		store:     store,
		logger:    logger.With("module", moduleName),
	}
}

// Name returns the module identifier.
func (n *NextJS) Name() string {
	return moduleName
}

// Routes returns the HTTP routes this module handles.
func (n *NextJS) Routes() []handlers.Route {
	return []handlers.Route{
		{Pattern: "POST /", Handler: n.handleServerAction},
		{Pattern: "GET /_next/static/", Handler: n.handleStatic},
		{Pattern: "GET /_next/data/", Handler: n.handleStatic},
		{Pattern: "GET /_next/image", Handler: n.handleStatic},
		{Pattern: "GET /_rsc", Handler: n.handleRSC},
		{Pattern: "GET /__nextjs_original-stack-frame", Handler: n.handleRSC},
		{Pattern: "GET /", Handler: n.handlePage},
	}
}

// handlePage serves a fake Next.js page for GET requests to
// configured endpoints.
func (n *NextJS) handlePage(w http.ResponseWriter, r *http.Request) {
	// Check for RSC-related headers that indicate a scanner
	if r.Header.Get("Rsc") != "" || r.Header.Get("Next-Router-State-Tree") != "" {
		n.handleRSC(w, r)
		return
	}

	n.recordEvent(r, "info", nil)
	servePage(w)
}
