package domain

import (
	"context"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Repository defines the interface for flag persistence operations.
// This is a PORT in hexagonal architecture.
type Repository interface {
	// Save persists a flag (create or update).
	Save(ctx context.Context, flag *Flag) error

	// FindByID retrieves a flag by its ID.
	FindByID(ctx context.Context, id types.ID) (*Flag, error)

	// FindByKey retrieves a flag by its key within a tenant.
	FindByKey(ctx context.Context, tenantID, key string) (*Flag, error)

	// FindAll retrieves all flags for a tenant with optional filters.
	FindAll(ctx context.Context, tenantID string, filters *Filters) ([]*Flag, error)

	// Delete removes a flag by ID.
	Delete(ctx context.Context, id types.ID) error

	// Exists checks if a flag key exists within a tenant.
	Exists(ctx context.Context, tenantID, key string) (bool, error)
}

// Filters contains optional filters for listing flags.
type Filters struct {
	// Enabled filters by enabled status.
	Enabled *bool

	// Keys filters by specific flag keys.
	Keys []string

	// Search performs text search on key, name, description.
	Search string

	// Pagination
	Limit  int
	Offset int
}

// NewFilters creates a new Filters with defaults.
func NewFilters() *Filters {
	return &Filters{
		Limit:  50,
		Offset: 0,
	}
}

// WithEnabled sets the enabled filter.
func (f *Filters) WithEnabled(enabled bool) *Filters {
	f.Enabled = &enabled
	return f
}

// WithKeys sets the keys filter.
func (f *Filters) WithKeys(keys []string) *Filters {
	f.Keys = keys
	return f
}

// WithSearch sets the search filter.
func (f *Filters) WithSearch(search string) *Filters {
	f.Search = search
	return f
}

// WithLimit sets the limit.
func (f *Filters) WithLimit(limit int) *Filters {
	f.Limit = limit
	return f
}

// WithOffset sets the offset.
func (f *Filters) WithOffset(offset int) *Filters {
	f.Offset = offset
	return f
}
