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
	repo      domain.Repository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewDeleteTemplateCommand creates a new DeleteTemplateCommand.
func NewDeleteTemplateCommand(
	repo domain.Repository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *DeleteTemplateCommand {
	return &DeleteTemplateCommand{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
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
	if err := c.publishEvents(ctx, template); err != nil {
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

// publishEvents publishes all domain events from the template aggregate.
func (c *DeleteTemplateCommand) publishEvents(ctx context.Context, template *domain.Template) error {
	events := template.Events()
	defer template.ClearEvents()

	for _, domainEvent := range events {
		event := messaging.NewEventFromContext(
			ctx,
			domainEvent.EventType(),
			"email",
			domainEvent.Payload(),
		)

		event.WithTenantID(domainEvent.AggregateTenantID())
		event.WithMetadata("aggregate_id", domainEvent.AggregateID().String())
		event.WithMetadata("aggregate_type", "email_template")

		if err := c.publisher.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", domainEvent.EventType(), err)
		}
	}

	return nil
}
