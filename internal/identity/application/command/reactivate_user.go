package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// ReactivateUserCommand handles user reactivation by admins.
type ReactivateUserCommand struct {
	userRepo       user.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewReactivateUserCommand creates a new ReactivateUserCommand.
func NewReactivateUserCommand(
	userRepo user.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *ReactivateUserCommand {
	return &ReactivateUserCommand{
		userRepo:       userRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// ReactivateUserRequest is the input for reactivating a user.
type ReactivateUserRequest struct {
	UserID        types.ID
	ReactivatedBy types.ID
}

// Handle executes the reactivate user command.
func (c *ReactivateUserCommand) Handle(ctx context.Context, req ReactivateUserRequest) error {
	const op = "ReactivateUserCommand.Handle"

	// Find user to reactivate
	u, err := c.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("%s: failed to find user: %w", op, err)
	}

	// Reactivate user (domain logic)
	if err := u.Reactivate(req.ReactivatedBy.String()); err != nil {
		if pkgerrors.IsValidation(err) {
			return pkgerrors.Validation(op, err.Error())
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	// Save user
	if err := c.userRepo.Save(ctx, u); err != nil {
		return fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	c.logger.Info("user reactivated",
		logger.String("user_id", u.ID().String()),
		logger.String("reactivated_by", req.ReactivatedBy.String()),
	)

	// Publish domain events
	events := messaging.AsDomainEvents(u.Events())
	defer u.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "identity", "user", events); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
	}

	return nil
}
