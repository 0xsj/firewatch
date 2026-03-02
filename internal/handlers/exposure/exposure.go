package exposure

import (
	"log/slog"

	"github.com/0xsj/firewatch/internal/config"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/internal/storage"
)

const moduleName = "exposure"

// Exposure is a honeypot module that emulates exposed sensitive
// files (.env, .git, config files) to detect scanners probing
// for leaked credentials and source code.
type Exposure struct {
	cfg       config.ExposureModuleConfig
	deception config.DeceptionConfig
	store     storage.Store
	logger    *slog.Logger
}

func New(cfg config.ExposureModuleConfig, deception config.DeceptionConfig, store storage.Store, logger *slog.Logger) *Exposure {
	return &Exposure{
		cfg:       cfg,
		deception: deception,
		store:     store,
		logger:    logger.With("module", moduleName),
	}
}

func (e *Exposure) Name() string { return moduleName }

func (e *Exposure) Routes() []handlers.Route {
	return []handlers.Route{
		{Pattern: "GET /.env", Handler: e.handleEnv},
		{Pattern: "GET /.env.local", Handler: e.handleEnv},
		{Pattern: "GET /.env.production", Handler: e.handleEnv},
		{Pattern: "GET /.env.backup", Handler: e.handleEnv},
		{Pattern: "GET /.git/", Handler: e.handleGit},
		{Pattern: "GET /.git/config", Handler: e.handleGit},
		{Pattern: "GET /.git/HEAD", Handler: e.handleGit},
		{Pattern: "GET /config.php", Handler: e.handleConfig},
		{Pattern: "GET /configuration.php", Handler: e.handleConfig},
		{Pattern: "GET /wp-config.php", Handler: e.handleConfig},
		{Pattern: "GET /web.config", Handler: e.handleConfig},
		{Pattern: "GET /.htaccess", Handler: e.handleConfig},
		{Pattern: "GET /.htpasswd", Handler: e.handleConfig},
		{Pattern: "GET /.DS_Store", Handler: e.handleConfig},
	}
}
