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

// UpdateTemplateCommand handles updating an email template.
type UpdateTemplateCommand struct {
	repo           domain.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewUpdateTemplateCommand creates a new UpdateTemplateCommand.
func NewUpdateTemplateCommand(
	repo domain.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *UpdateTemplateCommand {
	return &UpdateTemplateCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// Handle executes the update template command.
func (c *UpdateTemplateCommand) Handle(ctx context.Context, templateID types.ID, req dto.UpdateTemplateRequest, updatedBy *types.ID) (*dto.UpdateTemplateResponse, error) {
	const op = "UpdateTemplateCommand.Handle"

	// Find existing template
	template, err := c.repo.FindByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Extract values from request (use empty string for nil pointers to skip update)
	name := ""
	if req.Name != nil {
		name = *req.Name
	}

	description := template.Description()
	if req.Description != nil {
		description = *req.Description
	}

	subject := ""
	if req.Subject != nil {
		subject = *req.Subject
	}

	bodyHTML := ""
	if req.BodyHTML != nil {
		bodyHTML = *req.BodyHTML
	}

	bodyText := template.BodyText()
	if req.BodyText != nil {
		bodyText = *req.BodyText
	}

	// Map variables if provided
	var variables domain.Variables
	if req.Variables != nil {
		variables, err = mapVariablesFromRequest(req.Variables)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	// Update template
	if err := template.Update(
		name,
		description,
		subject,
		bodyHTML,
		bodyText,
		variables,
		updatedBy,
	); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Persist changes
	if err := c.repo.Save(ctx, template); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	c.logger.Info("email template updated",
		logger.String("template_id", template.ID().String()),
		logger.String("slug", template.Slug()),
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

	return &dto.UpdateTemplateResponse{
		Template: dto.MapTemplateToDTO(template),
	}, nil
}
