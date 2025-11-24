package subscriber

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/notifications/application/command"
	"github.com/0xsj/hexagonal-go/internal/notifications/application/dto"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// UserEventSubscriber subscribes to user events and sends notifications.
// Implements messaging.EventHandler.
type UserEventSubscriber struct {
	sendCmd *command.SendNotificationCommand
	logger  logger.Logger
}

// NewUserEventSubscriber creates a new user event subscriber.
func NewUserEventSubscriber(
	sendCmd *command.SendNotificationCommand,
	logger logger.Logger,
) *UserEventSubscriber {
	return &UserEventSubscriber{
		sendCmd: sendCmd,
		logger:  logger,
	}
}

// Register registers the subscriber for user events.
func (s *UserEventSubscriber) Register(sub messaging.Subscriber) error {
	// Subscribe to user registration events
	if err := sub.Subscribe("identity.user.registered", s); err != nil {
		return fmt.Errorf("failed to subscribe to user.registered: %w", err)
	}

	s.logger.Info("subscribed to user events",
		logger.String("event_type", "identity.user.registered"),
	)

	return nil
}

// Handle processes user events and sends appropriate notifications.
// Implements messaging.EventHandler.
func (s *UserEventSubscriber) Handle(ctx context.Context, event messaging.Event) error {
	const op = "notifications.UserEventSubscriber.Handle"

	switch event.Type() {
	case "identity.user.registered":
		return s.handleUserRegistered(ctx, event)
	default:
		s.logger.Warn("unhandled event type",
			logger.String("event_type", event.Type()),
		)
		return nil
	}
}

// handleUserRegistered sends a welcome email to newly registered users.
func (s *UserEventSubscriber) handleUserRegistered(ctx context.Context, event messaging.Event) error {
	const op = "notifications.UserEventSubscriber.handleUserRegistered"

	// Extract data from event
	data := event.Data()
	email, _ := data["email"].(string)
	userID, _ := data["user_id"].(string)

	if email == "" {
		s.logger.Error("missing email in user registered event",
			logger.String("event_id", event.ID()),
		)
		return fmt.Errorf("%s: missing email in event data", op)
	}

	s.logger.Debug("handling user registered event",
		logger.String("event_id", event.ID()),
		logger.String("email", email),
		logger.String("user_id", userID),
	)

	// Build welcome email
	req := dto.SendEmailRequest{
		TenantID:      messaging.GetTenantID(event),
		Recipient:     email,
		Subject:       "Welcome to Hexagonal App!",
		TextBody:      buildWelcomeTextBody(email),
		HTMLBody:      buildWelcomeHTMLBody(email),
		UserID:        userID,
		CorrelationID: messaging.GetCorrelationID(event),
		EventType:     event.Type(),
	}

	// Use background context to avoid cancellation issues
	sendCtx := context.Background()

	// Send notification
	_, err := s.sendCmd.Handle(sendCtx, req)
	if err != nil {
		s.logger.Error("failed to send welcome email",
			logger.String("event_id", event.ID()),
			logger.String("email", email),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	s.logger.Info("welcome email sent",
		logger.String("email", email),
		logger.String("user_id", userID),
	)

	return nil
}

// buildWelcomeTextBody builds the plain text welcome email body.
func buildWelcomeTextBody(email string) string {
	return fmt.Sprintf(`Welcome to Hexagonal App!

Hello %s,

Thank you for registering. Your account has been created successfully.

Please verify your email address to get started.

Best regards,
The Hexagonal Team
`, email)
}

// buildWelcomeHTMLBody builds the HTML welcome email body.
func buildWelcomeHTMLBody(email string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to Hexagonal App</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2563eb;">Welcome to Hexagonal App!</h1>
        <p>Hello %s,</p>
        <p>Thank you for registering. Your account has been created successfully.</p>
        <p>Please verify your email address to get started.</p>
        <br>
        <p>Best regards,<br>The Hexagonal Team</p>
    </div>
</body>
</html>
`, email)
}
