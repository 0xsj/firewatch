package config

// Default returns a Config with sensible defaults.
// Modules are disabled by default — the user opts in.
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Domain: "localhost",
			Port:   8080,
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		Modules: ModulesConfig{
			NextJS: NextJSModuleConfig{
				Enabled: false,
				Endpoints: []string{
					"/",
					"/_next/server/pages",
					"/_rsc",
				},
			},
			WordPress: WordPressModuleConfig{
				Enabled:     false,
				FakeVersion: "6.4.2",
			},
			Exposure: ExposureModuleConfig{
				Enabled: false,
				FakeEnv: "DB_HOST=localhost\nDB_USER=root\nDB_PASS=changeme\nSECRET_KEY=fake_secret_key_do_not_use\n",
			},
			API:   APIModuleConfig{Enabled: false},
			Admin: AdminModuleConfig{Enabled: false},
			Cloud: CloudModuleConfig{Enabled: false},
			CVE:   CVEModuleConfig{Enabled: false},
		},
		Fingerprinting: FingerprintConfig{
			JA3:        true,
			JA4:        true,
			GeoIP:      false,
			ReverseDNS: false,
		},
		Alerts: AlertsConfig{
			Slack:   SlackAlertConfig{MinSeverity: "medium"},
			Discord: DiscordAlertConfig{MinSeverity: "medium"},
			Webhook: WebhookAlertConfig{MinSeverity: "medium"},
		},
		Storage: StorageConfig{
			Type: "sqlite",
			Path: "./firewatch.db",
		},
		Deception: DeceptionConfig{
			HoneyTokens: true,
			Breadcrumbs: true,
			FakeErrors:  true,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}
