package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// LoginCommand handles user login/authentication.
type LoginCommand struct {
	repo      user.Repository
	publisher messaging.Publisher
	logger    logger.Logger
}

// NewLoginCommand creates a new LoginCommand.
func NewLoginCommand(
	repo user.Repository,
	publisher messaging.Publisher,
	logger logger.Logger,
) *LoginCommand {
	return &LoginCommand{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// LoginRequest is the input for login.
type LoginRequest struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

// Handle executes the login command.
func (c *LoginCommand) Handle(ctx context.Context, req LoginRequest) (*dto.LoginResponse, error) {
	const op = "LoginCommand.Handle"

	// Parse email
	email, err := user.NewEmail(req.Email)
	if err != nil {
		return nil, pkgerrors.Validation(op, "invalid email format")
	}

	// Find user by email
	u, err := c.repo.FindByEmail(ctx, email)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			return nil, user.ErrInvalidCredentials(op)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Authenticate
	if err := u.Authenticate(req.Password, req.IPAddress, req.UserAgent); err != nil {
		// Save failed attempt
		if saveErr := c.repo.Save(ctx, u); saveErr != nil {
			c.logger.Error("failed to save user after failed login",
				logger.String("user_id", u.ID().String()),
				logger.Err(saveErr),
			)
		}

		// Publish failed login event
		c.publishEvents(ctx, u)

		return nil, err
	}

	// Save successful login
	if err := c.repo.Save(ctx, u); err != nil {
		return nil, fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	c.logger.Info("user logged in",
		logger.String("user_id", u.ID().String()),
		logger.String("email", u.Email().String()),
	)

	// Publish login event
	if err := c.publishEvents(ctx, u); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
	}

	// Return response with placeholder tokens
	return &dto.LoginResponse{
		User:         dto.NewUserResponse(u),
		AccessToken:  "jwt-access-token-placeholder",
		RefreshToken: "jwt-refresh-token-placeholder",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}, nil
}

// publishEvents publishes all domain events from the aggregate.
func (c *LoginCommand) publishEvents(ctx context.Context, u *user.User) error {
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
