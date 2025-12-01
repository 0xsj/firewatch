// internal/audit/application/query/get_entry.go
package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/audit/application/dto"
	"github.com/0xsj/hexagonal-go/internal/audit/domain"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// GetEntryQuery handles retrieving a single audit entry.
type GetEntryQuery struct {
	repo   domain.Repository
	logger logger.Logger
}

// NewGetEntryQuery creates a new GetEntryQuery.
func NewGetEntryQuery(
	repo domain.Repository,
	logger logger.Logger,
) *GetEntryQuery {
	return &GetEntryQuery{
		repo:   repo,
		logger: logger,
	}
}

// Handle retrieves an audit entry by ID.
func (q *GetEntryQuery) Handle(ctx context.Context, req dto.GetEntryRequest) (*dto.GetEntryResponse, error) {
	const op = "GetEntryQuery.Handle"

	// Parse ID
	id, err := types.ParseID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid id: %w", op, err)
	}

	// Fetch from repository
	entry, err := q.repo.FindByID(ctx, id)
	if err != nil {
		q.logger.Error("failed to get audit entry",
			logger.String("id", req.ID),
			logger.Err(err),
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &dto.GetEntryResponse{
		Entry: dto.EntryToDTO(entry),
	}, nil
}
