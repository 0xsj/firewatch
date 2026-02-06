package config

import (
	"github.com/0xsj/firewatch/pkg/errors"
	"github.com/0xsj/firewatch/pkg/validate"
)

const op = errors.Op("config.Validate")

// Validate checks the config for invalid or missing values.
func (c *Config) Validate() error {
	var errs []error

	// Server
	if err := validate.Port(c.Server.Port); err != nil {
		errs = append(errs, err)
	}
	if c.Server.TLS.Enabled {
		if err := validate.NonEmpty("tls.cert", c.Server.TLS.Cert); err != nil {
			errs = append(errs, err)
		}
		if err := validate.NonEmpty("tls.key", c.Server.TLS.Key); err != nil {
			errs = append(errs, err)
		}
	}

	// Storage
	if err := validate.OneOf("storage.type", c.Storage.Type, []string{"sqlite", "postgres"}); err != nil {
		errs = append(errs, err)
	}
	if err := validate.NonEmpty("storage.path", c.Storage.Path); err != nil {
		errs = append(errs, err)
	}

	// Logging
	if err := validate.OneOf("logging.level", c.Logging.Level, []string{"debug", "info", "warn", "error"}); err != nil {
		errs = append(errs, err)
	}
	if err := validate.OneOf("logging.format", c.Logging.Format, []string{"json", "text"}); err != nil {
		errs = append(errs, err)
	}

	// Alerts — validate URLs only if set
	if c.Alerts.Slack.WebhookURL != "" {
		if err := validate.URL(c.Alerts.Slack.WebhookURL); err != nil {
			errs = append(errs, err)
		}
		if err := validate.Severity(c.Alerts.Slack.MinSeverity); err != nil {
			errs = append(errs, err)
		}
	}
	if c.Alerts.Discord.WebhookURL != "" {
		if err := validate.URL(c.Alerts.Discord.WebhookURL); err != nil {
			errs = append(errs, err)
		}
		if err := validate.Severity(c.Alerts.Discord.MinSeverity); err != nil {
			errs = append(errs, err)
		}
	}
	if c.Alerts.Webhook.URL != "" {
		if err := validate.URL(c.Alerts.Webhook.URL); err != nil {
			errs = append(errs, err)
		}
		if err := validate.Severity(c.Alerts.Webhook.MinSeverity); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.E(op, errors.KindValidation, errors.CodeConfigInvalid, errors.Join(errs...))
	}
	return nil
}
