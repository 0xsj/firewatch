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

// RemoveOverrideCommand handles removing an override from a feature flag.
type RemoveOverrideCommand struct {
	repo           domain.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewRemoveOverrideCommand creates a new RemoveOverrideCommand.
func NewRemoveOverrideCommand(
	repo domain.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *RemoveOverrideCommand {
	return &RemoveOverrideCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// RemoveOverrideRequest is the input for removing an override.
type RemoveOverrideRequest struct {
	FlagID     types.ID
	TargetType string
	TargetID   string
	RemovedBy  string
}

// Handle executes the remove override command.
func (c *RemoveOverrideCommand) Handle(ctx context.Context, req RemoveOverrideRequest) (*dto.RemoveOverrideResponse, error) {
	const op = "RemoveOverrideCommand.Handle"

	// Find existing flag
	flag, err := c.repo.FindByID(ctx, req.FlagID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Remove override
	if err := flag.RemoveOverride(req.TargetType, req.TargetID, req.RemovedBy); err != nil {
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
			logger.String("flag_id", req.FlagID.String()),
			logger.Err(err),
		)
	}

	c.logger.Info("override removed from flag",
		logger.String("flag_id", req.FlagID.String()),
		logger.String("flag_key", flag.Key()),
		logger.String("target_type", req.TargetType),
		logger.String("target_id", req.TargetID),
	)

	return &dto.RemoveOverrideResponse{
		Flag: dto.MapFlagToDTO(flag),
	}, nil
}
