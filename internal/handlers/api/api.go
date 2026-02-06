package api

import (
	"log/slog"

	"github.com/0xsj/firewatch/internal/config"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/internal/storage"
)

const moduleName = "api"

// API is a honeypot module that emulates common API endpoints
// to detect enumeration, authentication probes, and API abuse.
type API struct {
	cfg    config.APIModuleConfig
	store  storage.Store
	logger *slog.Logger
}

func New(cfg config.APIModuleConfig, store storage.Store, logger *slog.Logger) *API {
	return &API{
		cfg:    cfg,
		store:  store,
		logger: logger.With("module", moduleName),
	}
}

func (a *API) Name() string { return moduleName }

func (a *API) Routes() []handlers.Route {
	return []handlers.Route{
		{Pattern: "GET /api/", Handler: a.handleREST},
		{Pattern: "POST /api/", Handler: a.handleREST},
		{Pattern: "GET /graphql", Handler: a.handleGraphQL},
		{Pattern: "POST /graphql", Handler: a.handleGraphQL},
		{Pattern: "GET /graphiql", Handler: a.handleGraphQL},
		{Pattern: "GET /swagger/", Handler: a.handleSwagger},
		{Pattern: "GET /swagger.json", Handler: a.handleSwagger},
		{Pattern: "GET /api-docs/", Handler: a.handleSwagger},
		{Pattern: "GET /openapi.json", Handler: a.handleSwagger},
	}
}
