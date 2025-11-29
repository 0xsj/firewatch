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
	repo           user.Repository
	tokenRepo      *repository.PostgresTokenRepository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewVerifyEmailCommand creates a new VerifyEmailCommand.
func NewVerifyEmailCommand(
	repo user.Repository,
	tokenRepo *repository.PostgresTokenRepository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *VerifyEmailCommand {
	return &VerifyEmailCommand{
		repo:           repo,
		tokenRepo:      tokenRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
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
