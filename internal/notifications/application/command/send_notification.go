package command

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/notifications/application/dto"
	"github.com/0xsj/hexagonal-go/internal/notifications/domain"
	"github.com/0xsj/hexagonal-go/pkg/email"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// SendNotificationCommand handles sending notifications.
type SendNotificationCommand struct {
	repo        domain.Repository
	emailSender email.Sender
	logger      logger.Logger
}

// NewSendNotificationCommand creates a new SendNotificationCommand.
func NewSendNotificationCommand(
	repo domain.Repository,
	emailSender email.Sender,
	logger logger.Logger,
) *SendNotificationCommand {
	return &SendNotificationCommand{
		repo:        repo,
		emailSender: emailSender,
		logger:      logger,
	}
}

// Handle executes the send notification command.
func (c *SendNotificationCommand) Handle(ctx context.Context, req dto.SendEmailRequest) (*dto.NotificationDTO, error) {
	const op = "SendNotificationCommand.Handle"

	// Create notification record
	notification := domain.NewNotification(
		req.TenantID,
		domain.ChannelEmail,
		req.Recipient,
		req.Subject,
		req.TextBody,
		req.UserID,
		req.CorrelationID,
		req.EventType,
	)

	// Save as pending
	if err := c.repo.Save(ctx, notification); err != nil {
		return nil, fmt.Errorf("%s: failed to save notification: %w", op, err)
	}

	c.logger.Debug("notification created",
		logger.String("notification_id", notification.ID().String()),
		logger.String("channel", notification.Channel().String()),
		logger.String("recipient", notification.Recipient()),
	)

	// Build email message
	msg := email.NewMessage(
		[]string{req.Recipient},
		req.Subject,
		req.TextBody,
	)
	if req.HTMLBody != "" {
		msg = msg.WithHTML(req.HTMLBody)
	}

	// Send email
	if err := c.emailSender.Send(ctx, msg); err != nil {
		// Mark as failed
		notification.MarkFailed(err.Error())
		if saveErr := c.repo.Save(ctx, notification); saveErr != nil {
			c.logger.Error("failed to save notification failure",
				logger.String("notification_id", notification.ID().String()),
				logger.Err(saveErr),
			)
		}

		c.logger.Error("failed to send email",
			logger.String("notification_id", notification.ID().String()),
			logger.String("recipient", req.Recipient),
			logger.Err(err),
		)
		return dto.MapNotificationToDTO(notification), fmt.Errorf("%s: failed to send email: %w", op, err)
	}

	// Mark as sent
	if err := notification.MarkSent(); err != nil {
		return nil, fmt.Errorf("%s: failed to mark sent: %w", op, err)
	}

	if err := c.repo.Save(ctx, notification); err != nil {
		return nil, fmt.Errorf("%s: failed to save notification: %w", op, err)
	}

	c.logger.Info("notification sent",
		logger.String("notification_id", notification.ID().String()),
		logger.String("channel", notification.Channel().String()),
		logger.String("recipient", notification.Recipient()),
	)

	return dto.MapNotificationToDTO(notification), nil
}
