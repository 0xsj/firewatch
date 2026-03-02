package alerts

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Manager dispatches alerts to all registered alerters. It handles
// severity filtering, concurrent dispatch, deduplication, and error logging.
type Manager struct {
	alerters    []alerterEntry
	logger      *slog.Logger
	mu          sync.Mutex
	seen        map[string]time.Time // dedupKey → last sent timestamp
	dedupWindow time.Duration        // 0 = disabled
	stopCh      chan struct{}        // signals cleanup goroutine to exit
}

type alerterEntry struct {
	alerter     Alerter
	minSeverity string
}

// NewManager creates an alert manager. If dedupWindow > 0, duplicate
// alerts (same IP + module + lead signature) within the window are suppressed.
func NewManager(logger *slog.Logger, dedupWindow time.Duration) *Manager {
	m := &Manager{
		logger:      logger.With("component", "alerts"),
		dedupWindow: dedupWindow,
	}

	if dedupWindow > 0 {
		m.seen = make(map[string]time.Time)
		m.stopCh = make(chan struct{})
		go m.cleanupLoop()
	}

	return m
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
// Duplicate alerts within the dedup window are suppressed.
func (m *Manager) Send(ctx context.Context, alert Alert) {
	if len(m.alerters) == 0 {
		return
	}

	// Dedup check.
	if m.dedupWindow > 0 {
		key := dedupKey(alert)
		m.mu.Lock()
		if last, ok := m.seen[key]; ok && time.Since(last) < m.dedupWindow {
			m.mu.Unlock()
			m.logger.Debug("alert suppressed (dedup)",
				"key", key,
				"ip", alert.SourceIP,
				"module", alert.Module,
			)
			return
		}
		m.seen[key] = time.Now()
		m.mu.Unlock()
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

// Stop halts the dedup cleanup goroutine. Safe to call even if dedup
// is disabled (no-op).
func (m *Manager) Stop() {
	if m.stopCh != nil {
		close(m.stopCh)
	}
}

// dedupKey builds a deduplication key from alert fields:
// source IP + module + lead signature (or path as fallback).
func dedupKey(a Alert) string {
	sig := a.Path
	if len(a.Signatures) > 0 {
		sig = a.Signatures[0]
	}
	return a.SourceIP + "|" + a.Module + "|" + sig
}

// cleanupLoop periodically evicts stale entries from the seen map.
func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(m.dedupWindow)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for key, ts := range m.seen {
				if now.Sub(ts) >= m.dedupWindow {
					delete(m.seen, key)
				}
			}
			m.mu.Unlock()
		}
	}
}
