package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/internal/identity/infrastructure/repository"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// ResetPasswordCommand handles password reset with token.
type ResetPasswordCommand struct {
	userRepo  user.Repository
	tokenRepo *repository.PostgresTokenRepository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewResetPasswordCommand creates a new ResetPasswordCommand.
func NewResetPasswordCommand(
	userRepo user.Repository,
	tokenRepo *repository.PostgresTokenRepository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *ResetPasswordCommand {
	return &ResetPasswordCommand{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		publisher: publisher,
		logger:    logger,
	}
}

// ResetPasswordRequest is the input for resetting a password.
type ResetPasswordRequest struct {
	Token       string
	NewPassword string
	IPAddress   string
}

// Handle executes the reset password command.
func (c *ResetPasswordCommand) Handle(ctx context.Context, req ResetPasswordRequest) error {
	const op = "ResetPasswordCommand.Handle"

	// Validate token
	resetToken, err := c.tokenRepo.FindPasswordResetToken(ctx, req.Token)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			return pkgerrors.Validation(op, "invalid or expired reset token")
		}
		return fmt.Errorf("%s: failed to find reset token: %w", op, err)
	}

	// Find user
	u, err := c.userRepo.FindByID(ctx, resetToken.UserID())
	if err != nil {
		return fmt.Errorf("%s: failed to find user: %w", op, err)
	}

	// Create new password
	newPassword, err := user.NewPassword(req.NewPassword, user.DefaultPasswordRequirements())
	if err != nil {
		return pkgerrors.Validation(op, err.Error())
	}

	// Reset password (domain logic)
	if err := u.ResetPassword(newPassword, req.IPAddress); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Save user
	if err := c.userRepo.Save(ctx, u); err != nil {
		return fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	// Mark token as used
	if err := c.tokenRepo.MarkPasswordResetTokenUsed(ctx, req.Token); err != nil {
		c.logger.Error("failed to mark token as used",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
		// Don't fail the operation if marking fails - password is already reset
	}

	c.logger.Info("password reset successful",
		logger.String("user_id", u.ID().String()),
		logger.String("ip_address", req.IPAddress),
	)

	// Publish user events
	if err := c.publishUserEvents(ctx, u); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
		// Don't fail - password is already reset
	}

	return nil
}

// publishUserEvents publishes all domain events from the user aggregate.
func (c *ResetPasswordCommand) publishUserEvents(ctx context.Context, u *user.User) error {
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
func (c *ResetPasswordCommand) convertUserEvent(ctx context.Context, domainEvent user.Event) *messaging.BaseEvent {
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
