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

// VerifyEmailCommand handles email verification.
type VerifyEmailCommand struct {
	repo      user.Repository
	tokenRepo *repository.PostgresTokenRepository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewVerifyEmailCommand creates a new VerifyEmailCommand.
func NewVerifyEmailCommand(
	repo user.Repository,
	tokenRepo *repository.PostgresTokenRepository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *VerifyEmailCommand {
	return &VerifyEmailCommand{
		repo:      repo,
		tokenRepo: tokenRepo,
		publisher: publisher,
		logger:    logger,
	}
}

// VerifyEmailRequest is the input for email verification.
type VerifyEmailRequest struct {
	Token string
}

// Handle executes the verify email command.
func (c *VerifyEmailCommand) Handle(ctx context.Context, req VerifyEmailRequest) error {
	const op = "VerifyEmailCommand.Handle"

	// Validate and find verification token
	verificationToken, err := c.tokenRepo.FindEmailVerificationToken(ctx, req.Token)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			return pkgerrors.Validation(op, "invalid or expired verification token")
		}
		return fmt.Errorf("%s: failed to find verification token: %w", op, err)
	}

	// Find user
	u, err := c.repo.FindByID(ctx, verificationToken.UserID())
	if err != nil {
		return fmt.Errorf("%s: failed to find user: %w", op, err)
	}

	// Verify email (domain logic)
	if err := u.VerifyEmail(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Save user
	if err := c.repo.Save(ctx, u); err != nil {
		return fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	// Mark token as used
	if err := c.tokenRepo.MarkEmailVerificationTokenUsed(ctx, req.Token); err != nil {
		c.logger.Error("failed to mark token as used",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
		// Don't fail - email is already verified
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
