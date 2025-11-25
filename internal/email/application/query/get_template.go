package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/email/application/dto"
	"github.com/0xsj/hexagonal-go/internal/email/domain"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// GetTemplateQuery handles retrieving a template by ID.
type GetTemplateQuery struct {
	repo domain.Repository
}

// NewGetTemplateQuery creates a new GetTemplateQuery.
func NewGetTemplateQuery(repo domain.Repository) *GetTemplateQuery {
	return &GetTemplateQuery{
		repo: repo,
	}
}

// Handle executes the get template query by ID.
func (q *GetTemplateQuery) Handle(ctx context.Context, templateID types.ID) (*dto.GetTemplateResponse, error) {
	const op = "GetTemplateQuery.Handle"

	if templateID.IsEmpty() {
		return nil, fmt.Errorf("%s: template ID is required", op)
	}

	template, err := q.repo.FindByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &dto.GetTemplateResponse{
		Template: dto.MapTemplateToDTO(template),
	}, nil
}

// HandleBySlug executes the get template query by slug.
func (q *GetTemplateQuery) HandleBySlug(ctx context.Context, req dto.GetTemplateBySlugRequest) (*dto.GetTemplateResponse, error) {
	const op = "GetTemplateQuery.HandleBySlug"

	if req.Slug == "" {
		return nil, fmt.Errorf("%s: slug is required", op)
	}

	locale, err := domain.ParseLocale(req.Locale)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	template, err := q.repo.FindBySlug(ctx, req.TenantID, req.Slug, locale)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &dto.GetTemplateResponse{
		Template: dto.MapTemplateToDTO(template),
	}, nil
}
