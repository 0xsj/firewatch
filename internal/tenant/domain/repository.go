package tenant

import (
	"context"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Repository defines the persistence port for the Tenant aggregate.
type Repository interface {
	// Save persists a new tenant.
	Save(ctx context.Context, tenant *Tenant) error

	// Update persists changes to an existing tenant.
	Update(ctx context.Context, tenant *Tenant) error

	// FindByID retrieves a tenant by its ID.
	// Returns nil if not found.
	FindByID(ctx context.Context, id types.ID) (*Tenant, error)

	// FindBySlug retrieves a tenant by its slug.
	// Returns nil if not found.
	FindBySlug(ctx context.Context, slug Slug) (*Tenant, error)

	// SlugExists checks if a slug is already taken.
	SlugExists(ctx context.Context, slug Slug) (bool, error)

	// List retrieves tenants matching the given filters.
	List(ctx context.Context, filters ListFilters) ([]*Tenant, error)

	// Count returns the total number of tenants matching the filters.
	Count(ctx context.Context, filters ListFilters) (int64, error)

	// Delete removes a tenant from persistence.
	// For soft delete, use Update with status = deleted.
	Delete(ctx context.Context, id types.ID) error
}

// ListFilters defines filtering options for listing tenants.
type ListFilters struct {
	// Status filters by tenant status.
	Status *Status

	// Plan filters by tenant plan.
	Plan *Plan

	// OwnerID filters by owner user ID.
	OwnerID *types.ID

	// Search performs a text search on name and slug.
	Search *string

	// Pagination
	Offset int
	Limit  int
}

// DefaultListFilters returns filters with sensible defaults.
func DefaultListFilters() ListFilters {
	return ListFilters{
		Offset: 0,
		Limit:  20,
	}
}

// WithStatus adds a status filter.
func (f ListFilters) WithStatus(status Status) ListFilters {
	f.Status = &status
	return f
}

// WithPlan adds a plan filter.
func (f ListFilters) WithPlan(plan Plan) ListFilters {
	f.Plan = &plan
	return f
}

// WithOwnerID adds an owner ID filter.
func (f ListFilters) WithOwnerID(ownerID types.ID) ListFilters {
	f.OwnerID = &ownerID
	return f
}

// WithSearch adds a search filter.
func (f ListFilters) WithSearch(search string) ListFilters {
	f.Search = &search
	return f
}

// WithPagination sets offset and limit.
func (f ListFilters) WithPagination(offset, limit int) ListFilters {
	f.Offset = offset
	f.Limit = limit
	return f
}
