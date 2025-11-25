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
	userRepo  user.Repository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewChangePasswordCommand creates a new ChangePasswordCommand.
func NewChangePasswordCommand(
	userRepo user.Repository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *ChangePasswordCommand {
	return &ChangePasswordCommand{
		userRepo:  userRepo,
		publisher: publisher,
		logger:    logger,
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
	if err := u.ChangePassword(req.OldPassword, newPassword, req.IPAddress); err != nil {
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

	// Publish user events
	if err := c.publishUserEvents(ctx, u); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
		// Don't fail - password is already changed
	}

	return nil
}

// publishUserEvents publishes all domain events from the user aggregate.
func (c *ChangePasswordCommand) publishUserEvents(ctx context.Context, u *user.User) error {
	events := u.Events()
	defer u.ClearEvents()

	for _, domainEvent := range events {
		event := c.convertUserEvent(ctx, domainEvent)

		if err := c.publisher.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", domainEvent.Type(), err)
		}

		c.logger.Debug("event published",
			logger.String("event_type", event.Type()),
			logger.String("event_id", event.ID()),
		)
	}

	return nil
}

// convertUserEvent converts a user domain event to a messaging event.
func (c *ChangePasswordCommand) convertUserEvent(ctx context.Context, domainEvent user.Event) *messaging.BaseEvent {
	event := messaging.NewEventFromContext(
		ctx,
		"identity."+domainEvent.Type(),
		"identity",
		domainEvent.Payload(),
	)

	// Add standard metadata
	event.WithTenantID(domainEvent.AggregateTenantID())
	event.WithUserID(domainEvent.AggregateID().String())

	// Add aggregate metadata
	event.WithMetadata("aggregate_id", domainEvent.AggregateID().String())
	event.WithMetadata("aggregate_type", "user")
	event.WithMetadata("event_version", domainEvent.Version())

	return event
}
