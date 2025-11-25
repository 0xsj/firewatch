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
	userRepo  user.Repository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewReactivateUserCommand creates a new ReactivateUserCommand.
func NewReactivateUserCommand(
	userRepo user.Repository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *ReactivateUserCommand {
	return &ReactivateUserCommand{
		userRepo:  userRepo,
		publisher: publisher,
		logger:    logger,
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

	// Publish user events
	if err := c.publishUserEvents(ctx, u); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
		// Don't fail - user is already reactivated
	}

	return nil
}

// publishUserEvents publishes all domain events from the user aggregate.
func (c *ReactivateUserCommand) publishUserEvents(ctx context.Context, u *user.User) error {
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
func (c *ReactivateUserCommand) convertUserEvent(ctx context.Context, domainEvent user.Event) *messaging.BaseEvent {
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
