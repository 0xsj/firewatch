// internal/audit/application/query/get_resource_trail.go
package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/audit/application/dto"
	"github.com/0xsj/hexagonal-go/internal/audit/domain"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// GetResourceTrailQuery handles retrieving audit trail for a specific resource.
type GetResourceTrailQuery struct {
	repo   domain.Repository
	logger logger.Logger
}

// NewGetResourceTrailQuery creates a new GetResourceTrailQuery.
func NewGetResourceTrailQuery(
	repo domain.Repository,
	logger logger.Logger,
) *GetResourceTrailQuery {
	return &GetResourceTrailQuery{
		repo:   repo,
		logger: logger,
	}
}

// Handle retrieves the audit trail for a specific resource.
func (q *GetResourceTrailQuery) Handle(ctx context.Context, req dto.GetResourceTrailRequest) (*dto.GetResourceTrailResponse, error) {
	const op = "GetResourceTrailQuery.Handle"

	// Build filters based on resource type
	filters := q.buildFiltersForResource(req)
	page := dto.MapPageFromRequest(req.Limit, req.Offset)

	// Fetch from repository
	result, err := q.repo.List(ctx, filters, page)
	if err != nil {
		q.logger.Error("failed to get resource audit trail",
			logger.String("resource_type", req.ResourceType),
			logger.String("resource_id", req.ResourceID),
			logger.Err(err),
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	q.logger.Debug("resource audit trail retrieved",
		logger.String("resource_type", req.ResourceType),
		logger.String("resource_id", req.ResourceID),
		logger.Int("count", len(result.Entries)),
		logger.Int("total", result.Total),
	)

	return dto.BuildResourceTrailResponse(req.ResourceType, req.ResourceID, result), nil
}

// buildFiltersForResource constructs domain filters based on resource type.
//
// Resource type mapping:
//   - "tenant": filters by TenantID (tenant is the aggregate)
//   - "user": filters by UserID (user as subject, not actor)
//   - other: filters by source domain (e.g., "flags" → events from flags domain)
//
// Note: For full resource trail functionality with arbitrary resource types,
// consider enhancing domain.Filters to support:
//   - EventTypePrefix for LIKE-based event type matching
//   - AggregateID for filtering by metadata.aggregate_id
func (q *GetResourceTrailQuery) buildFiltersForResource(req dto.GetResourceTrailRequest) domain.Filters {
	filters := domain.Filters{}

	// Apply tenant filter if provided
	if req.TenantID != "" {
		filters.TenantID = &req.TenantID
	}

	// Map resource type to appropriate filter
	switch req.ResourceType {
	case "tenant":
		// For tenant resources, the resource ID is the tenant ID
		filters.TenantID = &req.ResourceID

	case "user":
		// For user resources, filter by user ID
		// Note: This finds events where user is the subject (aggregate),
		// which is different from events BY a user (actor)
		filters.UserID = &req.ResourceID

	default:
		// For other resource types, we filter by source domain
		// This returns all events from that domain within the tenant
		// Full resource ID filtering would require payload/metadata querying
		filters.Source = &req.ResourceType

		q.logger.Debug("resource trail using source filter",
			logger.String("resource_type", req.ResourceType),
			logger.String("resource_id", req.ResourceID),
			logger.String("note", "full resource ID filtering requires repository enhancement"),
		)
	}

	// Apply timestamp filters
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
