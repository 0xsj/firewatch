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
	userRepo       user.Repository
	tokenRepo      *repository.PostgresTokenRepository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewResetPasswordCommand creates a new ResetPasswordCommand.
func NewResetPasswordCommand(
	userRepo user.Repository,
	tokenRepo *repository.PostgresTokenRepository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *ResetPasswordCommand {
	return &ResetPasswordCommand{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
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
