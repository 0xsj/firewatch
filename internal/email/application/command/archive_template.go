package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/email/application/dto"
	"github.com/0xsj/hexagonal-go/internal/email/domain"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// ArchiveTemplateCommand handles archiving an email template.
type ArchiveTemplateCommand struct {
	repo           domain.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewArchiveTemplateCommand creates a new ArchiveTemplateCommand.
func NewArchiveTemplateCommand(
	repo domain.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *ArchiveTemplateCommand {
	return &ArchiveTemplateCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// Handle executes the archive template command.
func (c *ArchiveTemplateCommand) Handle(ctx context.Context, templateID types.ID, archivedBy *types.ID) (*dto.ArchiveTemplateResponse, error) {
	const op = "ArchiveTemplateCommand.Handle"

	// Find existing template
	template, err := c.repo.FindByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Archive template
	if err := template.Archive(archivedBy); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Persist changes
	if err := c.repo.Save(ctx, template); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	c.logger.Info("email template archived",
		logger.String("template_id", template.ID().String()),
		logger.String("slug", template.Slug()),
		logger.String("locale", template.Locale().String()),
	)

	// Publish domain events
	events := messaging.AsDomainEvents(template.Events())
	defer template.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "email", "template", events); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("template_id", template.ID().String()),
			logger.Err(err),
		)
	}

	return &dto.ArchiveTemplateResponse{
		Template: dto.MapTemplateToDTO(template),
	}, nil
}
