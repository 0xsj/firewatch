# Pipeline / Orchestrator Pattern

## The Pattern

A central orchestrator coordinates a sequence of processing stages, where each stage transforms or enriches data before passing it to the next.

```
Events → [Extract IOCs] → [Enrich] → [Detect Campaigns] → [Persist]
```

## Implementation in Firewatch

The `Collector` orchestrates the full intel pipeline:

```go
func (c *Collector) Collect(ctx context.Context, f storage.EventFilter) (*CollectResult, error) {
    // Stage 1: Fetch raw data
    events, err := c.store.ListEvents(ctx, f)

    // Stage 2: Extract IOCs
    iocs := c.extractor.FromEvents(events)

    // Stage 3: Enrich each IOC
    for _, indicator := range iocs {
        c.enrich(ctx, indicator)
    }

    // Stage 4: Persist IOCs
    for _, indicator := range iocs {
        c.store.SaveIOC(ctx, indicator)
    }

    // Stage 5: Detect campaigns
    campaigns := c.campaigns.DetectCampaigns(events)

    // Stage 6: Persist campaigns
    for _, campaign := range campaigns {
        c.store.SaveCampaign(ctx, campaign)
    }

    return &CollectResult{IOCs: iocs, Campaigns: campaigns}, nil
}
```

## Key Design Decisions

### 1. Graceful degradation

Enrichment failures don't stop the pipeline:

```go
func (c *Collector) enrich(ctx context.Context, indicator *models.IOC) {
    for _, e := range c.enrichers {
        if err := e.Enrich(ctx, indicator); err != nil {
            c.logger.Debug("enrichment failed", ...)
            // Continue — don't return the error
        }
    }
}
```

A DNS timeout shouldn't prevent an IOC from being recorded. Log it, move on.

### 2. In-place mutation

Enrichers modify IOCs in-place rather than returning new objects:

```go
func (d *DNS) Enrich(ctx context.Context, ioc *models.IOC) error {
    names, err := netutil.ReverseLookupContext(ctx, ioc.Value)
    if len(names) > 0 {
        ioc.Hostname = names[0]  // Mutate directly
    }
    return err
}
```

This avoids allocating new IOC objects at each enrichment stage. Each enricher only touches its own fields.

### 3. Type-gating in enrichers

Enrichers skip irrelevant IOC types early:

```go
func (d *DNS) Enrich(ctx context.Context, ioc *models.IOC) error {
    if ioc.Type != models.IOCTypeIP {
        return nil  // DNS only applies to IPs
    }
    // ...
}
```

The orchestrator doesn't need to know which enrichers apply to which IOC types — each enricher is self-filtering.

### 4. Result aggregation

The orchestrator collects results from all stages into a single return value:

```go
type CollectResult struct {
    IOCs      []*models.IOC
    Campaigns []*models.Campaign
}
```

Callers get everything they need from one call.

## Pipeline vs Chain

| Aspect       | Pipeline (Collector)           | Chain (Middleware)            |
|--------------|--------------------------------|------------------------------|
| Flow         | Linear, stage by stage         | Nested, wrapping             |
| Data         | Transformed between stages     | Passed through unchanged     |
| Control      | Orchestrator drives            | Each link calls next         |
| Error        | Graceful degradation           | Propagate up the chain       |

Middleware chains are for request processing. Pipelines are for data processing.
