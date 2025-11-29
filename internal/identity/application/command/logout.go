package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/session"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// LogoutCommand handles user logout.
type LogoutCommand struct {
	sessionRepo    session.Repository
	jwtService     jwt.Service
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewLogoutCommand creates a new LogoutCommand.
func NewLogoutCommand(
	sessionRepo session.Repository,
	jwtService jwt.Service,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *LogoutCommand {
	return &LogoutCommand{
		sessionRepo:    sessionRepo,
		jwtService:     jwtService,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// LogoutRequest is the input for logout.
type LogoutRequest struct {
	SessionID   types.ID // From JWT claims
	AccessToken string   // To blacklist
}

// Handle executes the logout command.
func (c *LogoutCommand) Handle(ctx context.Context, req LogoutRequest) error {
	const op = "LogoutCommand.Handle"

	// Find session
	sess, err := c.sessionRepo.FindByID(ctx, req.SessionID)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			// Session already gone - that's fine
			c.logger.Warn("session not found during logout",
				logger.String("session_id", req.SessionID.String()),
			)
			return nil
		}
		return fmt.Errorf("%s: failed to find session: %w", op, err)
	}

	// Logout session (domain logic)
	if err := sess.Logout(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Save session
	if err := c.sessionRepo.Save(ctx, sess); err != nil {
		return fmt.Errorf("%s: failed to save session: %w", op, err)
	}

	// Invalidate JWT (add to blacklist)
	if req.AccessToken != "" {
		if err := c.jwtService.Invalidate(ctx, req.AccessToken); err != nil {
			c.logger.Error("failed to invalidate JWT",
				logger.String("session_id", req.SessionID.String()),
				logger.Err(err),
			)
			// Don't fail logout if blacklist fails - session is already logged out
		}
	}

	c.logger.Info("user logged out",
		logger.String("session_id", sess.ID().String()),
		logger.String("user_id", sess.UserID().String()),
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

	return nil
}
