package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/auth"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/session"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
)

// LoginCommand handles user login/authentication.
type LoginCommand struct {
	userRepo    user.Repository
	sessionRepo session.Repository
	jwtService  jwt.Service
	publisher   messaging.Publisher
	logger      logger.Logger
}

// NewLoginCommand creates a new LoginCommand.
func NewLoginCommand(
	userRepo user.Repository,
	sessionRepo session.Repository,
	jwtService jwt.Service,
	publisher messaging.Publisher,
	logger logger.Logger,
) *LoginCommand {
	return &LoginCommand{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtService:  jwtService,
		publisher:   publisher,
		logger:      logger,
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
	u, err := c.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			return nil, user.ErrInvalidCredentials(op)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Authenticate
	if err := u.Authenticate(req.Password, req.IPAddress, req.UserAgent); err != nil {
		// Save failed attempt
		if saveErr := c.userRepo.Save(ctx, u); saveErr != nil {
			c.logger.Error("failed to save user after failed login",
				logger.String("user_id", u.ID().String()),
				logger.Err(saveErr),
			)
		}

		// Publish failed login event
		c.publishUserEvents(ctx, u)

		return nil, err
	}

	// Save successful login
	if err := c.userRepo.Save(ctx, u); err != nil {
		return nil, fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	// Create refresh token
	refreshToken, err := auth.NewToken(
		auth.TokenTypeRefresh,
		u.ID(),
		u.TenantID(),
		auth.RefreshTokenTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create refresh token: %w", op, err)
	}

	// Create session
	sess := session.New(
		u.ID(),
		u.TenantID(),
		auth.ProviderPassword,
		refreshToken.Value(),
		req.IPAddress,
		req.UserAgent,
		auth.RefreshTokenTTL,
	)

	// Save session
	if err := c.sessionRepo.Save(ctx, sess); err != nil {
		return nil, fmt.Errorf("%s: failed to save session: %w", op, err)
	}

	// Generate JWT access token
	claims := auth.NewClaims(
		u.ID(),
		u.TenantID(),
		u.Email().String(),
		u.Role().String(),
		auth.ProviderPassword,
		sess.ID(),
	)

	accessToken, err := c.jwtService.Generate(ctx, claims.ToJWT())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to generate access token: %w", op, err)
	}

	c.logger.Info("user logged in",
		logger.String("user_id", u.ID().String()),
		logger.String("email", u.Email().String()),
		logger.String("session_id", sess.ID().String()),
	)

	// Publish events
	c.publishUserEvents(ctx, u)
	c.publishSessionEvents(ctx, sess)

	// Return response
	return &dto.LoginResponse{
		User:         dto.NewUserResponse(u),
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Value(),
		ExpiresIn:    int(auth.AccessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// publishUserEvents publishes all domain events from the user aggregate.
func (c *LoginCommand) publishUserEvents(ctx context.Context, u *user.User) error {
	events := u.Events()
	defer u.ClearEvents()

	for _, domainEvent := range events {
		event := c.convertUserEvent(ctx, domainEvent)

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

// publishSessionEvents publishes all domain events from the session aggregate.
func (c *LoginCommand) publishSessionEvents(ctx context.Context, sess *session.Session) error {
	events := sess.Events()
	defer sess.ClearEvents()

	for _, domainEvent := range events {
		event := c.convertSessionEvent(ctx, domainEvent)

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

// convertUserEvent converts a user domain event to a messaging event.
func (c *LoginCommand) convertUserEvent(ctx context.Context, domainEvent user.Event) *messaging.BaseEvent {
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

	return event
}

// convertSessionEvent converts a session domain event to a messaging event.
func (c *LoginCommand) convertSessionEvent(ctx context.Context, domainEvent session.Event) *messaging.BaseEvent {
	event := messaging.NewEventFromContext(
		ctx,
		"identity."+domainEvent.Type(),
		"identity",
		domainEvent.Payload(),
	)

	// Add standard metadata
	event.WithTenantID(domainEvent.AggregateTenantID())

	// Add aggregate metadata
	event.WithMetadata("aggregate_id", domainEvent.AggregateID().String())
	event.WithMetadata("aggregate_type", "session")

	return event
}
