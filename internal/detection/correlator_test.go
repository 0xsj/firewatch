package detection

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/crypto"
	"github.com/0xsj/firewatch/pkg/timeutil"
)

// correlatorMockStore records calls for verification.
type correlatorMockStore struct {
	mu         sync.Mutex
	events     []*models.Event
	campaigns  []*models.Campaign
	eventLinks []eventLink
}

type eventLink struct {
	EventID    string
	AttackerID string
	CampaignID string
}

func (m *correlatorMockStore) SaveEvent(_ context.Context, e *models.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, e)
	return nil
}

func (m *correlatorMockStore) GetEvent(_ context.Context, _ string) (*models.Event, error) {
	return nil, nil
}

func (m *correlatorMockStore) ListEvents(_ context.Context, f storage.EventFilter) ([]*models.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*models.Event
	for _, e := range m.events {
		if !f.Since.IsZero() && e.Timestamp < f.Since.UTC().Format(time.RFC3339) {
			continue
		}
		result = append(result, e)
	}
	return result, nil
}

func (m *correlatorMockStore) CountEvents(_ context.Context, _ storage.EventFilter) (int64, error) {
	return int64(len(m.events)), nil
}

func (m *correlatorMockStore) UpdateEventLinks(_ context.Context, eventID, attackerID, campaignID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventLinks = append(m.eventLinks, eventLink{eventID, attackerID, campaignID})
	return nil
}

func (m *correlatorMockStore) SaveAttacker(_ context.Context, _ *models.Attacker) error { return nil }
func (m *correlatorMockStore) GetAttacker(_ context.Context, _ string) (*models.Attacker, error) {
	return nil, nil
}
func (m *correlatorMockStore) GetAttackerByIP(_ context.Context, _ string) (*models.Attacker, error) {
	return nil, nil
}
func (m *correlatorMockStore) ListAttackers(_ context.Context, _ storage.AttackerFilter) ([]*models.Attacker, error) {
	return nil, nil
}

func (m *correlatorMockStore) SaveCampaign(_ context.Context, c *models.Campaign) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.campaigns = append(m.campaigns, c)
	return nil
}

func (m *correlatorMockStore) GetCampaign(_ context.Context, _ string) (*models.Campaign, error) {
	return nil, nil
}
func (m *correlatorMockStore) ListCampaigns(_ context.Context, _ storage.CampaignFilter) ([]*models.Campaign, error) {
	return nil, nil
}

func (m *correlatorMockStore) SaveIOC(_ context.Context, _ *models.IOC) error { return nil }
func (m *correlatorMockStore) ListIOCs(_ context.Context, _ storage.IOCFilter) ([]*models.IOC, error) {
	return nil, nil
}

func (m *correlatorMockStore) SaveHoneyToken(context.Context, *models.HoneyToken) error { return nil }
func (m *correlatorMockStore) GetHoneyTokenByValue(context.Context, string) (*models.HoneyToken, error) {
	return nil, nil
}
func (m *correlatorMockStore) ListHoneyTokens(context.Context, storage.HoneyTokenFilter) ([]*models.HoneyToken, error) {
	return nil, nil
}

func (m *correlatorMockStore) Close() error { return nil }

func correlatorTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func makeEvent(id, ip, module, path, severity string, sigs []string) *models.Event {
	return &models.Event{
		ID:         id,
		Timestamp:  timeutil.FormatRFC3339(timeutil.NowUTC()),
		RequestID:  crypto.UUID4(),
		SourceIP:   ip,
		SourcePort: 12345,
		Module:     module,
		Method:     "GET",
		Path:       path,
		Severity:   severity,
		Signatures: sigs,
	}
}

// newTestCorrelator creates a correlator without starting the background goroutine.
func newTestCorrelator(store storage.Store) *CampaignCorrelator {
	return &CampaignCorrelator{
		cfg: CorrelatorConfig{
			Window:       30 * time.Minute,
			TickInterval: time.Hour, // never fires in tests
		},
		store:    store,
		detector: NewCampaignDetector(correlatorTestLogger()),
		logger:   correlatorTestLogger().With("component", "campaign-correlator"),
		known:    make(map[string]string),
		stopCh:   make(chan struct{}),
	}
}

func TestCorrelator_NoEvents(t *testing.T) {
	store := &correlatorMockStore{}
	cc := newTestCorrelator(store)

	cc.correlate()

	if len(store.campaigns) != 0 {
		t.Errorf("expected no campaigns, got %d", len(store.campaigns))
	}
}

func TestCorrelator_SingleIP_NoCampaign(t *testing.T) {
	store := &correlatorMockStore{
		events: []*models.Event{
			makeEvent("e1", "10.0.0.1", "wordpress", "/wp-login.php", "medium", []string{"wp-login-001"}),
			makeEvent("e2", "10.0.0.1", "wordpress", "/xmlrpc.php", "medium", []string{"wp-login-001"}),
		},
	}
	cc := newTestCorrelator(store)

	cc.correlate()

	if len(store.campaigns) != 0 {
		t.Errorf("expected no campaigns from single IP, got %d", len(store.campaigns))
	}
}

func TestCorrelator_SignatureCluster(t *testing.T) {
	store := &correlatorMockStore{
		events: []*models.Event{
			makeEvent("e1", "10.0.0.1", "wordpress", "/wp-login.php", "medium", []string{"wp-probe-001", "wp-probe-002"}),
			makeEvent("e2", "10.0.0.2", "wordpress", "/wp-login.php", "high", []string{"wp-probe-001", "wp-probe-002"}),
		},
	}
	cc := newTestCorrelator(store)

	cc.correlate()

	if len(store.campaigns) == 0 {
		t.Fatal("expected at least 1 campaign from signature cluster")
	}

	found := false
	for _, c := range store.campaigns {
		if c.AttackerCount >= 2 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected campaign with 2+ attacker IPs")
	}
}

func TestCorrelator_CoordinatedAttack(t *testing.T) {
	store := &correlatorMockStore{
		events: []*models.Event{
			makeEvent("e1", "10.0.0.1", "wordpress", "/wp-login.php", "medium", nil),
			makeEvent("e2", "10.0.0.1", "nextjs", "/_next/data", "low", nil),
			makeEvent("e3", "10.0.0.2", "wordpress", "/xmlrpc.php", "medium", nil),
			makeEvent("e4", "10.0.0.2", "nextjs", "/_rsc", "low", nil),
		},
	}
	cc := newTestCorrelator(store)

	cc.correlate()

	if len(store.campaigns) == 0 {
		t.Fatal("expected at least 1 campaign from coordinated attack")
	}

	found := false
	for _, c := range store.campaigns {
		if c.AttackerCount >= 2 && len(c.ModulesTargeted) >= 2 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected coordinated campaign with 2+ IPs and 2+ modules")
	}
}

func TestCorrelator_CampaignUpdatedOnNextTick(t *testing.T) {
	store := &correlatorMockStore{
		events: []*models.Event{
			makeEvent("e1", "10.0.0.1", "wordpress", "/wp-login.php", "medium", []string{"wp-probe-001"}),
			makeEvent("e2", "10.0.0.2", "wordpress", "/wp-login.php", "medium", []string{"wp-probe-001"}),
		},
	}
	cc := newTestCorrelator(store)

	// First correlate
	cc.correlate()

	if len(store.campaigns) == 0 {
		t.Fatal("expected campaign after first correlate")
	}
	firstID := store.campaigns[0].ID

	// Add a third IP with the same signature
	store.events = append(store.events,
		makeEvent("e3", "10.0.0.3", "wordpress", "/wp-login.php", "high", []string{"wp-probe-001"}),
	)

	// Second correlate
	cc.correlate()

	if len(store.campaigns) < 2 {
		t.Fatal("expected at least 2 SaveCampaign calls (initial + update)")
	}
	lastCampaign := store.campaigns[len(store.campaigns)-1]
	if lastCampaign.ID != firstID {
		t.Errorf("campaign ID changed: got %q, want %q (should be stable)", lastCampaign.ID, firstID)
	}
	if lastCampaign.AttackerCount < 3 {
		t.Errorf("expected 3+ attacker IPs after update, got %d", lastCampaign.AttackerCount)
	}
}

func TestCorrelator_EventsLinkedCorrectly(t *testing.T) {
	store := &correlatorMockStore{
		events: []*models.Event{
			makeEvent("e1", "10.0.0.1", "wordpress", "/wp-login.php", "medium", []string{"wp-probe-001"}),
			makeEvent("e2", "10.0.0.2", "wordpress", "/wp-login.php", "medium", []string{"wp-probe-001"}),
		},
	}
	cc := newTestCorrelator(store)

	cc.correlate()

	if len(store.eventLinks) < 2 {
		t.Fatalf("expected 2+ event links, got %d", len(store.eventLinks))
	}

	for _, link := range store.eventLinks {
		if link.AttackerID != "" {
			t.Errorf("expected empty attacker_id, got %q", link.AttackerID)
		}
		if link.CampaignID == "" {
			t.Error("expected non-empty campaign_id")
		}
	}

	linkedIDs := make(map[string]bool)
	for _, link := range store.eventLinks {
		linkedIDs[link.EventID] = true
	}
	if !linkedIDs["e1"] || !linkedIDs["e2"] {
		t.Errorf("expected events e1 and e2 to be linked, got %v", linkedIDs)
	}
}

func TestCorrelator_Stop(t *testing.T) {
	store := &correlatorMockStore{}
	cfg := CorrelatorConfig{
		Window:       30 * time.Minute,
		TickInterval: 50 * time.Millisecond,
	}
	cc := NewCampaignCorrelator(cfg, store, correlatorTestLogger())

	// Stop should return without deadlocking
	cc.Stop()

	// Give some time to ensure the goroutine has exited
	time.Sleep(100 * time.Millisecond)
}
