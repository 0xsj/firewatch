package detection

import (
	"sync"
	"time"
)

// RequestRecord captures the key attributes of a request for behavioral analysis.
type RequestRecord struct {
	Timestamp time.Time
	Path      string
	Module    string
	Category  string // detection category if matched
}

// BehaviorResult holds the behavioral analysis findings for an IP.
type BehaviorResult struct {
	ScanSweep        bool   // Many unique paths in window
	BruteForce       bool   // Same path hit repeatedly
	ModuleHopping    bool   // Multiple modules targeted
	ProgressiveRecon bool   // Category escalation (recon → exploit)
	Severity         string // Computed from findings
	Signatures       []string
}

// BehaviorTrackerConfig holds behavioral analysis parameters.
type BehaviorTrackerConfig struct {
	Window          time.Duration
	SweepThreshold  int
	BruteThreshold  int
	ModuleThreshold int
	CleanupInterval time.Duration
}

// BehaviorTracker tracks request sequences per-IP for temporal pattern detection.
type BehaviorTracker struct {
	cfg       BehaviorTrackerConfig
	mu        sync.RWMutex
	histories map[string]*ipHistory
	stopCh    chan struct{}
}

type ipHistory struct {
	records    []RequestRecord
	lastAccess time.Time
}

// NewBehaviorTracker creates a new tracker with background cleanup.
func NewBehaviorTracker(cfg BehaviorTrackerConfig) *BehaviorTracker {
	bt := &BehaviorTracker{
		cfg:       cfg,
		histories: make(map[string]*ipHistory),
		stopCh:    make(chan struct{}),
	}
	go bt.cleanup()
	return bt
}

// Stop terminates the cleanup goroutine.
func (bt *BehaviorTracker) Stop() {
	close(bt.stopCh)
}

// Record adds a request observation for the given IP.
func (bt *BehaviorTracker) Record(ip string, rec RequestRecord) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	h, exists := bt.histories[ip]
	if !exists {
		h = &ipHistory{}
		bt.histories[ip] = h
	}

	h.records = append(h.records, rec)
	h.lastAccess = time.Now()

	// Prune records outside the window
	cutoff := time.Now().Add(-bt.cfg.Window)
	pruned := h.records[:0]
	for _, r := range h.records {
		if r.Timestamp.After(cutoff) {
			pruned = append(pruned, r)
		}
	}
	h.records = pruned
}

// Analyze evaluates the behavioral patterns for an IP.
// Returns nil if insufficient data for any detection.
func (bt *BehaviorTracker) Analyze(ip string) *BehaviorResult {
	bt.mu.RLock()
	h, exists := bt.histories[ip]
	if !exists || len(h.records) < 3 {
		bt.mu.RUnlock()
		return nil
	}

	// Copy records under lock to avoid holding it during analysis
	records := make([]RequestRecord, len(h.records))
	copy(records, h.records)
	bt.mu.RUnlock()

	result := &BehaviorResult{}

	// Scan sweep: many unique paths
	uniquePaths := make(map[string]bool)
	for _, r := range records {
		uniquePaths[r.Path] = true
	}
	if len(uniquePaths) >= bt.cfg.SweepThreshold {
		result.ScanSweep = true
		result.Signatures = append(result.Signatures, "behavior-scan-sweep")
	}

	// Brute force: same path hit repeatedly
	pathCounts := make(map[string]int)
	for _, r := range records {
		pathCounts[r.Path]++
	}
	for _, count := range pathCounts {
		if count >= bt.cfg.BruteThreshold {
			result.BruteForce = true
			result.Signatures = append(result.Signatures, "behavior-brute-force")
			break
		}
	}

	// Module hopping: multiple modules targeted
	uniqueModules := make(map[string]bool)
	for _, r := range records {
		if r.Module != "" {
			uniqueModules[r.Module] = true
		}
	}
	if len(uniqueModules) >= bt.cfg.ModuleThreshold {
		result.ModuleHopping = true
		result.Signatures = append(result.Signatures, "behavior-module-hopping")
	}

	// Progressive recon: category escalation
	categories := make(map[string]time.Time) // category → first seen
	for _, r := range records {
		if r.Category != "" {
			if _, exists := categories[r.Category]; !exists {
				categories[r.Category] = r.Timestamp
			}
		}
	}
	if hasEscalation(categories) {
		result.ProgressiveRecon = true
		result.Signatures = append(result.Signatures, "behavior-progressive-recon")
	}

	if len(result.Signatures) == 0 {
		return nil
	}

	// Compute severity
	result.Severity = computeBehaviorSeverity(result)

	return result
}

// hasEscalation checks if categories escalated from recon to exploit.
func hasEscalation(categories map[string]time.Time) bool {
	reconTime, hasRecon := categories[string(CategoryRecon)]
	if !hasRecon {
		reconTime, hasRecon = categories[string(CategoryEnumeration)]
	}
	exploitTime, hasExploit := categories[string(CategoryExploit)]

	if hasRecon && hasExploit && exploitTime.After(reconTime) {
		return true
	}
	return false
}

func computeBehaviorSeverity(r *BehaviorResult) string {
	if r.ProgressiveRecon || (r.BruteForce && r.ModuleHopping) {
		return "critical"
	}
	if r.ScanSweep || r.BruteForce {
		return "high"
	}
	if r.ModuleHopping {
		return "medium"
	}
	return "medium"
}

func (bt *BehaviorTracker) cleanup() {
	ticker := time.NewTicker(bt.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bt.mu.Lock()
			now := time.Now()
			staleThreshold := bt.cfg.Window * 2

			for ip, h := range bt.histories {
				if now.Sub(h.lastAccess) > staleThreshold {
					delete(bt.histories, ip)
				}
			}
			bt.mu.Unlock()

		case <-bt.stopCh:
			return
		}
	}
}
