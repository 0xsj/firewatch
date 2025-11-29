package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/flags/application/dto"
	"github.com/0xsj/hexagonal-go/internal/flags/domain"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// GetFlagQuery handles fetching a single feature flag.
type GetFlagQuery struct {
	repo domain.Repository
}

// NewGetFlagQuery creates a new GetFlagQuery.
func NewGetFlagQuery(repo domain.Repository) *GetFlagQuery {
	return &GetFlagQuery{
		repo: repo,
	}
}

// GetFlagByIDRequest is the input for getting a flag by ID.
type GetFlagByIDRequest struct {
	ID types.ID
}

// GetFlagByKeyRequest is the input for getting a flag by key.
type GetFlagByKeyRequest struct {
	TenantID string
	Key      string
}

// HandleByID executes the query to get a flag by ID.
func (q *GetFlagQuery) HandleByID(ctx context.Context, req GetFlagByIDRequest) (*dto.GetFlagResponse, error) {
	const op = "GetFlagQuery.HandleByID"

	flag, err := q.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &dto.GetFlagResponse{
		Flag: dto.MapFlagToDTO(flag),
	}, nil
}

// HandleByKey executes the query to get a flag by key.
func (q *GetFlagQuery) HandleByKey(ctx context.Context, req GetFlagByKeyRequest) (*dto.GetFlagResponse, error) {
	const op = "GetFlagQuery.HandleByKey"

	flag, err := q.repo.FindByKey(ctx, req.TenantID, req.Key)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &dto.GetFlagResponse{
		Flag: dto.MapFlagToDTO(flag),
	}, nil
}
