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

// EnableFlagCommand handles enabling a feature flag.
type EnableFlagCommand struct {
	repo           domain.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewEnableFlagCommand creates a new EnableFlagCommand.
func NewEnableFlagCommand(
	repo domain.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *EnableFlagCommand {
	return &EnableFlagCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// EnableFlagRequest is the input for enabling a flag.
type EnableFlagRequest struct {
	ID        types.ID
	EnabledBy string
}

// Handle executes the enable flag command.
func (c *EnableFlagCommand) Handle(ctx context.Context, req EnableFlagRequest) (*dto.EnableFlagResponse, error) {
	const op = "EnableFlagCommand.Handle"

	// Find existing flag
	flag, err := c.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Enable flag
	if err := flag.Enable(req.EnabledBy); err != nil {
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

	c.logger.Info("flag enabled",
		logger.String("flag_id", req.ID.String()),
		logger.String("flag_key", flag.Key()),
		logger.String("enabled_by", req.EnabledBy),
	)

	return &dto.EnableFlagResponse{
		Flag: dto.MapFlagToDTO(flag),
	}, nil
}
