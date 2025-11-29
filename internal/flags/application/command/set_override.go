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

// SetOverrideCommand handles setting an override for a feature flag.
type SetOverrideCommand struct {
	repo           domain.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewSetOverrideCommand creates a new SetOverrideCommand.
func NewSetOverrideCommand(
	repo domain.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *SetOverrideCommand {
	return &SetOverrideCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// SetOverrideRequest is the input for setting an override.
type SetOverrideRequest struct {
	FlagID     types.ID
	TargetType string
	TargetID   string
	VariantKey string
	SetBy      string
}

// Handle executes the set override command.
func (c *SetOverrideCommand) Handle(ctx context.Context, req SetOverrideRequest) (*dto.SetOverrideResponse, error) {
	const op = "SetOverrideCommand.Handle"

	// Find existing flag
	flag, err := c.repo.FindByID(ctx, req.FlagID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Set override
	if err := flag.SetOverride(req.TargetType, req.TargetID, req.VariantKey, req.SetBy); err != nil {
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

	c.logger.Info("override set for flag",
		logger.String("flag_id", req.FlagID.String()),
		logger.String("flag_key", flag.Key()),
		logger.String("target_type", req.TargetType),
		logger.String("target_id", req.TargetID),
		logger.String("variant_key", req.VariantKey),
	)

	return &dto.SetOverrideResponse{
		Flag: dto.MapFlagToDTO(flag),
	}, nil
}
