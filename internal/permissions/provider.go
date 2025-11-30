//go:build wireinject
// +build wireinject

package permissions

import (
	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/internal/permissions/application/command"
	"github.com/0xsj/hexagonal-go/internal/permissions/application/query"
	"github.com/0xsj/hexagonal-go/internal/permissions/domain"
	"github.com/0xsj/hexagonal-go/internal/permissions/infrastructure/repository"
	v1 "github.com/0xsj/hexagonal-go/internal/permissions/interface/http/v1"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// PermissionsSet provides all dependencies for the Permissions domain.
var PermissionsSet = wire.NewSet(
	// Infrastructure - Role Repository
	repository.NewPostgresRoleRepository,
	wire.Bind(new(domain.RoleRepository), new(*repository.PostgresRoleRepository)),

	// Infrastructure - Assignment Repository
	repository.NewPostgresAssignmentRepository,
	wire.Bind(new(domain.AssignmentRepository), new(*repository.PostgresAssignmentRepository)),

	// Application - Commands
	command.NewCreateRoleCommand,
	command.NewUpdateRoleCommand,
	command.NewDeleteRoleCommand,
	command.NewAssignRoleCommand,
	command.NewRevokeRoleCommand,

	// Application - Queries
	query.NewGetRoleQuery,
	query.NewListRolesQuery,
	query.NewGetUserPermissionsQuery,
	query.NewCheckPermissionQuery,

	// Interface - HTTP Handler
	v1.NewHandler,
)

// ProvideModule wires up the complete Permissions module.
func ProvideModule(
	db database.DB,
	eventPublisher *messaging.DomainEventPublisher,
	log logger.Logger,
) (*v1.Handler, error) {
	wire.Build(PermissionsSet)
	return &v1.Handler{}, nil
}
