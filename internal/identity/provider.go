//go:build wireinject
// +build wireinject

package identity

import (
	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/internal/identity/application/command"
	"github.com/0xsj/hexagonal-go/internal/identity/application/query"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/session"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/internal/identity/infrastructure/repository"
	v1 "github.com/0xsj/hexagonal-go/internal/identity/interface/http/v1"
	"github.com/0xsj/hexagonal-go/pkg/cache"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
)

// IdentitySet provides all dependencies for the Identity domain.
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

	// Application - Commands
	command.NewRegisterUserCommand,
	command.NewLoginCommand,
	command.NewLogoutCommand,
	command.NewRefreshTokenCommand,
	command.NewVerifyEmailCommand,
	command.NewRequestPasswordResetCommand,
	command.NewResetPasswordCommand,

	// Application - Queries
	query.NewGetUserQuery,
	query.NewGetCurrentUserQuery,
	query.NewListUsersQuery,
	query.NewListSessionsQuery,

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
