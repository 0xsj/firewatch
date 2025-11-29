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

// ChangePasswordCommand handles authenticated password changes.
type ChangePasswordCommand struct {
	userRepo       user.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewChangePasswordCommand creates a new ChangePasswordCommand.
func NewChangePasswordCommand(
	userRepo user.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *ChangePasswordCommand {
	return &ChangePasswordCommand{
		userRepo:       userRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// ChangePasswordRequest is the input for changing a password.
type ChangePasswordRequest struct {
	UserID      types.ID
	OldPassword string
	NewPassword string
	IPAddress   string
}

// Handle executes the change password command.
func (c *ChangePasswordCommand) Handle(ctx context.Context, req ChangePasswordRequest) error {
	const op = "ChangePasswordCommand.Handle"

	// Find user
	u, err := c.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("%s: failed to find user: %w", op, err)
	}

	// Create new password
	newPassword, err := user.NewPassword(req.NewPassword, user.DefaultPasswordRequirements())
	if err != nil {
		return pkgerrors.Validation(op, err.Error())
	}

	// Change password (domain logic validates old password)
	if err := u.ChangePassword(req.OldPassword, newPassword, "user"); err != nil {
		c.logger.Warn("password change failed",
			logger.String("user_id", req.UserID.String()),
			logger.String("ip_address", req.IPAddress),
			logger.Err(err),
		)
		return pkgerrors.Unauthorized(op, "current password is incorrect")
	}

	// Save user
	if err := c.userRepo.Save(ctx, u); err != nil {
		return fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	c.logger.Info("password changed",
		logger.String("user_id", u.ID().String()),
		logger.String("ip_address", req.IPAddress),
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
