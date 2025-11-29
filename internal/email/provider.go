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

	// Infrastructure - Template Adapter (for pkg/email.TemplateRepository)
	repository.NewTemplateRepositoryAdapter,
	wire.Bind(new(email.TemplateRepository), new(*repository.TemplateRepositoryAdapter)),

	// Renderer
	ProvideRenderer,

	// Template Service
	email.NewTemplateService,

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
	eventPublisher *messaging.DomainEventPublisher,
	log logger.Logger,
) (*v1.Handler, error) {
	wire.Build(EmailSet)
	return &v1.Handler{}, nil
}

// ProvideTemplateService wires up the TemplateService for use by other modules.
func ProvideTemplateService(
	db database.DB,
) (*email.TemplateService, error) {
	wire.Build(
		repository.NewPostgresRepository,
		repository.NewTemplateRepositoryAdapter,
		wire.Bind(new(email.TemplateRepository), new(*repository.TemplateRepositoryAdapter)),
		ProvideRenderer,
		email.NewTemplateService,
	)
	return &email.TemplateService{}, nil
}
