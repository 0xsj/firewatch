// internal/audit/application/query/get_actor_activity.go
package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/audit/application/dto"
	"github.com/0xsj/hexagonal-go/internal/audit/domain"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// GetActorActivityQuery handles retrieving audit entries for a specific actor.
type GetActorActivityQuery struct {
	repo   domain.Repository
	logger logger.Logger
}

// NewGetActorActivityQuery creates a new GetActorActivityQuery.
func NewGetActorActivityQuery(
	repo domain.Repository,
	logger logger.Logger,
) *GetActorActivityQuery {
	return &GetActorActivityQuery{
		repo:   repo,
		logger: logger,
	}
}

// Handle retrieves audit entries for a specific actor (user).
func (q *GetActorActivityQuery) Handle(ctx context.Context, req dto.GetActorActivityRequest) (*dto.GetActorActivityResponse, error) {
	const op = "GetActorActivityQuery.Handle"

	// Build filters
	filters := q.buildFiltersForActor(req)
	page := dto.MapPageFromRequest(req.Limit, req.Offset)

	// Fetch from repository
	result, err := q.repo.List(ctx, filters, page)
	if err != nil {
		q.logger.Error("failed to get actor activity",
			logger.String("user_id", req.UserID),
			logger.String("tenant_id", req.TenantID),
			logger.Err(err),
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	q.logger.Debug("actor activity retrieved",
		logger.String("user_id", req.UserID),
		logger.Int("count", len(result.Entries)),
		logger.Int("total", result.Total),
	)

	return dto.BuildActorActivityResponse(req.UserID, result), nil
}

// buildFiltersForActor constructs domain filters for actor activity.
func (q *GetActorActivityQuery) buildFiltersForActor(req dto.GetActorActivityRequest) domain.Filters {
	filters := domain.Filters{}

	// Required: filter by actor (user ID)
	filters.UserID = &req.UserID

	// Optional: narrow by tenant
	if req.TenantID != "" {
		filters.TenantID = &req.TenantID
	}

	// Optional: narrow by event type
	if req.EventType != "" {
		filters.EventType = &req.EventType
	}

	// Optional: narrow by source domain
	if req.Source != "" {
		filters.Source = &req.Source
	}

	// Optional: timestamp range
	if req.FromTimestamp != nil {
		ts := types.NewTimestamp(*req.FromTimestamp)
		filters.FromTimestamp = &ts
	}
	if req.ToTimestamp != nil {
		ts := types.NewTimestamp(*req.ToTimestamp)
		filters.ToTimestamp = &ts
	}

	return filters
}
