package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/tenant/application/dto"
	tenant "github.com/0xsj/hexagonal-go/internal/tenant/domain"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// GetTenantQuery handles fetching a tenant by ID.
type GetTenantQuery struct {
	repo tenant.Repository
}

// NewGetTenantQuery creates a new GetTenantQuery.
func NewGetTenantQuery(repo tenant.Repository) *GetTenantQuery {
	return &GetTenantQuery{
		repo: repo,
	}
}

// GetTenantRequest is the input for getting a tenant.
type GetTenantRequest struct {
	TenantID string
}

// Handle executes the get tenant query.
func (q *GetTenantQuery) Handle(ctx context.Context, req GetTenantRequest) (*dto.TenantDTO, error) {
	const op = "GetTenantQuery.Handle"

	// Parse tenant ID
	tenantID, err := types.ParseID(req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid tenant_id: %w", op, err)
	}

	// Fetch tenant
	t, err := q.repo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to find tenant: %w", op, err)
	}
	if t == nil {
		return nil, tenant.ErrTenantNotFound(op, req.TenantID)
	}

	return dto.ToTenantDTO(t), nil
}
