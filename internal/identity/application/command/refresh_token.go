package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/auth"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/session"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
)

// RefreshTokenCommand handles token refresh.
type RefreshTokenCommand struct {
	userRepo       user.Repository
	sessionRepo    session.Repository
	jwtService     jwt.Service
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewRefreshTokenCommand creates a new RefreshTokenCommand.
func NewRefreshTokenCommand(
	userRepo user.Repository,
	sessionRepo session.Repository,
	jwtService jwt.Service,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *RefreshTokenCommand {
	return &RefreshTokenCommand{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		jwtService:     jwtService,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// RefreshTokenRequest is the input for token refresh.
type RefreshTokenRequest struct {
	RefreshToken string
}

// RefreshTokenResponse is the output for token refresh.
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Handle executes the refresh token command.
func (c *RefreshTokenCommand) Handle(ctx context.Context, req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	const op = "RefreshTokenCommand.Handle"

	// Find session by refresh token
	sess, err := c.sessionRepo.FindByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			return nil, pkgerrors.Unauthorized(op, "invalid refresh token")
		}
		return nil, fmt.Errorf("%s: failed to find session: %w", op, err)
	}

	// Validate session is active
	if !sess.IsActive() {
		return nil, pkgerrors.Unauthorized(op, "session is not active")
	}

	// Get user (to verify account is still active)
	u, err := c.userRepo.FindByID(ctx, sess.UserID())
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			return nil, pkgerrors.Unauthorized(op, "user not found")
		}
		return nil, fmt.Errorf("%s: failed to find user: %w", op, err)
	}

	// Verify user can still login
	if u.Status() != user.StatusActive {
		return nil, pkgerrors.Unauthorized(op, "user account is not active")
	}

	// Create new refresh token (token rotation)
	newRefreshToken, err := auth.NewToken(
		auth.TokenTypeRefresh,
		u.ID(),
		u.TenantID(),
		auth.RefreshTokenTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create refresh token: %w", op, err)
	}

	// Refresh session with new token
	if err := sess.Refresh(newRefreshToken.Value(), auth.RefreshTokenTTL); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Save session
	if err := c.sessionRepo.Save(ctx, sess); err != nil {
		return nil, fmt.Errorf("%s: failed to save session: %w", op, err)
	}

	// Generate new JWT access token
	claims := auth.NewClaims(
		u.ID(),
		u.TenantID(),
		u.Email().String(),
		u.Role().String(),
		sess.Provider(),
		sess.ID(),
	)

	accessToken, err := c.jwtService.Generate(ctx, claims.ToJWT())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to generate access token: %w", op, err)
	}

	c.logger.Info("token refreshed",
		logger.String("user_id", u.ID().String()),
		logger.String("session_id", sess.ID().String()),
	)

	// Publish domain events
	events := messaging.AsDomainEvents(sess.Events())
	defer sess.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "identity", "session", events); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("session_id", sess.ID().String()),
			logger.Err(err),
		)
	}

	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken.Value(),
		ExpiresIn:    int(auth.AccessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}
