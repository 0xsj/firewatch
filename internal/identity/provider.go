//go:build wireinject
// +build wireinject

package identity

import (
	"os"

	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/internal/identity/application/command"
	"github.com/0xsj/hexagonal-go/internal/identity/application/query"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/oauth"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/session"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/internal/identity/infrastructure/repository"
	v1 "github.com/0xsj/hexagonal-go/internal/identity/interface/http/v1"
	"github.com/0xsj/hexagonal-go/pkg/cache"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	pkgoauth "github.com/0xsj/hexagonal-go/pkg/oauth"
	"github.com/0xsj/hexagonal-go/pkg/oauth/github"
	"github.com/0xsj/hexagonal-go/pkg/oauth/google"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
)

// IdentitySet provides all dependencies for the Identity domain.
var IdentitySet = wire.NewSet(
	// Infrastructure - User Repository
	repository.NewPostgresUserRepository,
	wire.Bind(new(user.Repository), new(*repository.PostgresUserRepository)),

	// Infrastructure - Session Repository (with optional caching)
	repository.NewPostgresSessionRepository,
	ProvideSessionRepository,

	// Infrastructure - Token Repository
	repository.NewPostgresTokenRepository,

	// Infrastructure - OAuth Repository
	repository.NewPostgresOAuthRepository,
	wire.Bind(new(oauth.Repository), new(*repository.PostgresOAuthRepository)),

	// OAuth Providers
	ProvideOAuthProviders,
	ProvideStateManager,

	// Application - Commands
	command.NewRegisterUserCommand,
	command.NewLoginCommand,
	command.NewLogoutCommand,
	command.NewRefreshTokenCommand,
	command.NewVerifyEmailCommand,
	command.NewRequestPasswordResetCommand,
	command.NewResetPasswordCommand,
	command.NewChangePasswordCommand,
	command.NewSuspendUserCommand,
	command.NewReactivateUserCommand,
	command.NewChangeUserRoleCommand,
	command.NewDeleteUserCommand,
	command.NewOAuthLoginCommand,

	// Application - Queries
	query.NewGetUserQuery,
	query.NewGetCurrentUserQuery,
	query.NewListUsersQuery,
	query.NewListSessionsQuery,

	// Interface - OAuth Handler
	v1.NewOAuthHandler,

	// Interface - HTTP Handler
	v1.NewHandler,
)

// ProvideSessionRepository provides a session repository with optional caching.
func ProvideSessionRepository(
	postgresRepo *repository.PostgresSessionRepository,
	cache cache.Cache,
) session.Repository {
	if cache == nil {
		return postgresRepo
	}
	return repository.NewCachedSessionRepository(postgresRepo, cache)
}

// ProvideOAuthProviders initializes OAuth providers based on environment configuration.
func ProvideOAuthProviders(log logger.Logger) map[string]pkgoauth.Provider {
	providers := make(map[string]pkgoauth.Provider)

	// Google OAuth
	if googleClientID := os.Getenv("GOOGLE_CLIENT_ID"); googleClientID != "" {
		googleProvider, err := google.NewProvider(pkgoauth.Config{
			ClientID:     googleClientID,
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		})
		if err != nil {
			log.Warn("failed to initialize Google OAuth provider", logger.Err(err))
		} else {
			providers["google"] = googleProvider
			log.Info("Google OAuth provider initialized")
		}
	}

	// GitHub OAuth
	if githubClientID := os.Getenv("GITHUB_CLIENT_ID"); githubClientID != "" {
		githubProvider, err := github.NewProvider(pkgoauth.Config{
			ClientID:     githubClientID,
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
		})
		if err != nil {
			log.Warn("failed to initialize GitHub OAuth provider", logger.Err(err))
		} else {
			providers["github"] = githubProvider
			log.Info("GitHub OAuth provider initialized")
		}
	}

	return providers
}

// ProvideStateManager provides an OAuth state manager.
func ProvideStateManager() *pkgoauth.StateManager {
	return pkgoauth.NewStateManager()
}

// ProvideModule wires up the complete Identity module.
func ProvideModule(
	db database.DB,
	publisher messaging.Publisher,
	jwtService jwt.Service,
	cache cache.Cache,
	log logger.Logger,
) (*v1.Handler, error) {
	wire.Build(IdentitySet)
	return &v1.Handler{}, nil
}
