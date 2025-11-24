package domain

import (
	"context"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Repository is the port for notification persistence.
type Repository interface {
	// Save persists a notification (insert or update).
	Save(ctx context.Context, notification *Notification) error

	// FindByID retrieves a notification by ID.
	FindByID(ctx context.Context, id types.ID) (*Notification, error)

	// FindPending retrieves pending notifications for retry processing.
	FindPending(ctx context.Context, limit int) ([]*Notification, error)

	// List retrieves notifications matching the filters.
	List(ctx context.Context, filters Filters, page Page) (*PagedResult, error)

	// Count returns the total number of notifications matching the filters.
	Count(ctx context.Context, filters Filters) (int, error)
}

// Filters for querying notifications.
type Filters struct {
	TenantID  *string
	UserID    *string
	Channel   *Channel
	Status    *Status
	EventType *string
	Recipient *string
}

// Page represents pagination parameters.
type Page struct {
	Limit  int
	Offset int
}

// DefaultPage returns default pagination.
func DefaultPage() Page {
	return Page{
		Limit:  50,
		Offset: 0,
	}
}

// PagedResult contains a page of notifications with total count.
type PagedResult struct {
	Notifications []*Notification
	Total         int
	Limit         int
	Offset        int
}

// HasMore returns true if there are more results.
func (r *PagedResult) HasMore() bool {
	return r.Offset+len(r.Notifications) < r.Total
}
