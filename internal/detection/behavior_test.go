package detection

import (
	"testing"
	"time"
)

func newTestTracker() *BehaviorTracker {
	return NewBehaviorTracker(BehaviorTrackerConfig{
		Window:          5 * time.Minute,
		SweepThreshold:  5, // Lower for testing
		BruteThreshold:  3,
		ModuleThreshold: 3,
		CleanupInterval: 1 * time.Minute,
	})
}

func TestBehavior_InsufficientData(t *testing.T) {
	bt := newTestTracker()
	defer bt.Stop()

	bt.Record("1.2.3.4", RequestRecord{
		Timestamp: time.Now(),
		Path:      "/test",
		Module:    "wordpress",
	})

	result := bt.Analyze("1.2.3.4")
	if result != nil {
		t.Error("expected nil result for insufficient data")
	}
}

func TestBehavior_SweepDetection(t *testing.T) {
	bt := newTestTracker()
	defer bt.Stop()

	now := time.Now()
	paths := []string{"/a", "/b", "/c", "/d", "/e", "/f"}
	for _, p := range paths {
		bt.Record("1.2.3.4", RequestRecord{
			Timestamp: now,
			Path:      p,
			Module:    "detection",
		})
	}

	result := bt.Analyze("1.2.3.4")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.ScanSweep {
		t.Error("expected ScanSweep = true")
	}
	if result.Severity != "high" {
		t.Errorf("severity = %q, want high", result.Severity)
	}
}

func TestBehavior_BruteForceDetection(t *testing.T) {
	bt := newTestTracker()
	defer bt.Stop()

	now := time.Now()
	for i := 0; i < 5; i++ {
		bt.Record("1.2.3.4", RequestRecord{
			Timestamp: now,
			Path:      "/wp-login.php",
			Module:    "wordpress",
		})
	}

	result := bt.Analyze("1.2.3.4")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.BruteForce {
		t.Error("expected BruteForce = true")
	}
}

func TestBehavior_ModuleHopping(t *testing.T) {
	bt := newTestTracker()
	defer bt.Stop()

	now := time.Now()
	modules := []string{"wordpress", "exposure", "api"}
	for _, mod := range modules {
		bt.Record("1.2.3.4", RequestRecord{
			Timestamp: now,
			Path:      "/" + mod,
			Module:    mod,
		})
	}

	result := bt.Analyze("1.2.3.4")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.ModuleHopping {
		t.Error("expected ModuleHopping = true")
	}
	if result.Severity != "medium" {
		t.Errorf("severity = %q, want medium", result.Severity)
	}
}

func TestBehavior_ProgressiveEscalation(t *testing.T) {
	bt := newTestTracker()
	defer bt.Stop()

	now := time.Now()
	bt.Record("1.2.3.4", RequestRecord{
		Timestamp: now.Add(-3 * time.Minute),
		Path:      "/.env",
		Module:    "detection",
		Category:  string(CategoryRecon),
	})
	bt.Record("1.2.3.4", RequestRecord{
		Timestamp: now.Add(-2 * time.Minute),
		Path:      "/api/v1",
		Module:    "api",
		Category:  string(CategoryEnumeration),
	})
	bt.Record("1.2.3.4", RequestRecord{
		Timestamp: now,
		Path:      "/latest/meta-data",
		Module:    "cloud",
		Category:  string(CategoryExploit),
	})

	result := bt.Analyze("1.2.3.4")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.ProgressiveRecon {
		t.Error("expected ProgressiveRecon = true")
	}
	if result.Severity != "critical" {
		t.Errorf("severity = %q, want critical", result.Severity)
	}
}

func TestBehavior_SeverityComputation(t *testing.T) {
	tests := []struct {
		name    string
		result  *BehaviorResult
		wantSev string
	}{
		{"brute+module=critical", &BehaviorResult{BruteForce: true, ModuleHopping: true}, "critical"},
		{"progressive=critical", &BehaviorResult{ProgressiveRecon: true}, "critical"},
		{"sweep=high", &BehaviorResult{ScanSweep: true}, "high"},
		{"brute=high", &BehaviorResult{BruteForce: true}, "high"},
		{"module=medium", &BehaviorResult{ModuleHopping: true}, "medium"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeBehaviorSeverity(tt.result)
			if got != tt.wantSev {
				t.Errorf("severity = %q, want %q", got, tt.wantSev)
			}
		})
	}
}

func TestBehavior_RecordPruning(t *testing.T) {
	bt := NewBehaviorTracker(BehaviorTrackerConfig{
		Window:          100 * time.Millisecond,
		SweepThreshold:  5,
		BruteThreshold:  3,
		ModuleThreshold: 3,
		CleanupInterval: 1 * time.Minute,
	})
	defer bt.Stop()

	// Record old entries
	old := time.Now().Add(-200 * time.Millisecond)
	for i := 0; i < 10; i++ {
		bt.Record("1.2.3.4", RequestRecord{
			Timestamp: old,
			Path:      "/old",
			Module:    "test",
		})
	}

	// New record should trigger pruning
	bt.Record("1.2.3.4", RequestRecord{
		Timestamp: time.Now(),
		Path:      "/new",
		Module:    "test",
	})

	bt.mu.RLock()
	h := bt.histories["1.2.3.4"]
	count := len(h.records)
	bt.mu.RUnlock()

	if count != 1 {
		t.Errorf("records = %d after pruning, want 1", count)
	}
}

func TestBehavior_Cleanup(t *testing.T) {
	bt := NewBehaviorTracker(BehaviorTrackerConfig{
		Window:          50 * time.Millisecond,
		SweepThreshold:  5,
		BruteThreshold:  3,
		ModuleThreshold: 3,
		CleanupInterval: 50 * time.Millisecond,
	})
	defer bt.Stop()

	bt.Record("1.2.3.4", RequestRecord{
		Timestamp: time.Now(),
		Path:      "/test",
		Module:    "test",
	})

	// Wait for cleanup (stale = window * 2 = 100ms)
	time.Sleep(200 * time.Millisecond)

	bt.mu.RLock()
	count := len(bt.histories)
	bt.mu.RUnlock()

	if count != 0 {
		t.Errorf("histories = %d after cleanup, want 0", count)
	}
}

func TestBehavior_ConcurrentSafety(t *testing.T) {
	bt := newTestTracker()
	defer bt.Stop()

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				bt.Record("1.2.3.4", RequestRecord{
					Timestamp: time.Now(),
					Path:      "/concurrent",
					Module:    "test",
				})
				bt.Analyze("1.2.3.4")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
