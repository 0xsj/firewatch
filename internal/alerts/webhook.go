package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/0xsj/firewatch/pkg/errors"
)

// WebhookAlerter sends alerts as JSON POST requests to a generic
// webhook endpoint. Suitable for SIEM integration, custom dashboards,
// or any HTTP-based alert consumer.
type WebhookAlerter struct {
	url     string
	headers map[string]string
	client  *http.Client
}

// NewWebhook creates a webhook alerter. Custom headers are included
// in every request (useful for auth tokens, API keys).
func NewWebhook(url string, headers map[string]string) *WebhookAlerter {
	return &WebhookAlerter{
		url:     url,
		headers: headers,
		client:  &http.Client{},
	}
}

func (w *WebhookAlerter) Name() string { return "webhook" }

// Send posts the alert as JSON to the configured URL.
func (w *WebhookAlerter) Send(ctx context.Context, alert Alert) error {
	body, err := json.Marshal(alert)
	if err != nil {
		return errors.E(errors.Op("alerts.webhook.Send"), errors.KindInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(body))
	if err != nil {
		return errors.E(errors.Op("alerts.webhook.Send"), errors.KindInternal, err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range w.headers {
		req.Header.Set(k, v)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return errors.E(errors.Op("alerts.webhook.Send"), errors.KindUnavailable, errors.CodeAlertSend, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.E(errors.Op("alerts.webhook.Send"), errors.KindUnavailable, errors.CodeAlertSend,
			fmt.Sprintf("webhook returned status %d", resp.StatusCode))
	}
	return nil
}
