package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/auth"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/internal/identity/infrastructure/repository"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// RegisterUserCommand handles user registration.
type RegisterUserCommand struct {
	repo           user.Repository
	tokenRepo      *repository.PostgresTokenRepository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewRegisterUserCommand creates a new RegisterUserCommand.
func NewRegisterUserCommand(
	repo user.Repository,
	tokenRepo *repository.PostgresTokenRepository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *RegisterUserCommand {
	return &RegisterUserCommand{
		repo:           repo,
		tokenRepo:      tokenRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
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

	// Create email verification token
	var verificationToken *auth.Token
	verificationToken, err = auth.NewToken(
		auth.TokenTypeVerification,
		u.ID(),
		u.TenantID(),
		auth.EmailVerificationTTL,
	)
	if err != nil {
		c.logger.Error("failed to create verification token",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
		// Don't fail registration - user can request new token
		verificationToken = nil
	} else {
		// Save verification token
		if err := c.tokenRepo.SaveEmailVerificationToken(ctx, verificationToken, "", ""); err != nil {
			c.logger.Error("failed to save verification token",
				logger.String("user_id", u.ID().String()),
				logger.Err(err),
			)
			// Don't fail registration
			verificationToken = nil
		} else {
			c.logger.Debug("verification token created",
				logger.String("user_id", u.ID().String()),
			)
		}
	}

	c.logger.Info("user registered",
		logger.String("user_id", u.ID().String()),
		logger.String("email", u.Email().String()),
		logger.String("tenant_id", u.TenantID()),
	)

	// Publish domain events
	events := messaging.AsDomainEvents(u.Events())
	defer u.ClearEvents()

	// Build publish options with verification token if available
	var opts []messaging.PublishOption
	if verificationToken != nil {
		opts = append(opts, messaging.WithEventMetadata("user.registered", map[string]any{
			"verification_token": verificationToken.Value(),
		}))
	}

	if err := c.eventPublisher.PublishAll(ctx, "identity", "user", events, opts...); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
	}

	return dto.NewUserResponse(u), nil
}
