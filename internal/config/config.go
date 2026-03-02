package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the root configuration for Firewatch.
type Config struct {
	Server         ServerConfig      `yaml:"server"`
	IPFilter       IPFilterConfig    `yaml:"ip_filter"`
	RateLimit      RateLimitConfig   `yaml:"rate_limit"`
	Modules        ModulesConfig     `yaml:"modules"`
	Fingerprinting FingerprintConfig `yaml:"fingerprinting"`
	Detection      DetectionConfig   `yaml:"detection"`
	Alerts         AlertsConfig      `yaml:"alerts"`
	Storage        StorageConfig     `yaml:"storage"`
	Deception      DeceptionConfig   `yaml:"deception"`
	Logging        LoggingConfig     `yaml:"logging"`
}

// IPFilterConfig controls IP allowlist/blocklist filtering.
type IPFilterConfig struct {
	Allowlist     []string `yaml:"allowlist"`
	AllowlistFile string   `yaml:"allowlist_file"`
	Blocklist     []string `yaml:"blocklist"`
	BlocklistFile string   `yaml:"blocklist_file"`
}

// DetectionConfig controls the detection engine.
type DetectionConfig struct {
	SignaturesFile string         `yaml:"signatures_file"`
	SignaturesDir  string         `yaml:"signatures_dir"`
	Behavior       BehaviorConfig `yaml:"behavior"`
	Campaign       CampaignConfig `yaml:"campaign"`
}

// CampaignConfig controls background campaign auto-correlation.
type CampaignConfig struct {
	Enabled       bool `yaml:"enabled"`
	WindowMinutes int  `yaml:"window_minutes"`
	TickSeconds   int  `yaml:"tick_seconds"`
}

// BehaviorConfig controls behavioral fingerprinting.
type BehaviorConfig struct {
	Enabled         bool `yaml:"enabled"`
	WindowMinutes   int  `yaml:"window_minutes"`
	SweepThreshold  int  `yaml:"sweep_threshold"`
	BruteThreshold  int  `yaml:"brute_threshold"`
	ModuleThreshold int  `yaml:"module_threshold"`
	CleanupMinutes  int  `yaml:"cleanup_minutes"`
}

// ServerConfig controls the HTTP/HTTPS server.
type ServerConfig struct {
	Domain string    `yaml:"domain"`
	Port   int       `yaml:"port"`
	TLS    TLSConfig `yaml:"tls"`
}

// TLSConfig controls TLS termination.
type TLSConfig struct {
	Enabled bool   `yaml:"enabled"`
	Cert    string `yaml:"cert"`
	Key     string `yaml:"key"`
}

// RateLimitConfig controls per-IP rate limiting.
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerSecond int  `yaml:"requests_per_second"`
	Burst             int  `yaml:"burst"`
	CleanupMinutes    int  `yaml:"cleanup_minutes"`
}

// ModulesConfig controls which honeypot modules are active.
type ModulesConfig struct {
	NextJS    NextJSModuleConfig    `yaml:"nextjs"`
	WordPress WordPressModuleConfig `yaml:"wordpress"`
	Exposure  ExposureModuleConfig  `yaml:"exposure"`
	API       APIModuleConfig       `yaml:"api"`
	Admin     AdminModuleConfig     `yaml:"admin"`
	Cloud     CloudModuleConfig     `yaml:"cloud"`
	CVE       CVEModuleConfig       `yaml:"cve"`
}

// NextJSModuleConfig for the Next.js honeypot.
type NextJSModuleConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Endpoints []string `yaml:"endpoints"`
}

// WordPressModuleConfig for the WordPress honeypot.
type WordPressModuleConfig struct {
	Enabled     bool   `yaml:"enabled"`
	FakeVersion string `yaml:"fake_version"`
}

// ExposureModuleConfig for the sensitive file honeypot.
type ExposureModuleConfig struct {
	Enabled bool   `yaml:"enabled"`
	FakeEnv string `yaml:"fake_env"`
}

// APIModuleConfig for the API enumeration honeypot.
type APIModuleConfig struct {
	Enabled bool `yaml:"enabled"`
}

// AdminModuleConfig for admin panel honeypots.
type AdminModuleConfig struct {
	Enabled bool `yaml:"enabled"`
}

// CloudModuleConfig for cloud metadata honeypots.
type CloudModuleConfig struct {
	Enabled bool `yaml:"enabled"`
}

// CVEModuleConfig for CVE-specific honeypots.
type CVEModuleConfig struct {
	Enabled bool     `yaml:"enabled"`
	CVEs    []string `yaml:"cves"`
}

// FingerprintConfig controls request fingerprinting.
type FingerprintConfig struct {
	JA3        bool   `yaml:"ja3"`
	JA4        bool   `yaml:"ja4"`
	GeoIP      bool   `yaml:"geoip"`
	GeoIPDB    string `yaml:"geoip_db"`
	ReverseDNS bool   `yaml:"reverse_dns"`
}

// AlertsConfig controls alert dispatch.
type AlertsConfig struct {
	Slack   SlackAlertConfig   `yaml:"slack"`
	Discord DiscordAlertConfig `yaml:"discord"`
	Webhook WebhookAlertConfig `yaml:"webhook"`
	Dedup   AlertDedupConfig   `yaml:"dedup"`
}

// AlertDedupConfig controls alert deduplication.
type AlertDedupConfig struct {
	Enabled bool   `yaml:"enabled"`
	Window  string `yaml:"window"` // e.g. "5m", "1h" — parsed via time.ParseDuration
}

// SlackAlertConfig for Slack webhook alerts.
type SlackAlertConfig struct {
	WebhookURL  string `yaml:"webhook_url"`
	MinSeverity string `yaml:"min_severity"`
}

// DiscordAlertConfig for Discord webhook alerts.
type DiscordAlertConfig struct {
	WebhookURL  string `yaml:"webhook_url"`
	MinSeverity string `yaml:"min_severity"`
}

// WebhookAlertConfig for generic webhook alerts.
type WebhookAlertConfig struct {
	URL         string            `yaml:"url"`
	Headers     map[string]string `yaml:"headers"`
	MinSeverity string            `yaml:"min_severity"`
}

// StorageConfig controls the backing database.
type StorageConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

// DeceptionConfig controls deception techniques.
type DeceptionConfig struct {
	HoneyTokens bool `yaml:"honey_tokens"`
	Breadcrumbs bool `yaml:"breadcrumbs"`
	FakeErrors  bool `yaml:"fake_errors"`
}

// LoggingConfig controls structured logging.
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Load reads a YAML config file and returns a Config with
// defaults applied for any unset fields.
func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return cfg, nil
}

// LoadOrDefault tries to load from path. If the file doesn't exist,
// it returns the default config. Any other error is returned.
func LoadOrDefault(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Default(), nil
	}
	return Load(path)
}

// EnabledModules returns the names of all enabled honeypot modules.
func (c *Config) EnabledModules() []string {
	var modules []string
	if c.Modules.NextJS.Enabled {
		modules = append(modules, "nextjs")
	}
	if c.Modules.WordPress.Enabled {
		modules = append(modules, "wordpress")
	}
	if c.Modules.Exposure.Enabled {
		modules = append(modules, "exposure")
	}
	if c.Modules.API.Enabled {
		modules = append(modules, "api")
	}
	if c.Modules.Admin.Enabled {
		modules = append(modules, "admin")
	}
	if c.Modules.Cloud.Enabled {
		modules = append(modules, "cloud")
	}
	if c.Modules.CVE.Enabled {
		modules = append(modules, "cve")
	}
	return modules
}
