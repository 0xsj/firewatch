package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// RegisterUserCommand handles user registration with password authentication.
// This is an application service that orchestrates the registration use case.
//
// Responsibilities:
//   - Validate input
//   - Check business rules (email uniqueness)
//   - Create domain aggregate
//   - Persist via repository
//   - Return DTO for API response
//
// Does NOT:
//   - Know about HTTP, JSON, or request/response details
//   - Contain domain logic (that's in User aggregate)
//   - Know about database implementation (uses repository interface)
type RegisterUserCommand struct {
	repo user.Repository
}

// NewRegisterUserCommand creates a new register user command.
func NewRegisterUserCommand(repo user.Repository) *RegisterUserCommand {
	return &RegisterUserCommand{
		repo: repo,
	}
}

// Handle executes the user registration command.
func (c *RegisterUserCommand) Handle(ctx context.Context, req dto.RegisterUserRequest) (*dto.UserDTO, error) {
	const op = "RegisterUserCommand.Handle"

	// 1. Parse and validate email
	email, err := user.NewEmail(req.Email)
	if err != nil {
		return nil, user.ErrEmailInvalid(op, req.Email)
	}

	// 2. Validate email for registration (checks disposable domains, etc.)
	if err := email.ValidateForRegistration(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// 3. Check email uniqueness within tenant
	// The repository automatically filters by tenant_id from context
	exists, err := c.repo.EmailExists(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to check email existence: %w", op, err)
	}
	if exists {
		return nil, user.ErrEmailAlreadyTaken(op, email.String())
	}

	// 4. Create and validate password
	password, err := user.NewPassword(req.Password, user.DefaultPasswordRequirements())
	if err != nil {
		return nil, user.ErrPasswordTooWeak(op, err.Error())
	}

	// 5. Create user aggregate
	// This is where domain logic lives - the User.Register factory method
	// enforces all business rules and emits domain events
	u, err := user.Register(
		types.NewID(),
		req.TenantID,
		email,
		password,
		user.DefaultRole(), // New users get "user" role by default
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create user: %w", op, err)
	}

	// 6. Persist the user
	// Repository handles database-specific details
	if err := c.repo.Save(ctx, u); err != nil {
		return nil, fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	// 7. TODO: Publish domain events (when we add messaging)
	// This would notify other services about the user registration
	// for _, event := range u.Events() {
	//     if err := c.publisher.Publish(ctx, event); err != nil {
	//         // Log but don't fail - events are eventually consistent
	//         logger.Warn("failed to publish event", logger.Err(err))
	//     }
	// }
	// u.ClearEvents()

	// 8. TODO: Send verification email (when we add email service)
	// token := generateVerificationToken(u.ID())
	// if err := c.emailService.SendVerificationEmail(ctx, u.Email(), token); err != nil {
	//     // Log but don't fail - user can resend verification
	//     logger.Warn("failed to send verification email", logger.Err(err))
	// }

	// 9. Map domain aggregate to DTO and return
	return dto.MapUserToDTO(u), nil
}
