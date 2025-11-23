//go:build wireinject
// +build wireinject

package identity

import (
	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/internal/identity/application/command"
	"github.com/0xsj/hexagonal-go/internal/identity/application/query"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/internal/identity/infrastructure/repository"
	v1 "github.com/0xsj/hexagonal-go/internal/identity/interface/http/v1"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

var IdentitySet = wire.NewSet(
	repository.NewMockUserRepository,
	wire.Bind(new(user.Repository), new(*repository.MockUserRepository)),

	command.NewRegisterUserCommand,
	command.NewLoginCommand,
	command.NewVerifyEmailCommand,

	query.NewGetUserQuery,
	query.NewListUsersQuery,

	v1.NewHandler,
)

func ProvideModule(
	db database.DB,
	log logger.Logger,
) (*v1.Handler, error) {
	wire.Build(IdentitySet)
	return &v1.Handler{}, nil
}
