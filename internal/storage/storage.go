package storage

import (
	"context"
	"time"

	"github.com/0xsj/firewatch/internal/storage/models"
)

// Store is the storage interface for all Firewatch data.
// Implementations must be safe for concurrent use.
type Store interface {
	// Events
	SaveEvent(ctx context.Context, event *models.Event) error
	GetEvent(ctx context.Context, id string) (*models.Event, error)
	ListEvents(ctx context.Context, f EventFilter) ([]*models.Event, error)
	CountEvents(ctx context.Context, f EventFilter) (int64, error)

	// Event links
	UpdateEventLinks(ctx context.Context, eventID, attackerID, campaignID string) error

	// Attackers
	SaveAttacker(ctx context.Context, attacker *models.Attacker) error
	GetAttacker(ctx context.Context, id string) (*models.Attacker, error)
	GetAttackerByIP(ctx context.Context, ip string) (*models.Attacker, error)
	ListAttackers(ctx context.Context, f AttackerFilter) ([]*models.Attacker, error)

	// Campaigns
	SaveCampaign(ctx context.Context, campaign *models.Campaign) error
	GetCampaign(ctx context.Context, id string) (*models.Campaign, error)
	ListCampaigns(ctx context.Context, f CampaignFilter) ([]*models.Campaign, error)

	// IOCs
	SaveIOC(ctx context.Context, ioc *models.IOC) error
	ListIOCs(ctx context.Context, f IOCFilter) ([]*models.IOC, error)

	// Honey tokens
	SaveHoneyToken(ctx context.Context, token *models.HoneyToken) error
	GetHoneyTokenByValue(ctx context.Context, value string) (*models.HoneyToken, error)
	ListHoneyTokens(ctx context.Context, f HoneyTokenFilter) ([]*models.HoneyToken, error)

	// Lifecycle
	Close() error
}

// EventFilter controls which events are returned by ListEvents.
type EventFilter struct {
	Since    time.Time // Events after this time
	Until    time.Time // Events before this time
	SourceIP string    // Filter by source IP
	Module   string    // Filter by honeypot module
	Severity string    // Filter by minimum severity
	Limit    int       // Max results (0 = no limit)
	Offset   int       // Skip N results
}

// AttackerFilter controls which attackers are returned.
type AttackerFilter struct {
	Since    time.Time
	Until    time.Time
	Severity string
	Tag      string
	Limit    int
	Offset   int
}

// CampaignFilter controls which campaigns are returned.
type CampaignFilter struct {
	Since  time.Time
	Until  time.Time
	Active bool // Only campaigns seen recently
	Limit  int
	Offset int
}

// IOCFilter controls which IOCs are returned.
type IOCFilter struct {
	Since    time.Time
	Until    time.Time
	Type     models.IOCType
	Severity string
	Limit    int
	Offset   int
}

// HoneyTokenFilter controls which honey tokens are returned.
type HoneyTokenFilter struct {
	Since    time.Time
	Until    time.Time
	Kind     string // filter by token kind
	SourceIP string
	Module   string
	Limit    int
	Offset   int
}
