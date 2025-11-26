package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/tenant/application/dto"
	tenant "github.com/0xsj/hexagonal-go/internal/tenant/domain"
)

// GetTenantBySlugQuery handles fetching a tenant by slug.
type GetTenantBySlugQuery struct {
	repo tenant.Repository
}

// NewGetTenantBySlugQuery creates a new GetTenantBySlugQuery.
func NewGetTenantBySlugQuery(repo tenant.Repository) *GetTenantBySlugQuery {
	return &GetTenantBySlugQuery{
		repo: repo,
	}
}

// GetTenantBySlugRequest is the input for getting a tenant by slug.
type GetTenantBySlugRequest struct {
	Slug string
}

// Handle executes the get tenant by slug query.
func (q *GetTenantBySlugQuery) Handle(ctx context.Context, req GetTenantBySlugRequest) (*dto.TenantDTO, error) {
	const op = "GetTenantBySlugQuery.Handle"

	// Parse and validate slug
	slug, err := tenant.NewSlug(req.Slug)
	if err != nil {
		return nil, err
	}

	// Fetch tenant
	t, err := q.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to find tenant: %w", op, err)
	}
	if t == nil {
		return nil, tenant.ErrTenantNotFoundBySlug(op, req.Slug)
	}

	return dto.ToTenantDTO(t), nil
}
