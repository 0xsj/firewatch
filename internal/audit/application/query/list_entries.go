// internal/audit/application/query/list_entries.go
package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/audit/application/dto"
	"github.com/0xsj/hexagonal-go/internal/audit/domain"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// ListEntriesQuery handles listing audit entries with filters.
type ListEntriesQuery struct {
	repo   domain.Repository
	logger logger.Logger
}

// NewListEntriesQuery creates a new ListEntriesQuery.
func NewListEntriesQuery(
	repo domain.Repository,
	logger logger.Logger,
) *ListEntriesQuery {
	return &ListEntriesQuery{
		repo:   repo,
		logger: logger,
	}
}

// Handle retrieves audit entries matching the filters.
func (q *ListEntriesQuery) Handle(ctx context.Context, req dto.ListEntriesRequest) (*dto.ListEntriesResponse, error) {
	const op = "ListEntriesQuery.Handle"

	// Map request to domain filters and pagination
	filters := dto.MapFiltersFromListRequest(req)
	page := dto.MapPageFromRequest(req.Limit, req.Offset)

	// Fetch from repository
	result, err := q.repo.List(ctx, filters, page)
	if err != nil {
		q.logger.Error("failed to list audit entries",
			logger.String("tenant_id", req.TenantID),
			logger.String("user_id", req.UserID),
			logger.String("event_type", req.EventType),
			logger.Err(err),
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	q.logger.Debug("audit entries listed",
		logger.Int("count", len(result.Entries)),
		logger.Int("total", result.Total),
	)

	return dto.BuildListEntriesResponse(result), nil
}
