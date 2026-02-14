# Background Correlation Pattern

## Problem

Cross-event analysis (like campaign detection) is expensive — it requires querying and comparing many events. Running it on every request would add unacceptable latency. But batch-only analysis (e.g., during intel export) means campaigns aren't detected until someone manually triggers an export.

## Solution: Background Ticker

Run a goroutine with `time.NewTicker` that periodically performs correlation. This provides near-real-time detection without per-request overhead.

```go
type CampaignCorrelator struct {
    cfg      CorrelatorConfig
    store    storage.Store
    detector *CampaignDetector
    mu       sync.Mutex
    known    map[string]string  // campaign name → ID
    stopCh   chan struct{}
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
```

### Lifecycle Pattern

This is the same pattern used by `BehaviorTracker` and `RateLimiter`:

1. **Constructor** starts the goroutine: `go cc.run()`
2. **Stop()** closes the stop channel: `close(cc.stopCh)`
3. **Server.Shutdown()** calls Stop on all background components

The `select` with `stopCh` ensures clean exit. No need for `sync.WaitGroup` since the goroutine exits immediately when the channel closes.

## Stable Identity Across Ticks

Campaign detection generates new UUIDs each time it runs. But the same logical campaign should keep the same ID across ticks (for upsert and stable event links).

**Solution:** Maintain a `known` map keyed by campaign name:

```go
if existingID, ok := cc.known[campaign.Name]; ok {
    campaign.ID = existingID  // Reuse
} else {
    cc.known[campaign.Name] = campaign.ID  // Remember
}
```

Campaign names are deterministic (derived from signature keys or module sets), so the same cluster always produces the same name.

Combined with `ON CONFLICT(id) DO UPDATE` in SQLite, this gives upsert semantics — new campaigns are inserted, existing ones are updated with new counts and IPs.

## Conditional SQL Updates

When multiple decorators update different fields of the same row, unconditional `SET field = ?` causes data loss.

**Problem:** ProfilingStore sets `attacker_id` with empty `campaign_id`. Later, the correlator sets `campaign_id` with empty `attacker_id`. Each overwrites the other's work.

**Solution:** Use `CASE WHEN` to preserve existing values when the new value is empty:

```sql
UPDATE events SET
    attacker_id = CASE WHEN ?1 != '' THEN ?1 ELSE attacker_id END,
    campaign_id = CASE WHEN ?2 != '' THEN ?2 ELSE campaign_id END
WHERE id = ?3
```

This is a common pattern when multiple writers update different columns of the same row at different times.

**Alternative considered:** Separate `UpdateAttackerID` and `UpdateCampaignID` methods. Rejected because it fragments the Store interface and the CASE WHEN approach is simpler.

## Sliding Window Queries

The correlator queries events within a time window:

```go
since := time.Now().Add(-cc.cfg.Window)
events, _ := cc.store.ListEvents(ctx, storage.EventFilter{
    Since: since,
})
```

**Trade-off:** Larger windows catch more campaigns but process more events per tick. Default: 30 minutes window, 60 second tick interval.

**Consideration:** Events at the window boundary may appear in multiple ticks. This is fine because campaign upsert is idempotent — the same events just re-confirm the same campaign.

## Testing Background Components

Ticker-based tests are flaky under race detector (timing-dependent). Two strategies:

### Strategy 1: Direct method call (preferred)

Create the struct without starting the goroutine, call the method directly:

```go
func newTestCorrelator(store storage.Store) *CampaignCorrelator {
    return &CampaignCorrelator{
        cfg:      CorrelatorConfig{Window: 30 * time.Minute, TickInterval: time.Hour},
        store:    store,
        detector: NewCampaignDetector(testLogger()),
        known:    make(map[string]string),
        stopCh:   make(chan struct{}),
    }
}

func TestCorrelator_SignatureCluster(t *testing.T) {
    store := &mockStore{events: [...]}
    cc := newTestCorrelator(store)
    cc.correlate()  // Direct call, deterministic
    // Assert results...
}
```

### Strategy 2: Short ticker + sleep (for lifecycle tests only)

```go
func TestCorrelator_Stop(t *testing.T) {
    cc := NewCampaignCorrelator(cfg, store, logger)  // Starts goroutine
    cc.Stop()                                         // Should not deadlock
    time.Sleep(100 * time.Millisecond)                // Confirm exit
}
```

Use Strategy 1 for logic tests (deterministic), Strategy 2 only for testing Start/Stop lifecycle.

## When to Use This Pattern

Background tickers are appropriate when:

- Analysis spans multiple entities (cross-event, cross-IP)
- Per-request computation would be too expensive
- Near-real-time (seconds) is acceptable vs true real-time
- The work is idempotent (re-running produces the same result)

**Not appropriate when:**

- Results are needed immediately per-request
- Analysis is cheap (just add it to middleware)
- State must be consistent with the latest request
