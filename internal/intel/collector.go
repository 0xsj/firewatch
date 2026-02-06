package intel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/0xsj/firewatch/internal/detection"
	"github.com/0xsj/firewatch/internal/intel/enrichment"
	"github.com/0xsj/firewatch/internal/intel/ioc"
	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
)

// Collector orchestrates IOC extraction, enrichment, and campaign
// detection across stored events. It's the entry point for all
// threat intelligence operations.
type Collector struct {
	store     storage.Store
	extractor *ioc.Extractor
	enrichers []enrichment.Enricher
	campaigns *detection.CampaignDetector
	logger    *slog.Logger
}

// NewCollector creates a threat intelligence collector.
func NewCollector(
	store storage.Store,
	enrichers []enrichment.Enricher,
	logger *slog.Logger,
) *Collector {
	return &Collector{
		store:     store,
		extractor: ioc.NewExtractor(),
		enrichers: enrichers,
		campaigns: detection.NewCampaignDetector(logger),
		logger:    logger.With("component", "intel-collector"),
	}
}

// CollectResult holds the output of a collection run.
type CollectResult struct {
	IOCs      []*models.IOC
	Campaigns []*models.Campaign
}

// Collect runs IOC extraction and campaign detection on events
// matching the given filter. Extracted IOCs are enriched and persisted.
func (c *Collector) Collect(ctx context.Context, f storage.EventFilter) (*CollectResult, error) {
	events, err := c.store.ListEvents(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("listing events: %w", err)
	}

	if len(events) == 0 {
		return &CollectResult{}, nil
	}

	c.logger.Info("starting collection",
		"events", len(events),
	)

	// Extract IOCs
	iocs := c.extractor.FromEvents(events)

	// Enrich each IOC
	for _, indicator := range iocs {
		c.enrich(ctx, indicator)
	}

	// Persist IOCs
	for _, indicator := range iocs {
		if err := c.store.SaveIOC(ctx, indicator); err != nil {
			c.logger.Error("failed to save IOC",
				"error", err,
				"type", indicator.Type,
				"value", indicator.Value,
			)
		}
	}

	// Detect campaigns
	campaigns := c.campaigns.DetectCampaigns(events)

	// Persist campaigns
	for _, campaign := range campaigns {
		if err := c.store.SaveCampaign(ctx, campaign); err != nil {
			c.logger.Error("failed to save campaign",
				"error", err,
				"campaign", campaign.Name,
			)
		}
	}

	c.logger.Info("collection complete",
		"iocs", len(iocs),
		"campaigns", len(campaigns),
	)

	return &CollectResult{
		IOCs:      iocs,
		Campaigns: campaigns,
	}, nil
}

// enrich runs all enrichers on a single IOC, logging failures
// but continuing on error (graceful degradation).
func (c *Collector) enrich(ctx context.Context, indicator *models.IOC) {
	for _, e := range c.enrichers {
		if err := e.Enrich(ctx, indicator); err != nil {
			c.logger.Debug("enrichment failed",
				"enricher", e.Name(),
				"ioc", indicator.Value,
				"error", err,
			)
		}
	}
}
