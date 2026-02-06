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

// SlackAlerter sends alerts to a Slack channel via incoming webhook.
type SlackAlerter struct {
	webhookURL string
	client     *http.Client
}

// NewSlack creates a Slack alerter with the given webhook URL.
func NewSlack(webhookURL string) *SlackAlerter {
	return &SlackAlerter{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}
}

func (s *SlackAlerter) Name() string { return "slack" }

// Send formats and posts an alert to Slack using Block Kit.
func (s *SlackAlerter) Send(ctx context.Context, alert Alert) error {
	payload := s.buildPayload(alert)

	body, err := json.Marshal(payload)
	if err != nil {
		return errors.E(errors.Op("alerts.slack.Send"), errors.KindInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(body))
	if err != nil {
		return errors.E(errors.Op("alerts.slack.Send"), errors.KindInternal, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return errors.E(errors.Op("alerts.slack.Send"), errors.KindUnavailable, errors.CodeAlertSend, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.E(errors.Op("alerts.slack.Send"), errors.KindUnavailable, errors.CodeAlertSend,
			fmt.Sprintf("slack returned status %d", resp.StatusCode))
	}
	return nil
}

func (s *SlackAlerter) buildPayload(alert Alert) map[string]any {
	emoji := severityEmoji(alert.Severity)
	title := fmt.Sprintf("%s %s — %s", emoji, strings.ToUpper(alert.Severity), alert.Title)

	fields := []map[string]any{
		{"type": "mrkdwn", "text": fmt.Sprintf("*Module:*\n%s", alert.Module)},
		{"type": "mrkdwn", "text": fmt.Sprintf("*Source IP:*\n%s", alert.SourceIP)},
		{"type": "mrkdwn", "text": fmt.Sprintf("*Method:*\n%s %s", alert.Method, alert.Path)},
	}

	if len(alert.Signatures) > 0 {
		fields = append(fields, map[string]any{
			"type": "mrkdwn",
			"text": fmt.Sprintf("*Signatures:*\n%s", strings.Join(alert.Signatures, ", ")),
		})
	}

	return map[string]any{
		"blocks": []map[string]any{
			{
				"type": "header",
				"text": map[string]any{"type": "plain_text", "text": title},
			},
			{
				"type":   "section",
				"fields": fields,
			},
			{
				"type": "context",
				"elements": []map[string]any{
					{"type": "mrkdwn", "text": fmt.Sprintf("Request ID: `%s` | %s", alert.RequestID, alert.Timestamp)},
				},
			},
		},
	}
}

func severityEmoji(severity string) string {
	switch severity {
	case "critical":
		return "\xf0\x9f\x94\xb4" // red circle
	case "high":
		return "\xf0\x9f\x9f\xa0" // orange circle
	case "medium":
		return "\xf0\x9f\x9f\xa1" // yellow circle
	case "low":
		return "\xf0\x9f\x94\xb5" // blue circle
	default:
		return "\xe2\xaa\xaa" // white circle
	}
}
