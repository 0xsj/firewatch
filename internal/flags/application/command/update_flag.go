package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/flags/application/dto"
	"github.com/0xsj/hexagonal-go/internal/flags/domain"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// UpdateFlagCommand handles updating a feature flag.
type UpdateFlagCommand struct {
	repo           domain.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewUpdateFlagCommand creates a new UpdateFlagCommand.
func NewUpdateFlagCommand(
	repo domain.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *UpdateFlagCommand {
	return &UpdateFlagCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// UpdateFlagRequest is the input for updating a flag.
type UpdateFlagRequest struct {
	ID          types.ID
	Name        *string
	Description *string
	UpdatedBy   string
}

// Handle executes the update flag command.
func (c *UpdateFlagCommand) Handle(ctx context.Context, req UpdateFlagRequest) (*dto.UpdateFlagResponse, error) {
	const op = "UpdateFlagCommand.Handle"

	// Find existing flag
	flag, err := c.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Apply updates
	name := ""
	if req.Name != nil {
		name = *req.Name
	}

	description := flag.Description()
	if req.Description != nil {
		description = *req.Description
	}

	if err := flag.Update(name, description, req.UpdatedBy); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Persist changes
	if err := c.repo.Save(ctx, flag); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Publish domain events
	events := messaging.AsDomainEvents(flag.Events())
	defer flag.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "flags", "flag", events); err != nil {
		c.logger.Error("failed to publish flag events",
			logger.String("flag_id", req.ID.String()),
			logger.Err(err),
		)
	}

	c.logger.Info("flag updated",
		logger.String("flag_id", req.ID.String()),
		logger.String("flag_key", flag.Key()),
	)

	return &dto.UpdateFlagResponse{
		Flag: dto.MapFlagToDTO(flag),
	}, nil
}
