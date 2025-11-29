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

// DeleteFlagCommand handles deleting a feature flag.
type DeleteFlagCommand struct {
	repo           domain.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewDeleteFlagCommand creates a new DeleteFlagCommand.
func NewDeleteFlagCommand(
	repo domain.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *DeleteFlagCommand {
	return &DeleteFlagCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// DeleteFlagRequest is the input for deleting a flag.
type DeleteFlagRequest struct {
	ID        types.ID
	DeletedBy string
}

// Handle executes the delete flag command.
func (c *DeleteFlagCommand) Handle(ctx context.Context, req DeleteFlagRequest) (*dto.DeleteFlagResponse, error) {
	const op = "DeleteFlagCommand.Handle"

	// Find existing flag
	flag, err := c.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Mark as deleted (emits event)
	flag.MarkDeleted(req.DeletedBy)

	// Publish domain events before deletion
	events := messaging.AsDomainEvents(flag.Events())
	defer flag.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "flags", "flag", events); err != nil {
		c.logger.Error("failed to publish flag events",
			logger.String("flag_id", req.ID.String()),
			logger.Err(err),
		)
	}

	// Delete from repository
	if err := c.repo.Delete(ctx, req.ID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	c.logger.Info("flag deleted",
		logger.String("flag_id", req.ID.String()),
		logger.String("flag_key", flag.Key()),
		logger.String("deleted_by", req.DeletedBy),
	)

	return &dto.DeleteFlagResponse{
		Success: true,
		ID:      req.ID.String(),
	}, nil
}
