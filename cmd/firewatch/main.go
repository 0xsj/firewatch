package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/0xsj/firewatch/internal/alerts"
	"github.com/0xsj/firewatch/internal/config"
	"github.com/0xsj/firewatch/internal/detection"
	"github.com/0xsj/firewatch/internal/fingerprint"
	apimod "github.com/0xsj/firewatch/internal/handlers/api"
	cloudmod "github.com/0xsj/firewatch/internal/handlers/cloud"
	exposuremod "github.com/0xsj/firewatch/internal/handlers/exposure"
	nextjsmod "github.com/0xsj/firewatch/internal/handlers/nextjs"
	wpmod "github.com/0xsj/firewatch/internal/handlers/wordpress"
	"github.com/0xsj/firewatch/internal/server"
	"github.com/0xsj/firewatch/internal/storage"
)

var version = "dev"

func main() {
	configPath := flag.String("config", "firewatch.yaml", "path to config file")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("firewatch", version)
		return
	}

	cfg, err := config.LoadOrDefault(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.Logging)

	// Storage.
	sqliteStore, err := storage.NewSQLite(cfg.Storage.Path)
	if err != nil {
		logger.Error("failed to open storage", "error", err)
		os.Exit(1)
	}
	defer sqliteStore.Close()

	// Alerting — set up before modules so AlertingStore can wrap the store.
	alertMgr := alerts.NewManager(logger)
	if url := cfg.Alerts.Slack.WebhookURL; url != "" {
		alertMgr.Register(alerts.NewSlack(url), cfg.Alerts.Slack.MinSeverity)
	}
	if url := cfg.Alerts.Discord.WebhookURL; url != "" {
		alertMgr.Register(alerts.NewDiscord(url), cfg.Alerts.Discord.MinSeverity)
	}
	if url := cfg.Alerts.Webhook.URL; url != "" {
		alertMgr.Register(alerts.NewWebhook(url, cfg.Alerts.Webhook.Headers), cfg.Alerts.Webhook.MinSeverity)
	}

	// Wrap the store so every SaveEvent dispatches alerts automatically.
	var store storage.Store = sqliteStore
	if alertMgr.Count() > 0 {
		store = storage.NewAlertingStore(sqliteStore, alertMgr)
	}

	// Fingerprint engine — JA3 capture requires TLS.
	var ja3Store *fingerprint.JA3Store
	if cfg.Server.TLS.Enabled && cfg.Fingerprinting.JA3 {
		ja3Store = fingerprint.NewJA3Store()
	}
	fpEngine := fingerprint.NewEngine(ja3Store)

	// Detection engine.
	detector := detection.NewDefault(logger)

	// Server — includes correlation, logging, fingerprint, and detection middleware.
	srv := server.New(cfg, store, fpEngine, detector, logger)

	// Wire JA3 capture into TLS handshake.
	if ja3Store != nil {
		srv.HTTPServer().TLSConfig.GetConfigForClient = ja3Store.TLSConfigCallback()
	}

	// Register enabled honeypot modules.
	moduleCount := 0
	mountModule := func(name string, enabled bool, create func()) {
		if enabled {
			create()
			moduleCount++
			logger.Info("module loaded", "module", name)
		}
	}

	mountModule("nextjs", cfg.Modules.NextJS.Enabled, func() {
		mod := nextjsmod.New(cfg.Modules.NextJS, store, logger)
		for _, route := range mod.Routes() {
			srv.Router().HandleFunc(route.Pattern, route.Handler)
		}
	})
	mountModule("wordpress", cfg.Modules.WordPress.Enabled, func() {
		mod := wpmod.New(cfg.Modules.WordPress, store, logger)
		for _, route := range mod.Routes() {
			srv.Router().HandleFunc(route.Pattern, route.Handler)
		}
	})
	mountModule("exposure", cfg.Modules.Exposure.Enabled, func() {
		mod := exposuremod.New(cfg.Modules.Exposure, store, logger)
		for _, route := range mod.Routes() {
			srv.Router().HandleFunc(route.Pattern, route.Handler)
		}
	})
	mountModule("api", cfg.Modules.API.Enabled, func() {
		mod := apimod.New(cfg.Modules.API, store, logger)
		for _, route := range mod.Routes() {
			srv.Router().HandleFunc(route.Pattern, route.Handler)
		}
	})
	mountModule("cloud", cfg.Modules.Cloud.Enabled, func() {
		mod := cloudmod.New(cfg.Modules.Cloud, store, logger)
		for _, route := range mod.Routes() {
			srv.Router().HandleFunc(route.Pattern, route.Handler)
		}
	})

	logger.Info("firewatch starting",
		"addr", fmt.Sprintf(":%d", cfg.Server.Port),
		"tls", cfg.Server.TLS.Enabled,
		"modules", moduleCount,
		"alerts", alertMgr.Count(),
	)

	if err := srv.ListenAndShutdown(); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	logger.Info("firewatch stopped")
}

func setupLogger(cfg config.LoggingConfig) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
