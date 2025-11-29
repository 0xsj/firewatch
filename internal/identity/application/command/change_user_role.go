package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// ChangeUserRoleCommand handles role changes by admins.
type ChangeUserRoleCommand struct {
	userRepo       user.Repository
	eventPublisher *messaging.DomainEventPublisher
	logger         logger.Logger
}

// NewChangeUserRoleCommand creates a new ChangeUserRoleCommand.
func NewChangeUserRoleCommand(
	userRepo user.Repository,
	eventPublisher *messaging.DomainEventPublisher,
	logger logger.Logger,
) *ChangeUserRoleCommand {
	return &ChangeUserRoleCommand{
		userRepo:       userRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// ChangeUserRoleRequest is the input for changing a user's role.
type ChangeUserRoleRequest struct {
	UserID    types.ID
	NewRole   user.Role
	Reason    string
	ChangedBy types.ID
}

// Handle executes the change user role command.
func (c *ChangeUserRoleCommand) Handle(ctx context.Context, req ChangeUserRoleRequest) error {
	const op = "ChangeUserRoleCommand.Handle"

	// Find user
	u, err := c.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("%s: failed to find user: %w", op, err)
	}

	// Change role (domain logic)
	if err := u.ChangeRole(req.NewRole, req.ChangedBy.String(), req.Reason); err != nil {
		if pkgerrors.IsValidation(err) {
			return pkgerrors.Validation(op, err.Error())
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	// Save user
	if err := c.userRepo.Save(ctx, u); err != nil {
		return fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	c.logger.Info("user role changed",
		logger.String("user_id", u.ID().String()),
		logger.String("new_role", string(req.NewRole)),
		logger.String("changed_by", req.ChangedBy.String()),
		logger.String("reason", req.Reason),
	)

	// Publish domain events
	events := messaging.AsDomainEvents(u.Events())
	defer u.ClearEvents()

	if err := c.eventPublisher.PublishAll(ctx, "identity", "user", events); err != nil {
		c.logger.Error("failed to publish events",
			logger.String("user_id", u.ID().String()),
			logger.Err(err),
		)
	}

	return nil
}
