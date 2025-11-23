package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// VerifyEmailCommand handles email verification for user accounts.
// Activates pending accounts after email is verified.
//
// Responsibilities:
//   - Validate verification token
//   - Find user by ID (from token)
//   - Mark email as verified in domain
//   - Activate account if pending
//   - Persist changes
//
// Note: Token validation logic will be added when we build the token service.
// For now, this assumes a valid user ID is extracted from the token.
type VerifyEmailCommand struct {
	repo user.Repository
}

// NewVerifyEmailCommand creates a new verify email command.
func NewVerifyEmailCommand(repo user.Repository) *VerifyEmailCommand {
	return &VerifyEmailCommand{
		repo: repo,
	}
}

// Handle executes the email verification command.
func (c *VerifyEmailCommand) Handle(ctx context.Context, req dto.VerifyEmailRequest) (*dto.MessageResponse, error) {
	const op = "VerifyEmailCommand.Handle"

	// 1. TODO: Validate and parse verification token (when we build token service)
	// For now, we'll accept a user ID directly as the "token"
	// In production, this would:
	//   - Verify JWT signature
	//   - Check expiration (24-48 hours)
	//   - Extract user ID from token payload
	//   - Verify token hasn't been used already
	//
	// userID, err := c.tokenService.ValidateVerificationToken(req.Token)
	// if err != nil {
	//     return nil, user.ErrEmailVerificationExpired(op, "")
	// }

	// Temporary: parse token as user ID directly
	userID, err := types.ParseID(req.Token)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid verification token: %w", op, err)
	}

	// 2. Find user by ID
	u, err := c.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, user.ErrUserNotFound(op, userID.String())
	}

	// 3. Verify email in domain
	// This method:
	//   - Marks email as verified
	//   - Sets email_verified_at timestamp
	//   - Activates account if status is pending
	//   - Emits UserEmailVerified event
	//   - Is idempotent (safe to call multiple times)
	if err := u.VerifyEmail(); err != nil {
		return nil, fmt.Errorf("%s: failed to verify email: %w", op, err)
	}

	// 4. Persist changes
	if err := c.repo.Save(ctx, u); err != nil {
		return nil, fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	// 5. TODO: Publish domain events (when we build messaging)
	// for _, event := range u.Events() {
	//     c.publisher.Publish(ctx, event)
	// }
	// u.ClearEvents()

	// 6. TODO: Send welcome email (when we build email service)
	// if u.EmailVerified() {
	//     c.emailService.SendWelcomeEmail(ctx, u.Email())
	// }

	// 7. Return success message
	return &dto.MessageResponse{
		Message: "Email verified successfully. Your account is now active.",
	}, nil
}
