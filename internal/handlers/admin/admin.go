package admin

import (
	"log/slog"

	"github.com/0xsj/firewatch/internal/config"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/internal/storage"
)

const moduleName = "admin"

// Admin is a honeypot module that emulates common admin panels
// (phpMyAdmin, Adminer, cPanel) to detect scanning and brute force.
type Admin struct {
	cfg    config.AdminModuleConfig
	store  storage.Store
	logger *slog.Logger
}

// New creates an Admin honeypot module.
func New(cfg config.AdminModuleConfig, store storage.Store, logger *slog.Logger) *Admin {
	return &Admin{
		cfg:    cfg,
		store:  store,
		logger: logger.With("module", moduleName),
	}
}

func (a *Admin) Name() string { return moduleName }

func (a *Admin) Routes() []handlers.Route {
	return []handlers.Route{
		{Pattern: "GET /phpmyadmin/", Handler: a.handlePhpMyAdminGet},
		{Pattern: "POST /phpmyadmin/", Handler: a.handlePhpMyAdminPost},
		{Pattern: "GET /pma/", Handler: a.handlePhpMyAdminGet},
		{Pattern: "GET /phpMyAdmin/", Handler: a.handlePhpMyAdminGet},
		{Pattern: "GET /adminer.php", Handler: a.handleAdminerGet},
		{Pattern: "POST /adminer.php", Handler: a.handleAdminerPost},
		{Pattern: "GET /adminer/", Handler: a.handleAdminerGet},
		{Pattern: "GET /cpanel", Handler: a.handleCPanelGet},
		{Pattern: "POST /cpanel", Handler: a.handleCPanelPost},
		{Pattern: "GET /cpanel/", Handler: a.handleCPanelGet},
		{Pattern: "GET /admin", Handler: a.handleGenericGet},
		{Pattern: "GET /admin/", Handler: a.handleGenericGet},
		{Pattern: "POST /admin/login", Handler: a.handleGenericPost},
		{Pattern: "GET /administrator/", Handler: a.handleGenericGet},
		{Pattern: "GET /manager/", Handler: a.handleGenericGet},
		{Pattern: "GET /admin/login", Handler: a.handleGenericGet},
	}
}
