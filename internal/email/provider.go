//go:build wireinject
// +build wireinject

package email

import (
	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/internal/email/application/command"
	"github.com/0xsj/hexagonal-go/internal/email/application/query"
	"github.com/0xsj/hexagonal-go/internal/email/domain"
	"github.com/0xsj/hexagonal-go/internal/email/infrastructure/repository"
	v1 "github.com/0xsj/hexagonal-go/internal/email/interface/http/v1"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/email"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// EmailSet provides all dependencies for the Email domain.
var EmailSet = wire.NewSet(
	// Infrastructure - Repository
	repository.NewPostgresRepository,
	wire.Bind(new(domain.Repository), new(*repository.PostgresRepository)),

	// Renderer
	ProvideRenderer,

	// Application - Commands
	command.NewCreateTemplateCommand,
	command.NewUpdateTemplateCommand,
	command.NewActivateTemplateCommand,
	command.NewArchiveTemplateCommand,
	command.NewDeleteTemplateCommand,

	// Application - Queries
	query.NewGetTemplateQuery,
	query.NewListTemplatesQuery,
	query.NewPreviewTemplateQuery,

	// Interface - HTTP Handler
	v1.NewHandler,
)

// ProvideRenderer provides an email template renderer.
func ProvideRenderer() *email.Renderer {
	return email.NewRenderer()
}

// ProvideModule wires up the complete Email module.
func ProvideModule(
	db database.DB,
	publisher messaging.Publisher,
	log logger.Logger,
) (*v1.Handler, error) {
	wire.Build(EmailSet)
	return &v1.Handler{}, nil
}
