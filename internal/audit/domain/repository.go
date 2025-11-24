package domain

import (
	"context"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Repository is the port for audit entry persistence.
type Repository interface {
	// Save persists an audit entry.
	Save(ctx context.Context, entry *AuditEntry) error

	// FindByID retrieves an audit entry by ID.
	FindByID(ctx context.Context, id types.ID) (*AuditEntry, error)

	// List retrieves audit entries matching the filters.
	List(ctx context.Context, filters Filters, page Page) (*PagedResult, error)

	// Count returns the total number of entries matching the filters.
	Count(ctx context.Context, filters Filters) (int, error)
}

// Filters for querying audit entries.
type Filters struct {
	TenantID      *string
	UserID        *string
	EventType     *string
	Source        *string
	CorrelationID *string
	FromTimestamp *types.Timestamp
	ToTimestamp   *types.Timestamp
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

// PagedResult contains a page of audit entries with total count.
type PagedResult struct {
	Entries []*AuditEntry
	Total   int
	Limit   int
	Offset  int
}

// HasMore returns true if there are more results.
func (r *PagedResult) HasMore() bool {
	return r.Offset+len(r.Entries) < r.Total
}
