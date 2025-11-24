package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// RegisterUserCommand handles user registration.
type RegisterUserCommand struct {
	repo      user.Repository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewRegisterUserCommand creates a new RegisterUserCommand.
func NewRegisterUserCommand(
	repo user.Repository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *RegisterUserCommand {
	return &RegisterUserCommand{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// RegisterRequest is the input for user registration.
type RegisterRequest struct {
	TenantID string
	Email    string
	Password string
	Role     user.Role
}

// Handle executes the register user command.
func (c *RegisterUserCommand) Handle(ctx context.Context, req RegisterRequest) (*dto.UserDTO, error) {
	const op = "RegisterUserCommand.Handle"

	// Check if email already exists
	email, err := user.NewEmail(req.Email)
	if err != nil {
		return nil, pkgerrors.Validation(op, "invalid email format")
	}

	exists, err := c.repo.EmailExists(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to check email: %w", op, err)
	}

	if exists {
		return nil, pkgerrors.Conflict(op, "email address is already registered")
	}

	// Hash password with default requirements
	password, err := user.NewPassword(req.Password, user.DefaultPasswordRequirements())
	if err != nil {
		return nil, pkgerrors.Validation(op, err.Error())
	}

	// Create user
	userID := types.NewID()
	u, err := user.Register(
		userID,
		req.TenantID,
		email,
		password,
		req.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Save to repository
	if err := c.repo.Save(ctx, u); err != nil {
		return nil, fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	c.logger.Info("user registered",
		logger.String("user_id", u.ID().String()),
		logger.String("email", u.Email().String()),
		logger.String("tenant_id", u.TenantID()),
	)

	// Publish domain events
	if err := c.publishEvents(ctx, u); err != nil {
		// Log but don't fail - event publishing is best-effort
		c.logger.Error("failed to publish events",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
	}

	return dto.NewUserResponse(u), nil
}

// publishEvents publishes all domain events from the aggregate.
func (c *RegisterUserCommand) publishEvents(ctx context.Context, u *user.User) error {
	events := u.Events()
	defer u.ClearEvents() // Clear after publishing

	for _, domainEvent := range events {
		// Convert domain event to messaging event
		event := c.convertDomainEvent(ctx, domainEvent)

		// Publish to event bus
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

// convertDomainEvent converts a domain event to a messaging event.
func (c *RegisterUserCommand) convertDomainEvent(ctx context.Context, domainEvent user.Event) *messaging.BaseEvent {
	// Create messaging event with context metadata
	event := messaging.NewEventFromContext(
		ctx,
		"identity."+domainEvent.Type(), // Prefix with domain: "identity.user.registered"
		"identity",
		domainEvent.Payload(),
	)

	// Add standard metadata for cross-cutting concerns
	event.WithTenantID(domainEvent.AggregateTenantID())
	event.WithUserID(domainEvent.AggregateID().String())

	// Add aggregate metadata for event sourcing/replay
	event.WithMetadata("aggregate_id", domainEvent.AggregateID().String())
	event.WithMetadata("aggregate_type", "user")

	return event
}
