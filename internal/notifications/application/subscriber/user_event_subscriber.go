package subscriber

import (
	"context"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/internal/notifications/application/command"
	"github.com/0xsj/hexagonal-go/internal/notifications/application/dto"
	"github.com/0xsj/hexagonal-go/pkg/email"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// Template slugs for email templates.
const (
	TemplateSlugWelcome            = "welcome"
	TemplateSlugEmailVerification  = "email-verification"
	TemplateSlugPasswordReset      = "password-reset"
	TemplateSlugPasswordChanged    = "password-changed"
	TemplateSlugAccountSuspended   = "account-suspended"
	TemplateSlugAccountReactivated = "account-reactivated"
)

// Default locale for templates.
const DefaultLocale = "en"

// UserEventSubscriber subscribes to user events and sends notifications.
// Implements messaging.EventHandler.
type UserEventSubscriber struct {
	sendCmd         *command.SendNotificationCommand
	templateService *email.TemplateService
	logger          logger.Logger
}

// NewUserEventSubscriber creates a new user event subscriber.
func NewUserEventSubscriber(
	sendCmd *command.SendNotificationCommand,
	templateService *email.TemplateService,
	logger logger.Logger,
) *UserEventSubscriber {
	return &UserEventSubscriber{
		sendCmd:         sendCmd,
		templateService: templateService,
		logger:          logger,
	}
}

// Register registers the subscriber for user events.
func (s *UserEventSubscriber) Register(sub messaging.Subscriber) error {
	events := []string{
		"identity.user.registered",
		"identity.password_reset_requested",
		"identity.user.password_changed",
		"identity.user.account_suspended",
		"identity.user.account_reactivated",
	}

	for _, eventType := range events {
		if err := sub.Subscribe(eventType, s); err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", eventType, err)
		}
	}

	s.logger.Info("subscribed to user events",
		logger.Int("event_count", len(events)),
	)

	return nil
}

// Handle processes user events and sends appropriate notifications.
// Implements messaging.EventHandler.
func (s *UserEventSubscriber) Handle(ctx context.Context, event messaging.Event) error {
	switch event.Type() {
	case "identity.user.registered":
		return s.handleUserRegistered(event)
	case "identity.password_reset_requested":
		return s.handlePasswordResetRequested(event)
	case "identity.user.password_changed":
		return s.handlePasswordChanged(event)
	case "identity.user.account_suspended":
		return s.handleAccountSuspended(event)
	case "identity.user.account_reactivated":
		return s.handleAccountReactivated(event)
	default:
		s.logger.Warn("unhandled event type",
			logger.String("event_type", event.Type()),
		)
		return nil
	}
}

// handleUserRegistered sends a welcome email to newly registered users.
func (s *UserEventSubscriber) handleUserRegistered(event messaging.Event) error {
	const op = "UserEventSubscriber.handleUserRegistered"

	ctx := context.Background()

	data := event.Data()
	emailAddr, _ := data["email"].(string)
	userID, _ := data["user_id"].(string)
	verificationToken, _ := data["verification_token"].(string)

	if emailAddr == "" {
		s.logger.Error("missing email in user registered event",
			logger.String("event_id", event.ID()),
		)
		return fmt.Errorf("%s: missing email in event data", op)
	}

	s.logger.Debug("handling user registered event",
		logger.String("event_id", event.ID()),
		logger.String("email", emailAddr),
		logger.String("user_id", userID),
	)

	tenantID := messaging.GetTenantID(event)
	templateData := email.TemplateData{
		"email":              emailAddr,
		"user_id":            userID,
		"verification_token": verificationToken,
		"verification_url":   fmt.Sprintf("https://app.example.com/verify?token=%s", verificationToken),
	}

	rendered, err := s.templateService.RenderTemplateWithFallback(
		ctx,
		stringPtr(tenantID),
		TemplateSlugWelcome,
		DefaultLocale,
		templateData,
		"Welcome to Hexagonal App!",
		buildWelcomeHTMLBody(emailAddr),
		buildWelcomeTextBody(emailAddr),
	)
	if err != nil {
		s.logger.Error("failed to render welcome template",
			logger.String("event_id", event.ID()),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	req := dto.SendEmailRequest{
		TenantID:      tenantID,
		Recipient:     emailAddr,
		Subject:       rendered.Subject,
		TextBody:      rendered.BodyText,
		HTMLBody:      rendered.BodyHTML,
		UserID:        userID,
		CorrelationID: messaging.GetCorrelationID(event),
		EventType:     event.Type(),
	}

	_, err = s.sendCmd.Handle(ctx, req)
	if err != nil {
		s.logger.Error("failed to send welcome email",
			logger.String("event_id", event.ID()),
			logger.String("email", emailAddr),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	s.logger.Info("welcome email sent",
		logger.String("email", emailAddr),
		logger.String("user_id", userID),
	)

	return nil
}

// handlePasswordResetRequested sends a password reset email.
func (s *UserEventSubscriber) handlePasswordResetRequested(event messaging.Event) error {
	const op = "UserEventSubscriber.handlePasswordResetRequested"

	ctx := context.Background()

	data := event.Data()
	emailAddr, _ := data["email"].(string)
	userID, _ := data["user_id"].(string)
	token, _ := data["token"].(string)

	if emailAddr == "" {
		s.logger.Error("missing email in password reset event",
			logger.String("event_id", event.ID()),
		)
		return fmt.Errorf("%s: missing email in event data", op)
	}

	s.logger.Debug("handling password reset requested event",
		logger.String("event_id", event.ID()),
		logger.String("email", emailAddr),
	)

	tenantID := messaging.GetTenantID(event)
	templateData := email.TemplateData{
		"email":     emailAddr,
		"user_id":   userID,
		"token":     token,
		"reset_url": fmt.Sprintf("https://app.example.com/reset-password?token=%s", token),
	}

	rendered, err := s.templateService.RenderTemplateWithFallback(
		ctx,
		stringPtr(tenantID),
		TemplateSlugPasswordReset,
		DefaultLocale,
		templateData,
		"Reset Your Password",
		buildPasswordResetHTMLBody(emailAddr, token),
		buildPasswordResetTextBody(emailAddr, token),
	)
	if err != nil {
		s.logger.Error("failed to render password reset template",
			logger.String("event_id", event.ID()),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	req := dto.SendEmailRequest{
		TenantID:      tenantID,
		Recipient:     emailAddr,
		Subject:       rendered.Subject,
		TextBody:      rendered.BodyText,
		HTMLBody:      rendered.BodyHTML,
		UserID:        userID,
		CorrelationID: messaging.GetCorrelationID(event),
		EventType:     event.Type(),
	}

	_, err = s.sendCmd.Handle(ctx, req)
	if err != nil {
		s.logger.Error("failed to send password reset email",
			logger.String("event_id", event.ID()),
			logger.String("email", emailAddr),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	s.logger.Info("password reset email sent",
		logger.String("email", emailAddr),
		logger.String("user_id", userID),
	)

	return nil
}

// handlePasswordChanged sends a password changed notification.
func (s *UserEventSubscriber) handlePasswordChanged(event messaging.Event) error {
	const op = "UserEventSubscriber.handlePasswordChanged"

	ctx := context.Background()

	data := event.Data()
	emailAddr, _ := data["email"].(string)
	userID, _ := data["user_id"].(string)
	changedBy, _ := data["changed_by"].(string)

	if emailAddr == "" {
		s.logger.Error("missing email in password changed event",
			logger.String("event_id", event.ID()),
		)
		return fmt.Errorf("%s: missing email in event data", op)
	}

	s.logger.Debug("handling password changed event",
		logger.String("event_id", event.ID()),
		logger.String("email", emailAddr),
		logger.String("user_id", userID),
	)

	tenantID := messaging.GetTenantID(event)
	templateData := email.TemplateData{
		"email":      emailAddr,
		"user_id":    userID,
		"changed_by": changedBy,
		"changed_at": time.Now().Format(time.RFC1123),
	}

	rendered, err := s.templateService.RenderTemplateWithFallback(
		ctx,
		stringPtr(tenantID),
		TemplateSlugPasswordChanged,
		DefaultLocale,
		templateData,
		"Your Password Has Been Changed",
		buildPasswordChangedHTMLBody(emailAddr),
		buildPasswordChangedTextBody(emailAddr),
	)
	if err != nil {
		s.logger.Error("failed to render password changed template",
			logger.String("event_id", event.ID()),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	req := dto.SendEmailRequest{
		TenantID:      tenantID,
		Recipient:     emailAddr,
		Subject:       rendered.Subject,
		TextBody:      rendered.BodyText,
		HTMLBody:      rendered.BodyHTML,
		UserID:        userID,
		CorrelationID: messaging.GetCorrelationID(event),
		EventType:     event.Type(),
	}

	_, err = s.sendCmd.Handle(ctx, req)
	if err != nil {
		s.logger.Error("failed to send password changed email",
			logger.String("event_id", event.ID()),
			logger.String("email", emailAddr),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	s.logger.Info("password changed email sent",
		logger.String("email", emailAddr),
		logger.String("user_id", userID),
	)

	return nil
}

// handleAccountSuspended sends an account suspended notification.
func (s *UserEventSubscriber) handleAccountSuspended(event messaging.Event) error {
	const op = "UserEventSubscriber.handleAccountSuspended"

	ctx := context.Background()

	data := event.Data()
	emailAddr, _ := data["email"].(string)
	userID, _ := data["user_id"].(string)
	reason, _ := data["reason"].(string)

	if emailAddr == "" {
		s.logger.Error("missing email in account suspended event",
			logger.String("event_id", event.ID()),
		)
		return fmt.Errorf("%s: missing email in event data", op)
	}

	s.logger.Debug("handling account suspended event",
		logger.String("event_id", event.ID()),
		logger.String("email", emailAddr),
		logger.String("user_id", userID),
	)

	tenantID := messaging.GetTenantID(event)
	templateData := email.TemplateData{
		"email":   emailAddr,
		"user_id": userID,
		"reason":  reason,
	}

	rendered, err := s.templateService.RenderTemplateWithFallback(
		ctx,
		stringPtr(tenantID),
		TemplateSlugAccountSuspended,
		DefaultLocale,
		templateData,
		"Your Account Has Been Suspended",
		buildAccountSuspendedHTMLBody(emailAddr, reason),
		buildAccountSuspendedTextBody(emailAddr, reason),
	)
	if err != nil {
		s.logger.Error("failed to render account suspended template",
			logger.String("event_id", event.ID()),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	req := dto.SendEmailRequest{
		TenantID:      tenantID,
		Recipient:     emailAddr,
		Subject:       rendered.Subject,
		TextBody:      rendered.BodyText,
		HTMLBody:      rendered.BodyHTML,
		UserID:        userID,
		CorrelationID: messaging.GetCorrelationID(event),
		EventType:     event.Type(),
	}

	_, err = s.sendCmd.Handle(ctx, req)
	if err != nil {
		s.logger.Error("failed to send account suspended email",
			logger.String("event_id", event.ID()),
			logger.String("email", emailAddr),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	s.logger.Info("account suspended email sent",
		logger.String("email", emailAddr),
		logger.String("user_id", userID),
	)

	return nil
}

// handleAccountReactivated sends an account reactivated notification.
func (s *UserEventSubscriber) handleAccountReactivated(event messaging.Event) error {
	const op = "UserEventSubscriber.handleAccountReactivated"

	ctx := context.Background()

	data := event.Data()
	emailAddr, _ := data["email"].(string)
	userID, _ := data["user_id"].(string)

	if emailAddr == "" {
		s.logger.Error("missing email in account reactivated event",
			logger.String("event_id", event.ID()),
		)
		return fmt.Errorf("%s: missing email in event data", op)
	}

	s.logger.Debug("handling account reactivated event",
		logger.String("event_id", event.ID()),
		logger.String("email", emailAddr),
		logger.String("user_id", userID),
	)

	tenantID := messaging.GetTenantID(event)
	templateData := email.TemplateData{
		"email":   emailAddr,
		"user_id": userID,
	}

	rendered, err := s.templateService.RenderTemplateWithFallback(
		ctx,
		stringPtr(tenantID),
		TemplateSlugAccountReactivated,
		DefaultLocale,
		templateData,
		"Your Account Has Been Reactivated",
		buildAccountReactivatedHTMLBody(emailAddr),
		buildAccountReactivatedTextBody(emailAddr),
	)
	if err != nil {
		s.logger.Error("failed to render account reactivated template",
			logger.String("event_id", event.ID()),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	req := dto.SendEmailRequest{
		TenantID:      tenantID,
		Recipient:     emailAddr,
		Subject:       rendered.Subject,
		TextBody:      rendered.BodyText,
		HTMLBody:      rendered.BodyHTML,
		UserID:        userID,
		CorrelationID: messaging.GetCorrelationID(event),
		EventType:     event.Type(),
	}

	_, err = s.sendCmd.Handle(ctx, req)
	if err != nil {
		s.logger.Error("failed to send account reactivated email",
			logger.String("event_id", event.ID()),
			logger.String("email", emailAddr),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	s.logger.Info("account reactivated email sent",
		logger.String("email", emailAddr),
		logger.String("user_id", userID),
	)

	return nil
}

// stringPtr returns a pointer to the string, or nil if empty.
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ============================================================================
// Fallback Templates
// ============================================================================

func buildWelcomeTextBody(emailAddr string) string {
	return fmt.Sprintf(`Welcome to Hexagonal App!

Hello %s,

Thank you for registering. Your account has been created successfully.

Please verify your email address to get started.

Best regards,
The Hexagonal Team
`, emailAddr)
}

func buildWelcomeHTMLBody(emailAddr string) string {
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
`, emailAddr)
}

func buildPasswordResetTextBody(emailAddr string, token string) string {
	return fmt.Sprintf(`Password Reset Request

Hello %s,

We received a request to reset your password. Click the link below to reset it:

https://app.example.com/reset-password?token=%s

If you didn't request this, you can safely ignore this email.

Best regards,
The Hexagonal Team
`, emailAddr, token)
}

func buildPasswordResetHTMLBody(emailAddr string, token string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2563eb;">Password Reset Request</h1>
        <p>Hello %s,</p>
        <p>We received a request to reset your password. Click the button below to reset it:</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="https://app.example.com/reset-password?token=%s" 
               style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px;">
                Reset Password
            </a>
        </p>
        <p>If you didn't request this, you can safely ignore this email.</p>
        <br>
        <p>Best regards,<br>The Hexagonal Team</p>
    </div>
</body>
</html>
`, emailAddr, token)
}

func buildPasswordChangedTextBody(emailAddr string) string {
	return fmt.Sprintf(`Password Changed

Hello %s,

Your password was successfully changed.

If you did not make this change, please contact support immediately.

Best regards,
The Hexagonal Team
`, emailAddr)
}

func buildPasswordChangedHTMLBody(emailAddr string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Password Changed</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2563eb;">Password Changed</h1>
        <p>Hello %s,</p>
        <p>Your password was successfully changed.</p>
        <p style="background-color: #fef3c7; padding: 15px; border-radius: 4px; border-left: 4px solid #f59e0b;">
            <strong>Security Notice:</strong> If you did not make this change, please contact support immediately.
        </p>
        <br>
        <p>Best regards,<br>The Hexagonal Team</p>
    </div>
</body>
</html>
`, emailAddr)
}

func buildAccountSuspendedTextBody(emailAddr string, reason string) string {
	text := fmt.Sprintf(`Account Suspended

Hello %s,

Your account has been suspended.
`, emailAddr)

	if reason != "" {
		text += fmt.Sprintf("\nReason: %s\n", reason)
	}

	text += `
If you believe this is an error, please contact our support team.

Best regards,
The Hexagonal Team
`
	return text
}

func buildAccountSuspendedHTMLBody(emailAddr string, reason string) string {
	reasonHTML := ""
	if reason != "" {
		reasonHTML = fmt.Sprintf("<p><strong>Reason:</strong> %s</p>", reason)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Account Suspended</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #dc2626;">Account Suspended</h1>
        <p>Hello %s,</p>
        <p>Your account has been suspended.</p>
        %s
        <p>If you believe this is an error, please contact our support team.</p>
        <br>
        <p>Best regards,<br>The Hexagonal Team</p>
    </div>
</body>
</html>
`, emailAddr, reasonHTML)
}

func buildAccountReactivatedTextBody(emailAddr string) string {
	return fmt.Sprintf(`Account Reactivated

Hello %s,

Good news! Your account has been reactivated. You can now log in and use all features.

Best regards,
The Hexagonal Team
`, emailAddr)
}

func buildAccountReactivatedHTMLBody(emailAddr string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Account Reactivated</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #16a34a;">Account Reactivated</h1>
        <p>Hello %s,</p>
        <p>Good news! Your account has been reactivated. You can now log in and use all features.</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="https://app.example.com/login" 
               style="background-color: #16a34a; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px;">
                Log In Now
            </a>
        </p>
        <br>
        <p>Best regards,<br>The Hexagonal Team</p>
    </div>
</body>
</html>
`, emailAddr)
}
