package alerts

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Manager dispatches alerts to all registered alerters. It handles
// severity filtering, concurrent dispatch, and error logging.
type Manager struct {
	alerters []alerterEntry
	logger   *slog.Logger
}

type alerterEntry struct {
	alerter     Alerter
	minSeverity string
}

// NewManager creates an alert manager.
func NewManager(logger *slog.Logger) *Manager {
	return &Manager{
		logger: logger.With("component", "alerts"),
	}
}

// Register adds an alerter with a minimum severity threshold.
// Alerts below the threshold are silently dropped for this alerter.
func (m *Manager) Register(a Alerter, minSeverity string) {
	m.alerters = append(m.alerters, alerterEntry{
		alerter:     a,
		minSeverity: minSeverity,
	})
	m.logger.Info("registered alerter",
		"name", a.Name(),
		"min_severity", minSeverity,
	)
}

// Send dispatches an alert to all registered alerters that meet the
// severity threshold. Sends are concurrent with a per-alerter timeout.
func (m *Manager) Send(ctx context.Context, alert Alert) {
	if len(m.alerters) == 0 {
		return
	}

	var wg sync.WaitGroup

	for _, entry := range m.alerters {
		if !MeetsSeverity(alert.Severity, entry.minSeverity) {
			continue
		}

		wg.Add(1)
		go func(e alerterEntry) {
			defer wg.Done()

			sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			if err := e.alerter.Send(sendCtx, alert); err != nil {
				m.logger.Error("alert send failed",
					"alerter", e.alerter.Name(),
					"alert_id", alert.ID,
					"error", err,
				)
			} else {
				m.logger.Debug("alert sent",
					"alerter", e.alerter.Name(),
					"alert_id", alert.ID,
					"severity", alert.Severity,
				)
			}
		}(entry)
	}

	wg.Wait()
}

// Count returns the number of registered alerters.
func (m *Manager) Count() int {
	return len(m.alerters)
}
