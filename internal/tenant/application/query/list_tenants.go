package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/tenant/application/dto"
	tenant "github.com/0xsj/hexagonal-go/internal/tenant/domain"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// ListTenantsQuery handles listing tenants with filters.
type ListTenantsQuery struct {
	repo tenant.Repository
}

// NewListTenantsQuery creates a new ListTenantsQuery.
func NewListTenantsQuery(repo tenant.Repository) *ListTenantsQuery {
	return &ListTenantsQuery{
		repo: repo,
	}
}

// ListTenantsRequest is the input for listing tenants.
type ListTenantsRequest struct {
	Status  *string
	Plan    *string
	OwnerID *string
	Search  *string
	Offset  int
	Limit   int
}

// Handle executes the list tenants query.
func (q *ListTenantsQuery) Handle(ctx context.Context, req ListTenantsRequest) (*dto.TenantListDTO, error) {
	const op = "ListTenantsQuery.Handle"

	// Build filters
	filters := tenant.DefaultListFilters()

	// Apply status filter
	if req.Status != nil {
		status, err := tenant.ParseStatus(*req.Status)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid status: %w", op, err)
		}
		filters = filters.WithStatus(status)
	}

	// Apply plan filter
	if req.Plan != nil {
		plan, err := tenant.ParsePlan(*req.Plan)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid plan: %w", op, err)
		}
		filters = filters.WithPlan(plan)
	}

	// Apply owner filter
	if req.OwnerID != nil {
		ownerID, err := types.ParseID(*req.OwnerID)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid owner_id: %w", op, err)
		}
		filters = filters.WithOwnerID(ownerID)
	}

	// Apply search filter
	if req.Search != nil && *req.Search != "" {
		filters = filters.WithSearch(*req.Search)
	}

	// Apply pagination
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	filters = filters.WithPagination(req.Offset, limit)

	// Fetch tenants
	tenants, err := q.repo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to list tenants: %w", op, err)
	}

	// Get total count
	total, err := q.repo.Count(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to count tenants: %w", op, err)
	}

	return dto.ToTenantListDTO(tenants, total, filters.Offset, filters.Limit), nil
}
