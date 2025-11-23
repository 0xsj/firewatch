package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
)

// ListUsersQuery handles retrieving a paginated list of users.
// Supports filtering, sorting, and pagination.
type ListUsersQuery struct {
	repo user.Repository
}

// NewListUsersQuery creates a new list users query.
func NewListUsersQuery(repo user.Repository) *ListUsersQuery {
	return &ListUsersQuery{
		repo: repo,
	}
}

// Handle executes the list users query.
func (q *ListUsersQuery) Handle(ctx context.Context, req dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
	const op = "ListUsersQuery.Handle"

	// 1. Validate request
	if err := validateListUsersRequest(req); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// 2. Map DTO request to domain filters
	filters := dto.MapFiltersFromRequest(req)

	// 3. Get users from repository
	// Repository automatically filters by tenant_id from context
	users, err := q.repo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to list users: %w", op, err)
	}

	// 4. Get total count (for pagination metadata)
	totalCount, err := q.repo.Count(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to count users: %w", op, err)
	}

	// 5. Map to response DTO
	return dto.MapListUsersResponse(users, totalCount, req), nil
}

// validateListUsersRequest validates the list users request.
func validateListUsersRequest(req dto.ListUsersRequest) error {
	// Validate limit
	if req.Limit < 1 {
		return fmt.Errorf("limit must be at least 1")
	}
	if req.Limit > 100 {
		return fmt.Errorf("limit cannot exceed 100")
	}

	// Validate offset
	if req.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}

	// Validate sort_by
	validSortFields := map[string]bool{
		"created_at":    true,
		"updated_at":    true,
		"email":         true,
		"last_login_at": true,
	}
	if req.SortBy != "" && !validSortFields[req.SortBy] {
		return fmt.Errorf("invalid sort_by field: %s", req.SortBy)
	}

	// Validate sort_order
	if req.SortOrder != "" && req.SortOrder != "asc" && req.SortOrder != "desc" {
		return fmt.Errorf("sort_order must be 'asc' or 'desc'")
	}

	return nil
}
