package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/flags/application/dto"
	"github.com/0xsj/hexagonal-go/internal/flags/domain"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// AddRuleCommand handles adding a targeting rule to a feature flag.
type AddRuleCommand struct {
	repo           domain.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewAddRuleCommand creates a new AddRuleCommand.
func NewAddRuleCommand(
	repo domain.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *AddRuleCommand {
	return &AddRuleCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// AddRuleRequest is the input for adding a rule.
type AddRuleRequest struct {
	FlagID     types.ID
	Type       string
	Attribute  string
	Operator   string
	Values     []string
	Percentage int
	VariantKey string
	Priority   int
	AddedBy    string
}

// Handle executes the add rule command.
func (c *AddRuleCommand) Handle(ctx context.Context, req AddRuleRequest) (*dto.AddRuleResponse, error) {
	const op = "AddRuleCommand.Handle"

	// Find existing flag
	flag, err := c.repo.FindByID(ctx, req.FlagID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Create rule
	ruleID := types.NewID()
	rule, err := domain.NewRule(
		ruleID,
		domain.RuleType(req.Type),
		req.Attribute,
		domain.Operator(req.Operator),
		req.Values,
		req.Percentage,
		req.VariantKey,
		req.Priority,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Add rule to flag
	if err := flag.AddRule(rule, req.AddedBy); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Persist changes
	if err := c.repo.Save(ctx, flag); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Publish domain events
	events := messaging.AsDomainEvents(flag.Events())
	defer flag.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "flags", "flag", events); err != nil {
		c.logger.Error("failed to publish flag events",
			logger.String("flag_id", req.FlagID.String()),
			logger.Err(err),
		)
	}

	c.logger.Info("rule added to flag",
		logger.String("flag_id", req.FlagID.String()),
		logger.String("flag_key", flag.Key()),
		logger.String("rule_id", ruleID.String()),
		logger.String("rule_type", req.Type),
	)

	return &dto.AddRuleResponse{
		Flag:   dto.MapFlagToDTO(flag),
		RuleID: ruleID.String(),
	}, nil
}
