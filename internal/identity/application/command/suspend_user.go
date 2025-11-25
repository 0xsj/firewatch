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
	userRepo  user.Repository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewSuspendUserCommand creates a new SuspendUserCommand.
func NewSuspendUserCommand(
	userRepo user.Repository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *SuspendUserCommand {
	return &SuspendUserCommand{
		userRepo:  userRepo,
		publisher: publisher,
		logger:    logger,
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

	// Publish user events
	if err := c.publishUserEvents(ctx, u); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
		// Don't fail - user is already suspended
	}

	return nil
}

// publishUserEvents publishes all domain events from the user aggregate.
func (c *SuspendUserCommand) publishUserEvents(ctx context.Context, u *user.User) error {
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
func (c *SuspendUserCommand) convertUserEvent(ctx context.Context, domainEvent user.Event) *messaging.BaseEvent {
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
