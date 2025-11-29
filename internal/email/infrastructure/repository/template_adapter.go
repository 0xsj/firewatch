package repository

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/email/domain"
	"github.com/0xsj/hexagonal-go/pkg/email"
)

// TemplateRepositoryAdapter adapts PostgresRepository to implement pkg/email.TemplateRepository.
// This allows the notifications domain to fetch templates without importing the email domain.
type TemplateRepositoryAdapter struct {
	repo *PostgresRepository
}

// NewTemplateRepositoryAdapter creates a new adapter.
func NewTemplateRepositoryAdapter(repo *PostgresRepository) *TemplateRepositoryAdapter {
	return &TemplateRepositoryAdapter{
		repo: repo,
	}
}

// FindActiveBySlug retrieves an active template by slug and locale.
// Implements pkg/email.TemplateRepository.
func (a *TemplateRepositoryAdapter) FindActiveBySlug(
	ctx context.Context,
	tenantID *string,
	slug string,
	locale string,
) (*email.ResolvedTemplate, error) {
	const op = "TemplateRepositoryAdapter.FindActiveBySlug"

	// Parse locale
	parsedLocale, err := domain.ParseLocale(locale)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid locale %q: %w", op, locale, err)
	}

	// Fetch from domain repository
	template, err := a.repo.FindActiveBySlug(ctx, tenantID, slug, parsedLocale)
	if err != nil {
		return nil, err
	}

	// Convert to pkg/email type
	return &email.ResolvedTemplate{
		Slug:     template.Slug(),
		Locale:   template.Locale().String(),
		Subject:  template.Subject(),
		BodyHTML: template.BodyHTML(),
		BodyText: template.BodyText(),
	}, nil
}
