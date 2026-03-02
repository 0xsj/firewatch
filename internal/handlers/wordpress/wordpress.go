package wordpress

import (
	"log/slog"

	"github.com/0xsj/firewatch/internal/config"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/internal/storage"
)

const moduleName = "wordpress"

// WordPress is a honeypot module that emulates a WordPress
// installation to detect brute force, XML-RPC abuse, and
// admin panel scanning.
type WordPress struct {
	cfg       config.WordPressModuleConfig
	deception config.DeceptionConfig
	store     storage.Store
	logger    *slog.Logger
}

// New creates a WordPress honeypot module.
func New(cfg config.WordPressModuleConfig, deception config.DeceptionConfig, store storage.Store, logger *slog.Logger) *WordPress {
	return &WordPress{
		cfg:       cfg,
		deception: deception,
		store:     store,
		logger:    logger.With("module", moduleName),
	}
}

func (wp *WordPress) Name() string { return moduleName }

func (wp *WordPress) Routes() []handlers.Route {
	return []handlers.Route{
		{Pattern: "GET /wp-login.php", Handler: wp.handleLoginGet},
		{Pattern: "POST /wp-login.php", Handler: wp.handleLoginPost},
		{Pattern: "GET /wp-admin/", Handler: wp.handleAdmin},
		{Pattern: "POST /xmlrpc.php", Handler: wp.handleXMLRPC},
		{Pattern: "GET /xmlrpc.php", Handler: wp.handleXMLRPC},
		{Pattern: "GET /wp-json/", Handler: wp.handleWPJSON},
		{Pattern: "GET /wp-includes/", Handler: wp.handleStatic},
		{Pattern: "GET /wp-content/", Handler: wp.handleStatic},
	}
}
