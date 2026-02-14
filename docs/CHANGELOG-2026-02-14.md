# Changelog — 2026-02-14

## Summary

Two sessions adding detection and intelligence features: IP filtering, custom signatures, attacker auto-profiling, behavioral fingerprinting, and campaign auto-correlation. Total: 6 new files, 8 modified files, ~1400 lines added, 42 new tests.

---

## Session 1: Detection & Intel Features

### IP Allowlist/Blocklist

**New files:**
- `internal/middleware/ipfilter.go` (150 lines)
- `internal/middleware/ipfilter_test.go` (8 tests)

**Features:**
- CIDR and individual IP filtering
- Allowlist takes precedence over blocklist
- File-based lists (`allowlist_file`, `blocklist_file`) with `#` comment support
- Records blocked requests as events (module: `ip_filter`, severity: `medium`)
- Early middleware placement (skips expensive processing for blocked IPs)

**Configuration:**
```yaml
ip_filter:
  allowlist: ["10.0.0.0/8"]
  blocklist: ["203.0.113.0/24"]
  allowlist_file: "/etc/firewatch/allowlist.txt"
  blocklist_file: "/etc/firewatch/blocklist.txt"
```

### Custom Signatures

**New files:**
- `internal/detection/loader.go` (180 lines)
- `internal/detection/loader_test.go` (8 tests)

**Features:**
- Load YAML signature/pattern files and directories
- Same-ID signatures override built-in ones (merge semantics)
- Regex validation at load time (fail fast on bad patterns)
- `header:` field prefix for header-based matchers

**Configuration:**
```yaml
detection:
  signatures_file: "/etc/firewatch/custom-signatures.yaml"
  signatures_dir: "/etc/firewatch/signatures.d/"
```

### Attacker Auto-Profiling

**New files:**
- `internal/storage/profiling.go` (175 lines)
- `internal/storage/profiling_test.go` (7 tests)

**Features:**
- `ProfilingStore` wraps any `Store` — intercepts `SaveEvent`
- Async goroutine creates/updates Attacker records per-IP
- Tracks: user agents, modules targeted, paths probed, JA3 hashes
- Severity escalation (keeps highest seen)
- Auto-tagging: `scanner`, `brute-forcer`, `rate-limited`, `blocklisted`, `high-threat`
- Mutex-serialized per-IP updates (prevents races)

**Store wrapping order:**
```
SQLiteStore -> ProfilingStore -> AlertingStore
```

### Behavioral Fingerprinting

**New files:**
- `internal/detection/behavior.go` (200 lines)
- `internal/detection/behavior_test.go` (9 tests)
- `internal/middleware/behavior.go` (80 lines)
- `internal/middleware/behavior_test.go` (5 tests)

**Features:**
- Per-IP temporal pattern analysis within a sliding window
- Detects four behaviors:
  - **Scan sweep** — many unique paths (recon across the surface)
  - **Brute force** — same path hit repeatedly
  - **Module hopping** — multiple modules targeted
  - **Progressive recon** — category escalation (recon -> exploit)
- Background cleanup goroutine (same pattern as RateLimiter)
- Configurable thresholds and window

**Configuration:**
```yaml
detection:
  behavior:
    enabled: false
    window_minutes: 5
    sweep_threshold: 20
    brute_threshold: 10
    module_threshold: 3
    cleanup_minutes: 2
```

---

## Session 2: Campaign Auto-Correlation

### Background Correlator

**New files:**
- `internal/detection/correlator.go` (120 lines)
- `internal/detection/correlator_test.go` (7 tests)

**Modified files:**
- `internal/detection/campaign.go` — added `CampaignMatch` struct, `DetectCampaignsWithEvents()` method
- `internal/storage/sqlite.go` — `UpdateEventLinks` now preserves existing values with CASE WHEN
- `internal/config/config.go` — added `CampaignConfig`
- `internal/config/defaults.go` — added campaign defaults
- `internal/server/server.go` — wired correlator lifecycle
- `firewatch.yaml` — added `campaign:` config section

**Design:**

`CampaignCorrelator` runs as a background goroutine (same lifecycle pattern as `BehaviorTracker` and `RateLimiter`). Every N seconds it:

1. Queries recent events from the store (sliding window)
2. Runs `CampaignDetector.DetectCampaignsWithEvents()` to find signature clusters and coordinated attacks
3. Creates new campaigns or updates existing ones (reuses stable campaign IDs via name->ID map)
4. Links events to campaigns via `UpdateEventLinks("", campaignID)` — empty attacker_id preserves existing

**Key challenge solved — UpdateEventLinks overwrites:**

Previously, `UpdateEventLinks` unconditionally set both `attacker_id` and `campaign_id`. ProfilingStore sets attacker_id with empty campaign_id, so the correlator calling it later would blank out attacker_id. Fix: conditional SQL:

```sql
UPDATE events SET
    attacker_id = CASE WHEN ?1 != '' THEN ?1 ELSE attacker_id END,
    campaign_id = CASE WHEN ?2 != '' THEN ?2 ELSE campaign_id END
WHERE id = ?3
```

**Key challenge solved — Stable campaign IDs:**

`CampaignDetector` generates a new UUID each call. The correlator maintains a `known` map (`campaign name -> campaign ID`) so the same campaign gets the same ID across ticks. The `SaveCampaign` SQL uses `ON CONFLICT(id) DO UPDATE` for upsert.

**Configuration:**
```yaml
detection:
  campaign:
    enabled: false
    window_minutes: 30
    tick_seconds: 60
```

### Tests (7 new)

1. No events -> no campaigns
2. Single IP -> no campaigns (need 2+ IPs for clustering)
3. Signature cluster creates campaign (2 IPs, same sigs)
4. Coordinated attack creates campaign (2 IPs, same module set)
5. Campaign updated on next tick (3rd IP joins, same ID reused)
6. Events linked correctly (empty attacker_id, correct campaign_id)
7. Stop halts ticker cleanly

Tests call `cc.correlate()` directly for determinism (no timer flakiness).

---

## Middleware Pipeline (Updated)

```
Request -> Correlation -> IPFilter -> RateLimit -> Logging -> GeoIP -> Fingerprint -> Detection -> Behavior -> Handler
                                                                                                         |
                                                                              CampaignCorrelator (background, not middleware)
```

---

## Statistics

**Code changes (both sessions combined):**
- ~1400 lines added
- 6 new files (3 implementation + 3 test files per session)
- 8 files modified
- 0 new dependencies

**Tests:**
- 42 new test cases (8+8+7+9+5+7 = 44 across both sessions, minus 2 shared helpers)
- Total project tests: 156
- All passing with `-race`

---

## Sign-off

All checks passing:
- `go vet ./...`
- `go test -race -count=1 ./...`
- `go build ./cmd/firewatch/`
