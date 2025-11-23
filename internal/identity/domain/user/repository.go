package user

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Repository defines the contract for user persistence.
// This is a PORT in hexagonal architecture - the domain defines the interface,
// and infrastructure provides the implementation.
//
// Design principles:
//   - Repository methods use domain types (User, Email, Status, etc.)
//   - Repository is unaware of database implementation details
//   - Methods are expressed in domain language, not database language
//   - All methods are tenant-aware (filtered by tenant_id from context)
type Repository interface {
	// Save persists a user (insert or update).
	// Uses optimistic locking via version number.
	// Returns ErrVersionMismatch if version conflict occurs.
	Save(ctx context.Context, user *User) error

	// FindByID retrieves a user by ID.
	// Returns ErrUserNotFound if user doesn't exist.
	// Automatically filters by tenant_id from context.
	FindByID(ctx context.Context, id types.ID) (*User, error)

	// FindByEmail retrieves a user by email address.
	// Returns ErrUserNotFound if user doesn't exist.
	// Automatically filters by tenant_id from context.
	FindByEmail(ctx context.Context, email Email) (*User, error)

	// EmailExists checks if an email is already registered.
	// Returns true if email exists within the tenant.
	// Automatically filters by tenant_id from context.
	EmailExists(ctx context.Context, email Email) (bool, error)

	// List retrieves users matching the given filters.
	// Automatically filters by tenant_id from context.
	// Returns paginated results.
	List(ctx context.Context, filters Filters) ([]*User, error)

	// Count returns the total number of users matching filters.
	// Automatically filters by tenant_id from context.
	Count(ctx context.Context, filters Filters) (int, error)

	// Delete removes a user by ID (hard delete).
	// Use User.Delete() for soft delete instead.
	// Automatically filters by tenant_id from context.
	Delete(ctx context.Context, id types.ID) error
}

// Filters defines query filters for listing users.
type Filters struct {
	// Status filter (optional)
	Status *Status

	// Role filter (optional)
	Role *Role

	// EmailVerified filter (optional)
	EmailVerified *bool

	// Search by email (partial match, case-insensitive)
	EmailContains string

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
	SortByCreatedAt   SortField = "created_at"
	SortByUpdatedAt   SortField = "updated_at"
	SortByEmail       SortField = "email"
	SortByLastLoginAt SortField = "last_login_at"
)

// SortOrder defines sort direction.
type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

// DefaultFilters returns filters with sensible defaults.
func DefaultFilters() Filters {
	return Filters{
		Status:        nil, // All statuses
		Role:          nil, // All roles
		EmailVerified: nil, // All
		EmailContains: "",
		Limit:         50,
		Offset:        0,
		SortBy:        SortByCreatedAt,
		SortOrder:     SortOrderDesc,
	}
}

// WithStatus adds a status filter.
func (f Filters) WithStatus(status Status) Filters {
	f.Status = &status
	return f
}

// WithRole adds a role filter.
func (f Filters) WithRole(role Role) Filters {
	f.Role = &role
	return f
}

// WithEmailVerified filters by email verification status.
func (f Filters) WithEmailVerified(verified bool) Filters {
	f.EmailVerified = &verified
	return f
}

// WithEmailContains adds an email search filter.
func (f Filters) WithEmailContains(search string) Filters {
	f.EmailContains = search
	return f
}

// WithPagination sets limit and offset.
func (f Filters) WithPagination(limit, offset int) Filters {
	f.Limit = limit
	f.Offset = offset
	return f
}

// WithSort sets sorting field and order.
func (f Filters) WithSort(sortBy SortField, order SortOrder) Filters {
	f.SortBy = sortBy
	f.SortOrder = order
	return f
}

// Validate validates the filters.
func (f Filters) Validate() error {
	if f.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	if f.Limit > 1000 {
		return fmt.Errorf("limit cannot exceed 1000")
	}
	if f.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}

	// Validate status if set
	if f.Status != nil {
		if err := f.Status.Validate(); err != nil {
			return err
		}
	}

	// Validate role if set
	if f.Role != nil {
		if err := f.Role.Validate(); err != nil {
			return err
		}
	}

	return nil
}
