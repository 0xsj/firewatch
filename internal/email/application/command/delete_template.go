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

// DeleteTemplateCommand handles deleting an email template.
type DeleteTemplateCommand struct {
	repo           domain.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewDeleteTemplateCommand creates a new DeleteTemplateCommand.
func NewDeleteTemplateCommand(
	repo domain.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *DeleteTemplateCommand {
	return &DeleteTemplateCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// Handle executes the delete template command.
func (c *DeleteTemplateCommand) Handle(ctx context.Context, templateID types.ID, deletedBy *types.ID) (*dto.DeleteTemplateResponse, error) {
	const op = "DeleteTemplateCommand.Handle"

	// Find existing template
	template, err := c.repo.FindByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Mark template for deletion (validates status)
	if err := template.MarkDeleted(deletedBy); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Delete from repository
	if err := c.repo.Delete(ctx, templateID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	c.logger.Info("email template deleted",
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

	return &dto.DeleteTemplateResponse{
		Success: true,
		ID:      templateID.String(),
	}, nil
}
