package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/email/application/dto"
	"github.com/0xsj/hexagonal-go/internal/email/domain"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// CreateTemplateCommand handles creating a new email template.
type CreateTemplateCommand struct {
	repo           domain.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewCreateTemplateCommand creates a new CreateTemplateCommand.
func NewCreateTemplateCommand(
	repo domain.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *CreateTemplateCommand {
	return &CreateTemplateCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// Handle executes the create template command.
func (c *CreateTemplateCommand) Handle(ctx context.Context, req dto.CreateTemplateRequest, createdBy *types.ID) (*dto.CreateTemplateResponse, error) {
	const op = "CreateTemplateCommand.Handle"

	// Parse locale
	locale, err := domain.ParseLocale(req.Locale)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Check if template with same slug/locale already exists
	exists, err := c.repo.ExistsBySlug(ctx, req.TenantID, req.Slug, locale)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if exists {
		return nil, domain.ErrTemplateAlreadyExists(req.Slug, locale)
	}

	// Map variables from request
	variables, err := mapVariablesFromRequest(req.Variables)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Create template aggregate
	template, err := domain.NewTemplate(
		types.NewID(),
		req.TenantID,
		req.Slug,
		locale,
		req.Name,
		req.Description,
		req.Subject,
		req.BodyHTML,
		req.BodyText,
		variables,
		createdBy,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Persist template
	if err := c.repo.Save(ctx, template); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	c.logger.Info("email template created",
		logger.String("template_id", template.ID().String()),
		logger.String("slug", template.Slug()),
		logger.String("locale", template.Locale().String()),
	)

	// Publish domain events
	events := messaging.AsDomainEvents(template.Events())
	defer template.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "email", "template", events); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("template_id", template.ID().String()),
			logger.Err(err),
		)
	}

	return &dto.CreateTemplateResponse{
		Template: dto.MapTemplateToDTO(template),
	}, nil
}

// mapVariablesFromRequest maps variable DTOs to domain variables.
func mapVariablesFromRequest(vars []dto.VariableRequest) (domain.Variables, error) {
	if len(vars) == 0 {
		return domain.Variables{}, nil
	}

	variables := make(domain.Variables, 0, len(vars))
	for _, v := range vars {
		varType, err := domain.ParseVariableType(v.Type)
		if err != nil {
			return nil, err
		}

		variable, err := domain.NewVariable(
			v.Name,
			varType,
			v.Required,
			v.Default,
			v.Description,
		)
		if err != nil {
			return nil, err
		}

		variables = append(variables, variable)
	}

	return variables, nil
}
