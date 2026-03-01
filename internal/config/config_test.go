package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Domain != "localhost" {
		t.Errorf("Server.Domain = %q, want localhost", cfg.Server.Domain)
	}
	if cfg.Server.TLS.Enabled {
		t.Error("TLS should be disabled by default")
	}
	if !cfg.RateLimit.Enabled {
		t.Error("RateLimit should be enabled by default")
	}
	if cfg.RateLimit.RequestsPerSecond != 10 {
		t.Errorf("RateLimit.RequestsPerSecond = %d, want 10", cfg.RateLimit.RequestsPerSecond)
	}
	if cfg.RateLimit.Burst != 20 {
		t.Errorf("RateLimit.Burst = %d, want 20", cfg.RateLimit.Burst)
	}
	if cfg.Storage.Type != "sqlite" {
		t.Errorf("Storage.Type = %q, want sqlite", cfg.Storage.Type)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level = %q, want info", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Logging.Format = %q, want json", cfg.Logging.Format)
	}
}

func TestDefault_ModulesDisabled(t *testing.T) {
	cfg := Default()

	if cfg.Modules.NextJS.Enabled {
		t.Error("NextJS should be disabled by default")
	}
	if cfg.Modules.WordPress.Enabled {
		t.Error("WordPress should be disabled by default")
	}
	if cfg.Modules.Exposure.Enabled {
		t.Error("Exposure should be disabled by default")
	}
	if cfg.Modules.API.Enabled {
		t.Error("API should be disabled by default")
	}
	if cfg.Modules.Admin.Enabled {
		t.Error("Admin should be disabled by default")
	}
	if cfg.Modules.Cloud.Enabled {
		t.Error("Cloud should be disabled by default")
	}
	if cfg.Modules.CVE.Enabled {
		t.Error("CVE should be disabled by default")
	}
}

func TestDefault_Fingerprinting(t *testing.T) {
	cfg := Default()

	if !cfg.Fingerprinting.JA3 {
		t.Error("JA3 should be enabled by default")
	}
	if !cfg.Fingerprinting.JA4 {
		t.Error("JA4 should be enabled by default")
	}
	if cfg.Fingerprinting.GeoIP {
		t.Error("GeoIP should be disabled by default")
	}
}

func TestValidate_DefaultConfig(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Errorf("Default config should be valid, got: %v", err)
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	cfg := Default()
	cfg.Server.Port = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for port 0")
	}

	cfg.Server.Port = 70000
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for port 70000")
	}
}

func TestValidate_TLSRequiresFiles(t *testing.T) {
	cfg := Default()
	cfg.Server.TLS.Enabled = true

	if err := cfg.Validate(); err == nil {
		t.Error("expected error when TLS enabled without cert/key")
	}

	cfg.Server.TLS.Cert = "/path/to/cert.pem"
	cfg.Server.TLS.Key = "/path/to/key.pem"
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid TLS config rejected: %v", err)
	}
}

func TestValidate_InvalidStorageType(t *testing.T) {
	cfg := Default()
	cfg.Storage.Type = "mongodb"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for unsupported storage type")
	}
}

func TestValidate_EmptyStoragePath(t *testing.T) {
	cfg := Default()
	cfg.Storage.Path = ""
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty storage path")
	}
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	cfg := Default()
	cfg.Logging.Level = "trace"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid log level")
	}
}

func TestValidate_InvalidLogFormat(t *testing.T) {
	cfg := Default()
	cfg.Logging.Format = "yaml"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid log format")
	}
}

func TestValidate_SlackAlertURL(t *testing.T) {
	cfg := Default()
	cfg.Alerts.Slack.WebhookURL = "not-a-url"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid Slack webhook URL")
	}

	cfg.Alerts.Slack.WebhookURL = "https://hooks.slack.com/services/T00/B00/xxx"
	cfg.Alerts.Slack.MinSeverity = "medium"
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid Slack config rejected: %v", err)
	}
}

func TestValidate_DiscordAlertURL(t *testing.T) {
	cfg := Default()
	cfg.Alerts.Discord.WebhookURL = "bad"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid Discord webhook URL")
	}
}

func TestValidate_WebhookAlertURL(t *testing.T) {
	cfg := Default()
	cfg.Alerts.Webhook.URL = "ftp://bad"
	cfg.Alerts.Webhook.MinSeverity = "invalid"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid webhook severity")
	}
}

func TestValidate_GeoIPRequiresDB(t *testing.T) {
	cfg := Default()
	cfg.Fingerprinting.GeoIP = true
	cfg.Fingerprinting.GeoIPDB = ""
	if err := cfg.Validate(); err == nil {
		t.Error("expected error when GeoIP enabled without DB path")
	}

	cfg.Fingerprinting.GeoIPDB = "/path/to/GeoLite2-City.mmdb"
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid GeoIP config rejected: %v", err)
	}
}

func TestEnabledModules_None(t *testing.T) {
	cfg := Default()
	modules := cfg.EnabledModules()
	if len(modules) != 0 {
		t.Errorf("EnabledModules() = %v, want empty", modules)
	}
}

func TestEnabledModules_Some(t *testing.T) {
	cfg := Default()
	cfg.Modules.WordPress.Enabled = true
	cfg.Modules.Cloud.Enabled = true
	cfg.Modules.CVE.Enabled = true

	modules := cfg.EnabledModules()
	if len(modules) != 3 {
		t.Fatalf("EnabledModules() = %v, want 3 modules", modules)
	}

	expected := map[string]bool{"wordpress": true, "cloud": true, "cve": true}
	for _, m := range modules {
		if !expected[m] {
			t.Errorf("unexpected module: %s", m)
		}
	}
}

func TestEnabledModules_All(t *testing.T) {
	cfg := Default()
	cfg.Modules.NextJS.Enabled = true
	cfg.Modules.WordPress.Enabled = true
	cfg.Modules.Exposure.Enabled = true
	cfg.Modules.API.Enabled = true
	cfg.Modules.Admin.Enabled = true
	cfg.Modules.Cloud.Enabled = true
	cfg.Modules.CVE.Enabled = true

	modules := cfg.EnabledModules()
	if len(modules) != 7 {
		t.Errorf("EnabledModules() returned %d, want 7", len(modules))
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "firewatch.yaml")
	data := []byte(`
server:
  port: 9090
  domain: honeypot.local
modules:
  wordpress:
    enabled: true
    fake_version: "6.5.0"
storage:
  type: sqlite
  path: ./test.db
logging:
  level: debug
  format: text
`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Server.Domain != "honeypot.local" {
		t.Errorf("Server.Domain = %q, want honeypot.local", cfg.Server.Domain)
	}
	if !cfg.Modules.WordPress.Enabled {
		t.Error("WordPress should be enabled")
	}
	if cfg.Modules.WordPress.FakeVersion != "6.5.0" {
		t.Errorf("FakeVersion = %q, want 6.5.0", cfg.Modules.WordPress.FakeVersion)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte("{{invalid"), 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoad_InvalidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.yaml")
	data := []byte(`
server:
  port: 0
storage:
  type: badtype
  path: ""
logging:
  level: invalid
  format: invalid
`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestLoadOrDefault_MissingFile(t *testing.T) {
	cfg, err := LoadOrDefault("/nonexistent/path.yaml")
	if err != nil {
		t.Fatalf("LoadOrDefault() error: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want default 8080", cfg.Server.Port)
	}
}

func TestLoadOrDefault_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "firewatch.yaml")
	data := []byte(`
server:
  port: 3000
storage:
  type: sqlite
  path: ./test.db
logging:
  level: warn
  format: json
`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	cfg, err := LoadOrDefault(path)
	if err != nil {
		t.Fatalf("LoadOrDefault() error: %v", err)
	}
	if cfg.Server.Port != 3000 {
		t.Errorf("Server.Port = %d, want 3000", cfg.Server.Port)
	}
}
