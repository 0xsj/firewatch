package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/0xsj/firewatch/internal/alerts"
	"github.com/0xsj/firewatch/internal/storage/models"
)

// AlertingStore wraps a Store and dispatches alerts whenever an
// event is saved. This wires alerting into the event flow without
// modifying any handler code.
type AlertingStore struct {
	Store
	alertMgr *alerts.Manager
}

// NewAlertingStore wraps an existing store with alert dispatch.
func NewAlertingStore(store Store, alertMgr *alerts.Manager) *AlertingStore {
	return &AlertingStore{
		Store:    store,
		alertMgr: alertMgr,
	}
}

// SaveEvent persists the event and dispatches an alert if the
// alert manager has any registered alerters.
func (s *AlertingStore) SaveEvent(ctx context.Context, event *models.Event) error {
	if err := s.Store.SaveEvent(ctx, event); err != nil {
		return err
	}

	if s.alertMgr.Count() == 0 {
		return nil
	}

	alert := alerts.Alert{
		ID:         event.ID,
		Timestamp:  event.Timestamp,
		Severity:   event.Severity,
		Module:     event.Module,
		Title:      buildTitle(event),
		Message:    fmt.Sprintf("%s %s from %s", event.Method, event.Path, event.SourceIP),
		SourceIP:   event.SourceIP,
		Path:       event.Path,
		Method:     event.Method,
		UserAgent:  event.UserAgent,
		Signatures: event.Signatures,
		RequestID:  event.RequestID,
	}

	// Dispatch asynchronously so we don't block event recording.
	go s.alertMgr.Send(ctx, alert)

	return nil
}

// buildTitle creates a human-readable alert title from the event.
func buildTitle(event *models.Event) string {
	if len(event.Signatures) > 0 {
		return fmt.Sprintf("[%s] %s", event.Module, strings.Join(event.Signatures, ", "))
	}
	return fmt.Sprintf("[%s] %s %s", event.Module, event.Method, event.Path)
}
