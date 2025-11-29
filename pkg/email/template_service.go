package email

import (
	"context"
	"fmt"
)

// TemplateService fetches and renders email templates.
type TemplateService struct {
	repo     TemplateRepository
	renderer *Renderer
}

// NewTemplateService creates a new TemplateService.
func NewTemplateService(repo TemplateRepository, renderer *Renderer) *TemplateService {
	return &TemplateService{
		repo:     repo,
		renderer: renderer,
	}
}

// RenderTemplate fetches an active template by slug and renders it with the provided data.
// Falls back from tenant-specific to system-wide templates.
// Returns rendered subject, HTML body, and text body.
func (s *TemplateService) RenderTemplate(
	ctx context.Context,
	tenantID *string,
	slug string,
	locale string,
	data TemplateData,
) (*RenderedContent, error) {
	const op = "TemplateService.RenderTemplate"

	// Fetch active template
	template, err := s.repo.FindActiveBySlug(ctx, tenantID, slug, locale)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to find template %q: %w", op, slug, err)
	}

	// Render template with data
	rendered, err := s.renderer.Render(
		template.Subject,
		template.BodyHTML,
		template.BodyText,
		data,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to render template %q: %w", op, slug, err)
	}

	return rendered, nil
}

// RenderTemplateWithFallback attempts to render a template, falling back to provided defaults if not found.
// Useful during migration when templates may not exist in the database yet.
func (s *TemplateService) RenderTemplateWithFallback(
	ctx context.Context,
	tenantID *string,
	slug string,
	locale string,
	data TemplateData,
	fallbackSubject string,
	fallbackHTML string,
	fallbackText string,
) (*RenderedContent, error) {
	const op = "TemplateService.RenderTemplateWithFallback"

	// Try to fetch from repository
	template, err := s.repo.FindActiveBySlug(ctx, tenantID, slug, locale)
	if err != nil {
		// Use fallback templates
		rendered, renderErr := s.renderer.Render(
			fallbackSubject,
			fallbackHTML,
			fallbackText,
			data,
		)
		if renderErr != nil {
			return nil, fmt.Errorf("%s: failed to render fallback template: %w", op, renderErr)
		}
		return rendered, nil
	}

	// Render from database template
	rendered, err := s.renderer.Render(
		template.Subject,
		template.BodyHTML,
		template.BodyText,
		data,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to render template %q: %w", op, slug, err)
	}

	return rendered, nil
}
