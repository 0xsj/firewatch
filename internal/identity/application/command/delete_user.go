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
	userRepo  user.Repository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewDeleteUserCommand creates a new DeleteUserCommand.
func NewDeleteUserCommand(
	userRepo user.Repository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *DeleteUserCommand {
	return &DeleteUserCommand{
		userRepo:  userRepo,
		publisher: publisher,
		logger:    logger,
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

	// Publish user events
	if err := c.publishUserEvents(ctx, u); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
		// Don't fail - user is already deleted
	}

	return nil
}

// publishUserEvents publishes all domain events from the user aggregate.
func (c *DeleteUserCommand) publishUserEvents(ctx context.Context, u *user.User) error {
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
func (c *DeleteUserCommand) convertUserEvent(ctx context.Context, domainEvent user.Event) *messaging.BaseEvent {
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
