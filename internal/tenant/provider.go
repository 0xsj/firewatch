//go:build wireinject
// +build wireinject

package tenant

import (
	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/internal/tenant/application/command"
	"github.com/0xsj/hexagonal-go/internal/tenant/application/query"
	tenant "github.com/0xsj/hexagonal-go/internal/tenant/domain"
	"github.com/0xsj/hexagonal-go/internal/tenant/infrastructure/repository"
	v1 "github.com/0xsj/hexagonal-go/internal/tenant/interface/http/v1"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// TenantSet provides all dependencies for the Tenant domain.
var TenantSet = wire.NewSet(
	// Infrastructure - Repository
	repository.NewPostgresRepository,
	wire.Bind(new(tenant.Repository), new(*repository.PostgresRepository)),

	// Application - Commands
	command.NewCreateTenantCommand,
	command.NewUpdateTenantCommand,
	command.NewUpdateSettingsCommand,
	command.NewSuspendTenantCommand,
	command.NewReactivateTenantCommand,
	command.NewChangePlanCommand,
	command.NewDeleteTenantCommand,

	// Application - Queries
	query.NewGetTenantQuery,
	query.NewGetTenantBySlugQuery,
	query.NewListTenantsQuery,

	// Interface - HTTP Handler
	v1.NewHandler,
)

// ProvideModule wires up the complete Tenant module.
// This is a standalone injector for testing or modular usage.
func ProvideModule(
	db database.DB,
	publisher messaging.Publisher,
	eventPublisher *messaging.DomainEventPublisher,
	log logger.Logger,
) (*v1.Handler, error) {
	wire.Build(TenantSet)
	return &v1.Handler{}, nil
}
