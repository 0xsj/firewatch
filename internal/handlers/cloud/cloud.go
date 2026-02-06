package cloud

import (
	"log/slog"

	"github.com/0xsj/firewatch/internal/config"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/internal/storage"
)

const moduleName = "cloud"

// Cloud is a honeypot module that emulates cloud provider metadata
// endpoints to detect SSRF and cloud credential theft attempts.
type Cloud struct {
	cfg    config.CloudModuleConfig
	store  storage.Store
	logger *slog.Logger
}

func New(cfg config.CloudModuleConfig, store storage.Store, logger *slog.Logger) *Cloud {
	return &Cloud{
		cfg:    cfg,
		store:  store,
		logger: logger.With("module", moduleName),
	}
}

func (c *Cloud) Name() string { return moduleName }

func (c *Cloud) Routes() []handlers.Route {
	return []handlers.Route{
		{Pattern: "GET /latest/meta-data/", Handler: c.handleMetadata},
		{Pattern: "GET /latest/meta-data/iam/", Handler: c.handleIAM},
		{Pattern: "GET /latest/meta-data/iam/security-credentials/", Handler: c.handleIAM},
		{Pattern: "GET /latest/user-data", Handler: c.handleMetadata},
		{Pattern: "GET /metadata/v1/", Handler: c.handleMetadata},
		{Pattern: "PUT /latest/api/token", Handler: c.handleIMDSv2},
	}
}
