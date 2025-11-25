package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/email/application/dto"
	"github.com/0xsj/hexagonal-go/internal/email/domain"
)

// ListTemplatesQuery handles listing templates with filters.
type ListTemplatesQuery struct {
	repo domain.Repository
}

// NewListTemplatesQuery creates a new ListTemplatesQuery.
func NewListTemplatesQuery(repo domain.Repository) *ListTemplatesQuery {
	return &ListTemplatesQuery{
		repo: repo,
	}
}

// Handle executes the list templates query.
func (q *ListTemplatesQuery) Handle(ctx context.Context, req dto.ListTemplatesRequest) (*dto.ListTemplatesResponse, error) {
	const op = "ListTemplatesQuery.Handle"

	// Build filters from request
	filters := domain.DefaultListFilters()

	filters.TenantID = req.TenantID
	filters.IncludeSystemTemplates = req.IncludeSystemTemplates

	if req.Status != nil {
		status, err := domain.ParseStatus(*req.Status)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		filters.Status = &status
	}

	if req.Locale != nil {
		locale, err := domain.ParseLocale(*req.Locale)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		filters.Locale = &locale
	}

	if req.SlugContains != "" {
		filters.SlugContains = req.SlugContains
	}

	if req.NameContains != "" {
		filters.NameContains = req.NameContains
	}

	if req.Limit > 0 {
		filters.Limit = req.Limit
	}

	if req.Offset > 0 {
		filters.Offset = req.Offset
	}

	if req.SortBy != "" {
		filters.SortBy = domain.SortField(req.SortBy)
	}

	if req.SortOrder != "" {
		filters.SortOrder = domain.SortOrder(req.SortOrder)
	}

	// Get templates
	templates, err := q.repo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Get total count
	total, err := q.repo.Count(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &dto.ListTemplatesResponse{
		Templates: dto.MapTemplatesToDTO(templates),
		Total:     total,
		Limit:     filters.Limit,
		Offset:    filters.Offset,
		HasMore:   filters.Offset+len(templates) < total,
	}, nil
}
