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
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// IdentitySet provides all dependencies for the Identity domain.
var IdentitySet = wire.NewSet(
	// Infrastructure - User Repository
	repository.NewPostgresUserRepository,
	wire.Bind(new(user.Repository), new(*repository.PostgresUserRepository)),

	// Infrastructure - Session Repository
	repository.NewPostgresSessionRepository,
	wire.Bind(new(session.Repository), new(*repository.PostgresSessionRepository)),

	// Application - Commands
	command.NewRegisterUserCommand,
	command.NewLoginCommand,
	command.NewVerifyEmailCommand,

	// Application - Queries
	query.NewGetUserQuery,
	query.NewListUsersQuery,

	// Interface - HTTP Handler
	v1.NewHandler,
)

// ProvideModule wires up the complete Identity module.
func ProvideModule(
	db database.DB,
	publisher messaging.Publisher,
	log logger.Logger,
) (*v1.Handler, error) {
	wire.Build(IdentitySet)
	return &v1.Handler{}, nil
}
