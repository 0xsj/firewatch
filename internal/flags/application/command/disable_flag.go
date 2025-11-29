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

// DisableFlagCommand handles disabling a feature flag.
type DisableFlagCommand struct {
	repo           domain.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewDisableFlagCommand creates a new DisableFlagCommand.
func NewDisableFlagCommand(
	repo domain.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *DisableFlagCommand {
	return &DisableFlagCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// DisableFlagRequest is the input for disabling a flag.
type DisableFlagRequest struct {
	ID         types.ID
	DisabledBy string
}

// Handle executes the disable flag command.
func (c *DisableFlagCommand) Handle(ctx context.Context, req DisableFlagRequest) (*dto.DisableFlagResponse, error) {
	const op = "DisableFlagCommand.Handle"

	// Find existing flag
	flag, err := c.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Disable flag
	if err := flag.Disable(req.DisabledBy); err != nil {
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

	c.logger.Info("flag disabled",
		logger.String("flag_id", req.ID.String()),
		logger.String("flag_key", flag.Key()),
		logger.String("disabled_by", req.DisabledBy),
	)

	return &dto.DisableFlagResponse{
		Flag: dto.MapFlagToDTO(flag),
	}, nil
}
