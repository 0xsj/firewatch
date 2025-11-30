//go:build wireinject
// +build wireinject

package flags

import (
	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/internal/flags/application/command"
	"github.com/0xsj/hexagonal-go/internal/flags/application/query"
	"github.com/0xsj/hexagonal-go/internal/flags/domain"
	"github.com/0xsj/hexagonal-go/internal/flags/infrastructure/repository"
	"github.com/0xsj/hexagonal-go/internal/flags/interface/http/admin"
	v1 "github.com/0xsj/hexagonal-go/internal/flags/interface/http/v1"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// FlagsSet provides all dependencies for the Flags domain.
var FlagsSet = wire.NewSet(
	// Infrastructure - Repository
	repository.NewPostgresRepository,
	wire.Bind(new(domain.Repository), new(*repository.PostgresRepository)),

	// Application - Commands
	command.NewCreateFlagCommand,
	command.NewUpdateFlagCommand,
	command.NewDeleteFlagCommand,
	command.NewEnableFlagCommand,
	command.NewDisableFlagCommand,
	command.NewAddRuleCommand,
	command.NewRemoveRuleCommand,
	command.NewSetOverrideCommand,
	command.NewRemoveOverrideCommand,

	// Application - Queries
	query.NewGetFlagQuery,
	query.NewListFlagsQuery,
	query.NewEvaluateFlagQuery,

	// Interface - HTTP Handler (API)
	v1.NewHandler,

	// Interface - HTTP Handler (Admin Dashboard)
	admin.NewHandler,
)

// ProvideModule wires up the complete Flags module.
func ProvideModule(
	db database.DB,
	eventPublisher *messaging.DomainEventPublisher,
	log logger.Logger,
) (*v1.Handler, error) {
	wire.Build(FlagsSet)
	return &v1.Handler{}, nil
}
