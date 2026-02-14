package storage

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/crypto"
	"github.com/0xsj/firewatch/pkg/errors"
)

// ProfilingStore wraps a Store and automatically creates/updates Attacker
// profiles whenever an event is saved.
type ProfilingStore struct {
	Store
	logger *slog.Logger
	mu     sync.Mutex // serializes per-IP updates
}

// NewProfilingStore wraps an existing store with attacker auto-profiling.
func NewProfilingStore(store Store, logger *slog.Logger) *ProfilingStore {
	return &ProfilingStore{
		Store:  store,
		logger: logger.With("component", "profiling"),
	}
}

// SaveEvent persists the event and asynchronously updates the attacker profile.
func (s *ProfilingStore) SaveEvent(ctx context.Context, event *models.Event) error {
	if err := s.Store.SaveEvent(ctx, event); err != nil {
		return err
	}

	go s.updateAttacker(event)

	return nil
}

func (s *ProfilingStore) updateAttacker(event *models.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()

	attacker, err := s.Store.GetAttackerByIP(ctx, event.SourceIP)
	if err != nil {
		if errors.GetKind(err) != errors.KindNotFound {
			s.logger.Error("failed to get attacker by IP",
				"error", err,
				"ip", event.SourceIP,
			)
			return
		}
		// Create new attacker
		attacker = &models.Attacker{
			ID:        crypto.UUID4(),
			FirstSeen: event.Timestamp,
			LastSeen:  event.Timestamp,
			IP:        event.SourceIP,
			Severity:  event.Severity,
		}
	}

	// Update fields
	attacker.LastSeen = event.Timestamp
	attacker.TotalEvents++

	if event.UserAgent != "" {
		attacker.UserAgents = uniqueAppend(attacker.UserAgents, event.UserAgent)
	}
	if event.Module != "" {
		attacker.ModulesTargeted = uniqueAppend(attacker.ModulesTargeted, event.Module)
	}
	if event.Path != "" {
		attacker.PathsProbed = boundedAppend(attacker.PathsProbed, event.Path, 100)
	}
	if event.Fingerprint.JA3Hash != "" {
		attacker.JA3Hashes = uniqueAppend(attacker.JA3Hashes, event.Fingerprint.JA3Hash)
	}

	// GeoIP enrichment
	if event.GeoIP != nil && attacker.GeoIP == nil {
		attacker.GeoIP = event.GeoIP
	}

	// Severity escalation
	attacker.Severity = higherSeverity(attacker.Severity, event.Severity)

	// Auto-tagging
	attacker.Tags = autoTag(attacker, event)

	// Save updated attacker
	if err := s.Store.SaveAttacker(ctx, attacker); err != nil {
		s.logger.Error("failed to save attacker profile",
			"error", err,
			"ip", event.SourceIP,
			"attacker_id", attacker.ID,
		)
		return
	}

	// Link event to attacker
	if err := s.Store.UpdateEventLinks(ctx, event.ID, attacker.ID, event.CampaignID); err != nil {
		s.logger.Error("failed to link event to attacker",
			"error", err,
			"event_id", event.ID,
			"attacker_id", attacker.ID,
		)
	}
}

// uniqueAppend appends a value to a slice only if not already present.
func uniqueAppend(slice []string, val string) []string {
	for _, s := range slice {
		if s == val {
			return slice
		}
	}
	return append(slice, val)
}

// boundedAppend appends a value to a slice, capping at maxLen (FIFO eviction).
func boundedAppend(slice []string, val string, maxLen int) []string {
	slice = append(slice, val)
	if len(slice) > maxLen {
		slice = slice[len(slice)-maxLen:]
	}
	return slice
}

var severityOrder = map[string]int{
	"info":     0,
	"low":      1,
	"medium":   2,
	"high":     3,
	"critical": 4,
}

// higherSeverity returns whichever severity is more severe.
func higherSeverity(a, b string) string {
	if severityOrder[b] > severityOrder[a] {
		return b
	}
	return a
}

// autoTag generates automatic tags based on event and attacker state.
func autoTag(a *models.Attacker, event *models.Event) []string {
	tags := make([]string, len(a.Tags))
	copy(tags, a.Tags)

	if event.Module == "rate_limit" {
		tags = uniqueAppend(tags, "rate-limited")
	}
	if event.Module == "ip_filter" {
		tags = uniqueAppend(tags, "blocklisted")
	}

	for _, sig := range event.Signatures {
		if strings.Contains(sig, "scanner") {
			tags = uniqueAppend(tags, "scanner")
		}
		if strings.Contains(sig, "brute") {
			tags = uniqueAppend(tags, "brute-forcer")
		}
	}

	if event.Severity == "critical" || event.Severity == "high" {
		tags = uniqueAppend(tags, "high-threat")
	}

	return tags
}
