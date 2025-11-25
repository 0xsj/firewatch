package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/email/application/dto"
	"github.com/0xsj/hexagonal-go/internal/email/domain"
	"github.com/0xsj/hexagonal-go/pkg/email"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// PreviewTemplateQuery handles rendering a template preview.
type PreviewTemplateQuery struct {
	repo     domain.Repository
	renderer *email.Renderer
}

// NewPreviewTemplateQuery creates a new PreviewTemplateQuery.
func NewPreviewTemplateQuery(repo domain.Repository, renderer *email.Renderer) *PreviewTemplateQuery {
	return &PreviewTemplateQuery{
		repo:     repo,
		renderer: renderer,
	}
}

// Handle executes the preview template query by ID.
func (q *PreviewTemplateQuery) Handle(ctx context.Context, req dto.PreviewTemplateRequest) (*dto.PreviewTemplateResponse, error) {
	const op = "PreviewTemplateQuery.Handle"

	templateID, err := types.ParseID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid template ID: %w", op, err)
	}

	template, err := q.repo.FindByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return q.renderPreview(op, template, req.Data)
}

// HandleBySlug executes the preview template query by slug.
func (q *PreviewTemplateQuery) HandleBySlug(ctx context.Context, req dto.PreviewTemplateBySlugRequest) (*dto.PreviewTemplateResponse, error) {
	const op = "PreviewTemplateQuery.HandleBySlug"

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

	return q.renderPreview(op, template, req.Data)
}

// renderPreview renders the template with provided data.
func (q *PreviewTemplateQuery) renderPreview(op string, template *domain.Template, data map[string]interface{}) (*dto.PreviewTemplateResponse, error) {
	// Validate required variables
	if err := q.validateRequiredVariables(template.Variables(), data); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Apply defaults
	data = q.applyDefaults(template.Variables(), data)

	// Render template
	rendered, err := q.renderer.Render(
		template.Subject(),
		template.BodyHTML(),
		template.BodyText(),
		data,
	)
	if err != nil {
		return nil, domain.ErrTemplateRenderFailed(err)
	}

	return &dto.PreviewTemplateResponse{
		Subject:  rendered.Subject,
		BodyHTML: rendered.BodyHTML,
		BodyText: rendered.BodyText,
	}, nil
}

// validateRequiredVariables checks that all required variables are provided.
func (q *PreviewTemplateQuery) validateRequiredVariables(vars domain.Variables, data map[string]interface{}) error {
	for _, v := range vars {
		if v.Required {
			if _, exists := data[v.Name]; !exists {
				if v.Default == nil {
					return domain.ErrTemplateMissingVariable(v.Name)
				}
			}
		}
	}
	return nil
}

// applyDefaults applies default values for missing optional variables.
func (q *PreviewTemplateQuery) applyDefaults(vars domain.Variables, data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		result[k] = v
	}

	for _, v := range vars {
		if _, exists := result[v.Name]; !exists && v.Default != nil {
			result[v.Name] = *v.Default
		}
	}

	return result
}
