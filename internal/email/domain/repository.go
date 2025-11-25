package domain

import (
	"context"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Repository defines the contract for email template persistence.
// This is a PORT in hexagonal architecture - the domain defines the interface,
// and infrastructure provides the implementation.
type Repository interface {
	// Save persists a template (insert or update).
	// Uses optimistic locking via version number.
	Save(ctx context.Context, template *Template) error

	// FindByID retrieves a template by ID.
	// Returns ErrTemplateNotFound if not found.
	FindByID(ctx context.Context, id types.ID) (*Template, error)

	// FindBySlug retrieves a template by slug, locale, and optional tenant.
	// If tenantID is nil, searches for system-wide templates.
	// If tenantID is provided, searches tenant-specific first, then falls back to system-wide.
	// Returns ErrTemplateNotFound if not found.
	FindBySlug(ctx context.Context, tenantID *string, slug string, locale Locale) (*Template, error)

	// FindActiveBySlug retrieves an active template by slug and locale.
	// This is the primary method used when sending emails.
	// Falls back from tenant-specific to system-wide templates.
	// Returns ErrTemplateNotFound if no active template found.
	FindActiveBySlug(ctx context.Context, tenantID *string, slug string, locale Locale) (*Template, error)

	// List retrieves templates matching the given filters.
	List(ctx context.Context, filters ListFilters) ([]*Template, error)

	// Count returns the total number of templates matching filters.
	Count(ctx context.Context, filters ListFilters) (int, error)

	// Delete removes a template by ID.
	// Returns ErrTemplateNotFound if not found.
	Delete(ctx context.Context, id types.ID) error

	// ExistsBySlug checks if a template with the given slug/locale exists.
	ExistsBySlug(ctx context.Context, tenantID *string, slug string, locale Locale) (bool, error)
}

// ListFilters defines query filters for listing templates.
type ListFilters struct {
	// TenantID filters by tenant. nil means system-wide only.
	// Use pointer to distinguish between "no filter" and "system-wide".
	TenantID *string

	// IncludeSystemTemplates includes system-wide templates in results.
	IncludeSystemTemplates bool

	// Status filters by template status.
	Status *Status

	// Locale filters by locale.
	Locale *Locale

	// SlugContains filters by slug (partial match, case-insensitive).
	SlugContains string

	// NameContains filters by name (partial match, case-insensitive).
	NameContains string

	// Pagination
	Limit  int
	Offset int

	// Sorting
	SortBy    SortField
	SortOrder SortOrder
}

// SortField defines fields that can be used for sorting.
type SortField string

const (
	SortByCreatedAt SortField = "created_at"
	SortByUpdatedAt SortField = "updated_at"
	SortByName      SortField = "name"
	SortBySlug      SortField = "slug"
)

// SortOrder defines sort direction.
type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

// DefaultListFilters returns filters with sensible defaults.
func DefaultListFilters() ListFilters {
	return ListFilters{
		IncludeSystemTemplates: true,
		Limit:                  20,
		Offset:                 0,
		SortBy:                 SortByCreatedAt,
		SortOrder:              SortOrderDesc,
	}
}
