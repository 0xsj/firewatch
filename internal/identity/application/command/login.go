package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
)

// LoginCommand handles user login with password authentication.
// This command authenticates a user and returns access tokens.
//
// Responsibilities:
//   - Find user by email
//   - Validate credentials via domain logic
//   - Update login state (last_login_at, failed attempts)
//   - Generate authentication tokens
//   - Return login response with user info and tokens
type LoginCommand struct {
	repo user.Repository
}

// NewLoginCommand creates a new login command.
func NewLoginCommand(repo user.Repository) *LoginCommand {
	return &LoginCommand{
		repo: repo,
	}
}

// Handle executes the login command.
func (c *LoginCommand) Handle(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	const op = "LoginCommand.Handle"

	// 1. Parse and validate email
	email, err := user.NewEmail(req.Email)
	if err != nil {
		// Don't reveal if email format is invalid - use generic credentials error
		return nil, user.ErrInvalidCredentials(op)
	}

	// 2. Find user by email
	// Note: We don't reveal if the user exists or not (security best practice)
	u, err := c.repo.FindByEmail(ctx, email)
	if err != nil {
		// Whether user not found or other error, return generic credentials error
		return nil, user.ErrInvalidCredentials(op)
	}

	// 3. Authenticate user
	// This method in the domain:
	//   - Checks account status (active, locked, suspended)
	//   - Verifies password
	//   - Tracks failed login attempts
	//   - Locks account if too many failures
	//   - Updates last_login_at on success
	if err := u.Authenticate(req.Password, req.IPAddress, req.UserAgent); err != nil {
		// Save the user even on failure (to persist failed attempt count)
		_ = c.repo.Save(ctx, u)
		return nil, err
	}

	// 4. Save successful login state
	// This persists:
	//   - Updated last_login_at
	//   - Reset failed_login_attempts to 0
	//   - Domain events (UserLoggedIn)
	if err := c.repo.Save(ctx, u); err != nil {
		return nil, fmt.Errorf("%s: failed to save login state: %w", op, err)
	}

	// 5. TODO: Generate JWT tokens (when we build auth package)
	// For now, return placeholder tokens
	// accessToken, err := c.jwtService.GenerateAccessToken(u)
	// refreshToken, err := c.jwtService.GenerateRefreshToken(u)
	accessToken := "jwt-access-token-placeholder"
	refreshToken := "jwt-refresh-token-placeholder"
	expiresIn := 3600 // 1 hour in seconds

	// 6. TODO: Publish domain events (when we build messaging)
	// for _, event := range u.Events() {
	//     c.publisher.Publish(ctx, event)
	// }
	// u.ClearEvents()

	// 7. Build and return login response
	return &dto.LoginResponse{
		User:         dto.MapUserToDTO(u),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}, nil
}
