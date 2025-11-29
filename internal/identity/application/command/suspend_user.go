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

// SuspendUserCommand handles user suspension by admins.
type SuspendUserCommand struct {
	userRepo       user.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewSuspendUserCommand creates a new SuspendUserCommand.
func NewSuspendUserCommand(
	userRepo user.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *SuspendUserCommand {
	return &SuspendUserCommand{
		userRepo:       userRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// SuspendUserRequest is the input for suspending a user.
type SuspendUserRequest struct {
	UserID      types.ID
	Reason      string
	SuspendedBy types.ID
}

// Handle executes the suspend user command.
func (c *SuspendUserCommand) Handle(ctx context.Context, req SuspendUserRequest) error {
	const op = "SuspendUserCommand.Handle"

	// Find user to suspend
	u, err := c.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("%s: failed to find user: %w", op, err)
	}

	// Suspend user (domain logic)
	if err := u.Suspend(req.Reason, req.SuspendedBy.String()); err != nil {
		if pkgerrors.IsValidation(err) {
			return pkgerrors.Validation(op, err.Error())
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	// Save user
	if err := c.userRepo.Save(ctx, u); err != nil {
		return fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	c.logger.Info("user suspended",
		logger.String("user_id", u.ID().String()),
		logger.String("suspended_by", req.SuspendedBy.String()),
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
