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
	repo           domain.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewActivateTemplateCommand creates a new ActivateTemplateCommand.
func NewActivateTemplateCommand(
	repo domain.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *ActivateTemplateCommand {
	return &ActivateTemplateCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
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
	events := messaging.AsDomainEvents(template.Events())
	defer template.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "email", "template", events); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("template_id", template.ID().String()),
			logger.Err(err),
		)
	}

	return &dto.ActivateTemplateResponse{
		Template: dto.MapTemplateToDTO(template),
	}, nil
}
