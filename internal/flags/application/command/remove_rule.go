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

// RemoveRuleCommand handles removing a targeting rule from a feature flag.
type RemoveRuleCommand struct {
	repo           domain.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewRemoveRuleCommand creates a new RemoveRuleCommand.
func NewRemoveRuleCommand(
	repo domain.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *RemoveRuleCommand {
	return &RemoveRuleCommand{
		repo:           repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// RemoveRuleRequest is the input for removing a rule.
type RemoveRuleRequest struct {
	FlagID    types.ID
	RuleID    types.ID
	RemovedBy string
}

// Handle executes the remove rule command.
func (c *RemoveRuleCommand) Handle(ctx context.Context, req RemoveRuleRequest) (*dto.RemoveRuleResponse, error) {
	const op = "RemoveRuleCommand.Handle"

	// Find existing flag
	flag, err := c.repo.FindByID(ctx, req.FlagID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Remove rule from flag
	if err := flag.RemoveRule(req.RuleID, req.RemovedBy); err != nil {
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

	c.logger.Info("rule removed from flag",
		logger.String("flag_id", req.FlagID.String()),
		logger.String("flag_key", flag.Key()),
		logger.String("rule_id", req.RuleID.String()),
	)

	return &dto.RemoveRuleResponse{
		Flag: dto.MapFlagToDTO(flag),
	}, nil
}
