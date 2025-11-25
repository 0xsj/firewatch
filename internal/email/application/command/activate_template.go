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

// ActivateTemplateCommand handles activating an email template.
type ActivateTemplateCommand struct {
	repo      domain.Repository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewActivateTemplateCommand creates a new ActivateTemplateCommand.
func NewActivateTemplateCommand(
	repo domain.Repository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *ActivateTemplateCommand {
	return &ActivateTemplateCommand{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// Handle executes the activate template command.
func (c *ActivateTemplateCommand) Handle(ctx context.Context, templateID types.ID, activatedBy *types.ID) (*dto.ActivateTemplateResponse, error) {
	const op = "ActivateTemplateCommand.Handle"

	// Find existing template
	template, err := c.repo.FindByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Activate template
	if err := template.Activate(activatedBy); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Persist changes
	if err := c.repo.Save(ctx, template); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	c.logger.Info("email template activated",
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

	return &dto.ActivateTemplateResponse{
		Template: dto.MapTemplateToDTO(template),
	}, nil
}

// publishEvents publishes all domain events from the template aggregate.
func (c *ActivateTemplateCommand) publishEvents(ctx context.Context, template *domain.Template) error {
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
