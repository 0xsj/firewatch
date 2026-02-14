package detection

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/0xsj/firewatch/internal/storage"
)

// CorrelatorConfig holds campaign correlation parameters.
type CorrelatorConfig struct {
	Window       time.Duration // How far back to look for events
	TickInterval time.Duration // How often to run correlation
}

// CampaignCorrelator runs periodic campaign detection in the background,
// linking events to discovered campaigns. Same lifecycle pattern as
// BehaviorTracker and RateLimiter.
type CampaignCorrelator struct {
	cfg      CorrelatorConfig
	store    storage.Store
	detector *CampaignDetector
	logger   *slog.Logger
	mu       sync.Mutex
	known    map[string]string // campaign name → campaign ID (stable tracking)
	stopCh   chan struct{}
}

// NewCampaignCorrelator creates and starts a background correlator.
func NewCampaignCorrelator(cfg CorrelatorConfig, store storage.Store, logger *slog.Logger) *CampaignCorrelator {
	cc := &CampaignCorrelator{
		cfg:      cfg,
		store:    store,
		detector: NewCampaignDetector(logger),
		logger:   logger.With("component", "campaign-correlator"),
		known:    make(map[string]string),
		stopCh:   make(chan struct{}),
	}
	go cc.run()
	return cc
}

// Stop terminates the background correlator goroutine.
func (cc *CampaignCorrelator) Stop() {
	close(cc.stopCh)
}

func (cc *CampaignCorrelator) run() {
	ticker := time.NewTicker(cc.cfg.TickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cc.stopCh:
			return
		case <-ticker.C:
			cc.correlate()
		}
	}
}

func (cc *CampaignCorrelator) correlate() {
	ctx := context.Background()

	since := time.Now().Add(-cc.cfg.Window)
	events, err := cc.store.ListEvents(ctx, storage.EventFilter{
		Since: since,
	})
	if err != nil {
		cc.logger.Error("failed to list events for correlation", "error", err)
		return
	}

	if len(events) == 0 {
		return
	}

	matches := cc.detector.DetectCampaignsWithEvents(events)
	if len(matches) == 0 {
		return
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()

	for _, match := range matches {
		campaign := match.Campaign

		// Reuse existing campaign ID if we've seen this campaign name before
		if existingID, ok := cc.known[campaign.Name]; ok {
			campaign.ID = existingID
		} else {
			cc.known[campaign.Name] = campaign.ID
		}

		if err := cc.store.SaveCampaign(ctx, campaign); err != nil {
			cc.logger.Error("failed to save campaign",
				"error", err,
				"campaign", campaign.Name,
			)
			continue
		}

		// Link events to their campaign (empty attacker_id preserves existing)
		for _, eventID := range match.EventIDs {
			if err := cc.store.UpdateEventLinks(ctx, eventID, "", campaign.ID); err != nil {
				cc.logger.Error("failed to link event to campaign",
					"error", err,
					"event_id", eventID,
					"campaign_id", campaign.ID,
				)
			}
		}

		cc.logger.Info("campaign correlated",
			"campaign_id", campaign.ID,
			"name", campaign.Name,
			"events", len(match.EventIDs),
			"ips", campaign.AttackerCount,
		)
	}
}
