package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/auth"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/internal/identity/infrastructure/repository"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// RequestPasswordResetCommand handles password reset requests.
type RequestPasswordResetCommand struct {
	userRepo  user.Repository
	tokenRepo *repository.PostgresTokenRepository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewRequestPasswordResetCommand creates a new RequestPasswordResetCommand.
func NewRequestPasswordResetCommand(
	userRepo user.Repository,
	tokenRepo *repository.PostgresTokenRepository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *RequestPasswordResetCommand {
	return &RequestPasswordResetCommand{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		publisher: publisher,
		logger:    logger,
	}
}

// RequestPasswordResetRequest is the input for requesting a password reset.
type RequestPasswordResetRequest struct {
	Email     string
	IPAddress string
	UserAgent string
}

// Handle executes the request password reset command.
func (c *RequestPasswordResetCommand) Handle(ctx context.Context, req RequestPasswordResetRequest) error {
	const op = "RequestPasswordResetCommand.Handle"

	// Find user by email
	email, err := user.NewEmail(req.Email)
	if err != nil {
		return pkgerrors.Validation(op, "invalid email format")
	}

	u, err := c.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not (security best practice)
		if pkgerrors.IsNotFound(err) {
			c.logger.Warn("password reset requested for non-existent email",
				logger.String("email", req.Email),
			)
			// Return success to avoid email enumeration
			return nil
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	// Create password reset token
	resetToken, err := auth.NewToken(
		auth.TokenTypeReset,
		u.ID(),
		u.TenantID(),
		auth.PasswordResetTTL,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to create reset token: %w", op, err)
	}

	// Save token to database
	if err := c.tokenRepo.SavePasswordResetToken(ctx, resetToken, req.IPAddress, req.UserAgent); err != nil {
		return fmt.Errorf("%s: failed to save reset token: %w", op, err)
	}

	// Publish event (notification subscriber will send email)
	event := messaging.NewEventFromContext(
		ctx,
		"identity.password_reset_requested",
		"identity",
		map[string]any{
			"user_id":   u.ID().String(),
			"tenant_id": u.TenantID(),
			"email":     u.Email().String(),
			"token":     resetToken.Value(),
		},
	)
	event.WithTenantID(u.TenantID())
	event.WithUserID(u.ID().String())

	if err := c.publisher.Publish(ctx, event); err != nil {
		c.logger.Error("failed to publish password reset event",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
		// Don't fail the request if event publishing fails
	}

	c.logger.Info("password reset requested",
		logger.String("user_id", u.ID().String()),
		logger.String("email", u.Email().String()),
	)

	return nil
}
