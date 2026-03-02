package config

// Default returns a Config with sensible defaults.
// Modules are disabled by default — the user opts in.
func Default() *Config {
	return &Config{
		IPFilter: IPFilterConfig{},
		Server: ServerConfig{
			Domain: "localhost",
			Port:   8080,
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		RateLimit: RateLimitConfig{
			Enabled:           true,
			RequestsPerSecond: 10,
			Burst:             20,
			CleanupMinutes:    5,
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
			GeoIPDB:    "",
			ReverseDNS: false,
		},
		Detection: DetectionConfig{
			Behavior: BehaviorConfig{
				Enabled:         false,
				WindowMinutes:   5,
				SweepThreshold:  20,
				BruteThreshold:  10,
				ModuleThreshold: 3,
				CleanupMinutes:  2,
			},
			Campaign: CampaignConfig{
				Enabled:       false,
				WindowMinutes: 30,
				TickSeconds:   60,
			},
		},
		Alerts: AlertsConfig{
			Slack:   SlackAlertConfig{MinSeverity: "medium"},
			Discord: DiscordAlertConfig{MinSeverity: "medium"},
			Webhook: WebhookAlertConfig{MinSeverity: "medium"},
			Dedup: AlertDedupConfig{
				Enabled: true,
				Window:  "5m",
			},
		},
		Storage: StorageConfig{
			Type: "sqlite",
			Path: "./firewatch.db",
		},
		QueryAPI: QueryAPIConfig{Enabled: false},
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
