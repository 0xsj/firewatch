package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/flags/application/dto"
	"github.com/0xsj/hexagonal-go/internal/flags/domain"
)

// ListFlagsQuery handles fetching a list of feature flags.
type ListFlagsQuery struct {
	repo domain.Repository
}

// NewListFlagsQuery creates a new ListFlagsQuery.
func NewListFlagsQuery(repo domain.Repository) *ListFlagsQuery {
	return &ListFlagsQuery{
		repo: repo,
	}
}

// ListFlagsRequest is the input for listing flags.
type ListFlagsRequest struct {
	TenantID string
	Enabled  *bool
	Search   string
	Limit    int
	Offset   int
}

// Handle executes the query to list flags.
func (q *ListFlagsQuery) Handle(ctx context.Context, req ListFlagsRequest) (*dto.ListFlagsResponse, error) {
	const op = "ListFlagsQuery.Handle"

	// Build filters
	filters := domain.NewFilters()

	if req.Enabled != nil {
		filters.WithEnabled(*req.Enabled)
	}

	if req.Search != "" {
		filters.WithSearch(req.Search)
	}

	if req.Limit > 0 {
		filters.WithLimit(req.Limit)
	}

	if req.Offset > 0 {
		filters.WithOffset(req.Offset)
	}

	// Fetch flags
	flags, err := q.repo.FindAll(ctx, req.TenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Determine if there are more results
	hasMore := len(flags) == filters.Limit

	return &dto.ListFlagsResponse{
		Flags:   dto.MapFlagsToSummaryDTO(flags),
		Total:   len(flags),
		Limit:   filters.Limit,
		Offset:  filters.Offset,
		HasMore: hasMore,
	}, nil
}
