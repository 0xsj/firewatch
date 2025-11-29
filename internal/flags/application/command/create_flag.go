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

// CreateFlagCommand handles creating a new feature flag.
type CreateFlagCommand struct {
	repo           domain.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewCreateFlagCommand creates a new CreateFlagCommand.
func NewCreateFlagCommand(
	repo domain.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *CreateFlagCommand {
	return &CreateFlagCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// CreateFlagRequest is the input for creating a flag.
type CreateFlagRequest struct {
	TenantID    string
	Key         string
	Name        string
	Description string
	Enabled     bool
	CreatedBy   string
}

// Handle executes the create flag command.
func (c *CreateFlagCommand) Handle(ctx context.Context, req CreateFlagRequest) (*dto.CreateFlagResponse, error) {
	const op = "CreateFlagCommand.Handle"

	// Check if flag key already exists
	exists, err := c.repo.Exists(ctx, req.TenantID, req.Key)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if exists {
		return nil, domain.ErrFlagAlreadyExists(op, req.Key)
	}

	// Create flag aggregate
	id := types.NewID()
	flag, err := domain.Create(
		id,
		req.TenantID,
		req.Key,
		req.Name,
		req.Description,
		req.Enabled,
		req.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Persist flag
	if err := c.repo.Save(ctx, flag); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Publish domain events
	events := messaging.AsDomainEvents(flag.Events())
	defer flag.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "flags", "flag", events); err != nil {
		c.logger.Error("failed to publish flag events",
			logger.String("flag_id", id.String()),
			logger.String("flag_key", req.Key),
			logger.Err(err),
		)
	}

	c.logger.Info("flag created",
		logger.String("flag_id", id.String()),
		logger.String("flag_key", req.Key),
		logger.String("tenant_id", req.TenantID),
	)

	return &dto.CreateFlagResponse{
		Flag: dto.MapFlagToDTO(flag),
	}, nil
}
