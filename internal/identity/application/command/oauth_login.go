package command

import (
	"context"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/auth"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/oauth"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/session"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	pkgoauth "github.com/0xsj/hexagonal-go/pkg/oauth"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// OAuthLoginCommand handles OAuth login (creates user if needed).
type OAuthLoginCommand struct {
	userRepo       user.Repository
	oauthRepo      oauth.Repository
	sessionRepo    session.Repository
	oauthProviders map[string]pkgoauth.Provider
	jwtService     jwt.Service
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewOAuthLoginCommand creates a new OAuthLoginCommand.
func NewOAuthLoginCommand(
	userRepo user.Repository,
	oauthRepo oauth.Repository,
	sessionRepo session.Repository,
	oauthProviders map[string]pkgoauth.Provider,
	jwtService jwt.Service,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *OAuthLoginCommand {
	return &OAuthLoginCommand{
		userRepo:       userRepo,
		oauthRepo:      oauthRepo,
		sessionRepo:    sessionRepo,
		oauthProviders: oauthProviders,
		jwtService:     jwtService,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// OAuthLoginRequest is the input for OAuth login.
type OAuthLoginRequest struct {
	Provider  string
	Code      string
	TenantID  string // Required for new users
	IPAddress string
	UserAgent string
}

// Handle executes the OAuth login command.
func (c *OAuthLoginCommand) Handle(ctx context.Context, req OAuthLoginRequest) (*dto.LoginResponse, error) {
	const op = "OAuthLoginCommand.Handle"

	// Get OAuth provider
	provider, exists := c.oauthProviders[req.Provider]
	if !exists {
		return nil, pkgerrors.Validation(op, fmt.Sprintf("unsupported OAuth provider: %s", req.Provider))
	}

	// Exchange code for tokens and user info
	userInfo, tokens, err := provider.Exchange(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to exchange OAuth code: %w", op, err)
	}

	// Parse provider enum
	oauthProvider, err := oauth.ParseProvider(req.Provider)
	if err != nil {
		return nil, pkgerrors.Validation(op, err.Error())
	}

	// Parse auth provider
	authProvider, err := auth.ParseProvider(req.Provider)
	if err != nil {
		return nil, pkgerrors.Validation(op, err.Error())
	}

	// Find existing OAuth account
	oauthAccount, err := c.oauthRepo.FindByProviderUserID(ctx, oauthProvider, userInfo.ID)
	if err != nil && !pkgerrors.IsNotFound(err) {
		return nil, fmt.Errorf("%s: failed to find OAuth account: %w", op, err)
	}

	var u *user.User

	if oauthAccount != nil {
		// Existing OAuth account - load user
		u, err = c.userRepo.FindByID(ctx, oauthAccount.UserID())
		if err != nil {
			return nil, fmt.Errorf("%s: failed to find user: %w", op, err)
		}

		// Update OAuth account tokens
		oauthAccount.UpdateTokens(
			tokens.AccessToken,
			tokens.RefreshToken,
			timeFromUnix(tokens.ExpiresAt),
		)
		oauthAccount.RecordUsage()

		if err := c.oauthRepo.Save(ctx, oauthAccount); err != nil {
			c.logger.Error("failed to update OAuth account",
				logger.String("user_id", u.ID().String()),
				logger.Err(err),
			)
			// Don't fail login
		}
	} else {
		// New OAuth user - check if email exists
		email, err := user.NewEmail(userInfo.Email)
		if err != nil {
			return nil, pkgerrors.Validation(op, "invalid email from OAuth provider")
		}

		// Try to find existing user by email
		u, err = c.userRepo.FindByEmail(ctx, email)
		if err != nil && !pkgerrors.IsNotFound(err) {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if u == nil {
			// Create new user (passwordless)
			if req.TenantID == "" {
				return nil, pkgerrors.Validation(op, "tenant_id is required for new users")
			}

			u, err = user.RegisterPasswordless(
				types.NewID(),
				req.TenantID,
				email,
				user.RoleUser,
				req.Provider,
			)
			if err != nil {
				return nil, fmt.Errorf("%s: failed to create user: %w", op, err)
			}

			if err := c.userRepo.Save(ctx, u); err != nil {
				return nil, fmt.Errorf("%s: failed to save user: %w", op, err)
			}

			c.logger.Info("new OAuth user created",
				logger.String("user_id", u.ID().String()),
				logger.String("email", u.Email().String()),
				logger.String("provider", req.Provider),
			)
		}

		// Create OAuth account link
		oauthAccount, err = oauth.NewAccount(
			types.NewID(),
			u.ID(),
			u.TenantID(),
			oauthProvider,
			userInfo.ID,
			userInfo.Email,
			tokens.AccessToken,
			tokens.RefreshToken,
			timeFromUnix(tokens.ExpiresAt),
			userInfo.Raw,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to create OAuth account: %w", op, err)
		}

		if err := c.oauthRepo.Save(ctx, oauthAccount); err != nil {
			return nil, fmt.Errorf("%s: failed to save OAuth account: %w", op, err)
		}

		c.logger.Info("OAuth account linked",
			logger.String("user_id", u.ID().String()),
			logger.String("provider", req.Provider),
		)
	}

	// Authenticate user (passwordless)
	if err := u.AuthenticatePasswordless(req.Provider, req.IPAddress, req.UserAgent); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Save user (updated last login)
	if err := c.userRepo.Save(ctx, u); err != nil {
		return nil, fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	// Create session
	refreshToken := auth.GenerateRefreshToken()
	sess := session.New(
		u.ID(),
		u.TenantID(),
		authProvider,
		refreshToken,
		req.IPAddress,
		req.UserAgent,
		auth.RefreshTokenTTL,
	)

	if err := c.sessionRepo.Save(ctx, sess); err != nil {
		return nil, fmt.Errorf("%s: failed to save session: %w", op, err)
	}

	// Generate JWT tokens
	accessClaims := auth.NewClaims(
		u.ID(),
		u.TenantID(),
		u.Email().String(),
		string(u.Role()),
		authProvider,
		sess.ID(),
	)

	accessToken, err := c.jwtService.Generate(ctx, accessClaims.ToJWT())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to generate access token: %w", op, err)
	}

	c.logger.Info("OAuth login successful",
		logger.String("user_id", u.ID().String()),
		logger.String("provider", req.Provider),
		logger.String("session_id", sess.ID().String()),
	)

	// Publish user events
	c.publishUserEvents(ctx, u)

	// Publish session events
	c.publishSessionEvents(ctx, sess)

	// Build response
	return &dto.LoginResponse{
		User:         dto.NewUserResponse(u),
		AccessToken:  accessToken,
		RefreshToken: sess.RefreshToken(),
		ExpiresIn:    int(auth.AccessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// publishUserEvents publishes all domain events from the user aggregate.
func (c *OAuthLoginCommand) publishUserEvents(ctx context.Context, u *user.User) {
	events := messaging.AsDomainEvents(u.Events())
	defer u.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "identity", "user", events); err != nil {
		c.logger.Error("failed to publish user events",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
	}
}

// publishSessionEvents publishes all domain events from the session aggregate.
func (c *OAuthLoginCommand) publishSessionEvents(ctx context.Context, sess *session.Session) {
	events := messaging.AsDomainEvents(sess.Events())
	defer sess.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "identity", "session", events); err != nil {
		c.logger.Error("failed to publish session events",
			logger.String("session_id", sess.ID().String()),
			logger.Err(err),
		)
	}
}

// timeFromUnix converts Unix timestamp to *time.Time.
func timeFromUnix(timestamp int64) *time.Time {
	if timestamp == 0 {
		return nil
	}
	t := time.Unix(timestamp, 0)
	return &t
}
