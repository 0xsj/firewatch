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

// DeleteUserCommand handles user deletion (soft delete) by admins.
type DeleteUserCommand struct {
	userRepo       user.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewDeleteUserCommand creates a new DeleteUserCommand.
func NewDeleteUserCommand(
	userRepo user.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *DeleteUserCommand {
	return &DeleteUserCommand{
		userRepo:       userRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// DeleteUserRequest is the input for deleting a user.
type DeleteUserRequest struct {
	UserID    types.ID
	Reason    string
	DeletedBy types.ID
}

// Handle executes the delete user command.
func (c *DeleteUserCommand) Handle(ctx context.Context, req DeleteUserRequest) error {
	const op = "DeleteUserCommand.Handle"

	// Find user to delete
	u, err := c.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("%s: failed to find user: %w", op, err)
	}

	// Delete user (domain logic - soft delete)
	if err := u.Delete(req.Reason, req.DeletedBy.String()); err != nil {
		if pkgerrors.IsValidation(err) {
			return pkgerrors.Validation(op, err.Error())
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	// Save user
	if err := c.userRepo.Save(ctx, u); err != nil {
		return fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	c.logger.Info("user deleted",
		logger.String("user_id", u.ID().String()),
		logger.String("deleted_by", req.DeletedBy.String()),
		logger.String("reason", req.Reason),
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
