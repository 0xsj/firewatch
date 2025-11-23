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

// VerifyEmailCommand handles email verification.
type VerifyEmailCommand struct {
	repo      user.Repository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewVerifyEmailCommand creates a new VerifyEmailCommand.
func NewVerifyEmailCommand(
	repo user.Repository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *VerifyEmailCommand {
	return &VerifyEmailCommand{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// VerifyEmailRequest is the input for email verification.
type VerifyEmailRequest struct {
	Token string // User ID used as token for now
}

// Handle executes the verify email command.
func (c *VerifyEmailCommand) Handle(ctx context.Context, req VerifyEmailRequest) error {
	const op = "VerifyEmailCommand.Handle"

	// Parse token as user ID (simplified - in production use JWT or unique tokens)
	userID, err := types.ParseID(req.Token)
	if err != nil {
		return pkgerrors.Validation(op, "invalid verification token")
	}

	// Find user
	u, err := c.repo.FindByID(ctx, userID)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			return pkgerrors.NotFound(op, "user not found")
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	// Verify email
	if err := u.VerifyEmail(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Save
	if err := c.repo.Save(ctx, u); err != nil {
		return fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	c.logger.Info("email verified",
		logger.String("user_id", u.ID().String()),
		logger.String("email", u.Email().String()),
	)

	// Publish domain events
	if err := c.publishEvents(ctx, u); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
	}

	return nil
}

// publishEvents publishes all domain events from the aggregate.
func (c *VerifyEmailCommand) publishEvents(ctx context.Context, u *user.User) error {
	events := u.Events()
	defer u.ClearEvents()

	for _, domainEvent := range events {
		event := messaging.NewEventFromContext(
			ctx,
			domainEvent.Type(),
			"identity",
			domainEvent.Payload(),
		).
			WithMetadata("aggregate_id", domainEvent.AggregateID()).
			WithMetadata("aggregate_type", "user").
			WithMetadata("event_version", domainEvent.Version())

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
