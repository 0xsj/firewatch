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

// ChangeUserRoleCommand handles role changes by admins.
type ChangeUserRoleCommand struct {
	userRepo  user.Repository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewChangeUserRoleCommand creates a new ChangeUserRoleCommand.
func NewChangeUserRoleCommand(
	userRepo user.Repository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *ChangeUserRoleCommand {
	return &ChangeUserRoleCommand{
		userRepo:  userRepo,
		publisher: publisher,
		logger:    logger,
	}
}

// ChangeUserRoleRequest is the input for changing a user's role.
type ChangeUserRoleRequest struct {
	UserID    types.ID
	NewRole   user.Role
	Reason    string
	ChangedBy types.ID
}

// Handle executes the change user role command.
func (c *ChangeUserRoleCommand) Handle(ctx context.Context, req ChangeUserRoleRequest) error {
	const op = "ChangeUserRoleCommand.Handle"

	// Find user
	u, err := c.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("%s: failed to find user: %w", op, err)
	}

	// Change role (domain logic)
	if err := u.ChangeRole(req.NewRole, req.ChangedBy.String(), req.Reason); err != nil {
		if pkgerrors.IsValidation(err) {
			return pkgerrors.Validation(op, err.Error())
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	// Save user
	if err := c.userRepo.Save(ctx, u); err != nil {
		return fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	c.logger.Info("user role changed",
		logger.String("user_id", u.ID().String()),
		logger.String("new_role", string(req.NewRole)),
		logger.String("changed_by", req.ChangedBy.String()),
		logger.String("reason", req.Reason),
	)

	// Publish user events
	if err := c.publishUserEvents(ctx, u); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
		// Don't fail - role is already changed
	}

	return nil
}

// publishUserEvents publishes all domain events from the user aggregate.
func (c *ChangeUserRoleCommand) publishUserEvents(ctx context.Context, u *user.User) error {
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
func (c *ChangeUserRoleCommand) convertUserEvent(ctx context.Context, domainEvent user.Event) *messaging.BaseEvent {
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
