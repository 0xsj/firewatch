package email

import (
	"context"
)

// ResolvedTemplate contains the raw template content before rendering.
type ResolvedTemplate struct {
	Slug     string
	Locale   string
	Subject  string
	BodyHTML string
	BodyText string
}

// TemplateRepository is the port for fetching email templates.
// Implementations live in infrastructure layers.
type TemplateRepository interface {
	// FindActiveBySlug retrieves an active template by slug and locale.
	// If tenantID is provided, searches tenant-specific first, then falls back to system-wide.
	// Returns ErrTemplateNotFound if no active template is found.
	FindActiveBySlug(ctx context.Context, tenantID *string, slug string, locale string) (*ResolvedTemplate, error)
}
