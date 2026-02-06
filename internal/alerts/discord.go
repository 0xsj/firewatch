package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/0xsj/firewatch/pkg/errors"
)

// DiscordAlerter sends alerts to a Discord channel via webhook.
type DiscordAlerter struct {
	webhookURL string
	client     *http.Client
}

// NewDiscord creates a Discord alerter with the given webhook URL.
func NewDiscord(webhookURL string) *DiscordAlerter {
	return &DiscordAlerter{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}
}

func (d *DiscordAlerter) Name() string { return "discord" }

// Send formats and posts an alert to Discord using an embed.
func (d *DiscordAlerter) Send(ctx context.Context, alert Alert) error {
	payload := d.buildPayload(alert)

	body, err := json.Marshal(payload)
	if err != nil {
		return errors.E(errors.Op("alerts.discord.Send"), errors.KindInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.webhookURL, bytes.NewReader(body))
	if err != nil {
		return errors.E(errors.Op("alerts.discord.Send"), errors.KindInternal, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return errors.E(errors.Op("alerts.discord.Send"), errors.KindUnavailable, errors.CodeAlertSend, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.E(errors.Op("alerts.discord.Send"), errors.KindUnavailable, errors.CodeAlertSend,
			fmt.Sprintf("discord returned status %d", resp.StatusCode))
	}
	return nil
}

func (d *DiscordAlerter) buildPayload(alert Alert) map[string]any {
	fields := []map[string]any{
		{"name": "Module", "value": alert.Module, "inline": true},
		{"name": "Source IP", "value": alert.SourceIP, "inline": true},
		{"name": "Request", "value": fmt.Sprintf("`%s %s`", alert.Method, alert.Path), "inline": false},
	}

	if len(alert.Signatures) > 0 {
		fields = append(fields, map[string]any{
			"name":   "Signatures",
			"value":  strings.Join(alert.Signatures, ", "),
			"inline": false,
		})
	}

	return map[string]any{
		"embeds": []map[string]any{
			{
				"title":       fmt.Sprintf("%s — %s", strings.ToUpper(alert.Severity), alert.Title),
				"description": alert.Message,
				"color":       severityColor(alert.Severity),
				"fields":      fields,
				"footer":      map[string]any{"text": fmt.Sprintf("Request ID: %s", alert.RequestID)},
				"timestamp":   alert.Timestamp,
			},
		},
	}
}

func severityColor(severity string) int {
	switch severity {
	case "critical":
		return 0xED4245 // red
	case "high":
		return 0xE67E22 // orange
	case "medium":
		return 0xF1C40F // yellow
	case "low":
		return 0x3498DB // blue
	default:
		return 0x95A5A6 // gray
	}
}
