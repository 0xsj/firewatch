package subscriber

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/notifications/application/jobs"
	"github.com/0xsj/hexagonal-go/pkg/email"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/worker"
)

// Template slugs for tenant notifications.
const (
	TemplateSlugTenantWelcome     = "tenant-welcome"
	TemplateSlugTenantSuspended   = "tenant-suspended"
	TemplateSlugTenantReactivated = "tenant-reactivated"
	TemplateSlugTenantPlanChanged = "tenant-plan-changed"
)

// TenantEventSubscriber subscribes to tenant domain events and enqueues notification jobs.
type TenantEventSubscriber struct {
	queue           worker.Queue
	templateService *email.TemplateService
	logger          logger.Logger
}

// NewTenantEventSubscriber creates a new TenantEventSubscriber.
func NewTenantEventSubscriber(
	queue worker.Queue,
	templateService *email.TemplateService,
	logger logger.Logger,
) *TenantEventSubscriber {
	return &TenantEventSubscriber{
		queue:           queue,
		templateService: templateService,
		logger:          logger,
	}
}

// Register subscribes to tenant events.
func (s *TenantEventSubscriber) Register(sub messaging.Subscriber) error {
	events := []string{
		"tenant.tenant.created",
		"tenant.tenant.suspended",
		"tenant.tenant.reactivated",
		"tenant.tenant.plan_changed",
	}

	for _, eventType := range events {
		if err := sub.Subscribe(eventType, s); err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", eventType, err)
		}
	}

	s.logger.Info("tenant event subscriber registered",
		logger.Int("event_count", len(events)),
	)

	return nil
}

// Handle processes tenant events.
func (s *TenantEventSubscriber) Handle(ctx context.Context, event messaging.Event) error {
	s.logger.Debug("handling tenant event",
		logger.String("event_type", event.Type()),
		logger.String("event_id", event.ID()),
	)

	switch event.Type() {
	case "tenant.tenant.created":
		return s.handleTenantCreated(ctx, event)
	case "tenant.tenant.suspended":
		return s.handleTenantSuspended(ctx, event)
	case "tenant.tenant.reactivated":
		return s.handleTenantReactivated(ctx, event)
	case "tenant.tenant.plan_changed":
		return s.handleTenantPlanChanged(ctx, event)
	default:
		s.logger.Warn("unhandled tenant event type",
			logger.String("event_type", event.Type()),
		)
		return nil
	}
}

// handleTenantCreated enqueues a welcome email job for the tenant owner.
func (s *TenantEventSubscriber) handleTenantCreated(ctx context.Context, event messaging.Event) error {
	const op = "TenantEventSubscriber.handleTenantCreated"

	data := event.Data()

	ownerEmail, _ := data["owner_email"].(string)
	if ownerEmail == "" {
		s.logger.Warn("tenant created event missing owner_email",
			logger.String("event_id", event.ID()),
		)
		return nil
	}

	tenantName, _ := data["name"].(string)
	tenantSlug, _ := data["slug"].(string)
	tenantID, _ := data["tenant_id"].(string)
	plan, _ := data["plan"].(string)

	templateData := email.TemplateData{
		"tenant_name": tenantName,
		"tenant_slug": tenantSlug,
		"tenant_id":   tenantID,
		"plan":        plan,
		"email":       ownerEmail,
	}

	rendered, err := s.templateService.RenderTemplateWithFallback(
		ctx,
		nil,
		TemplateSlugTenantWelcome,
		DefaultLocale,
		templateData,
		"Welcome to Your New Workspace!",
		buildTenantWelcomeHTMLBody(ownerEmail, tenantName, tenantSlug),
		buildTenantWelcomeTextBody(ownerEmail, tenantName, tenantSlug),
	)
	if err != nil {
		s.logger.Error("failed to render tenant welcome template",
			logger.String("event_id", event.ID()),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	payload := jobs.SendEmailPayload{
		TenantID:      tenantID,
		Recipient:     ownerEmail,
		Subject:       rendered.Subject,
		HTMLBody:      rendered.BodyHTML,
		TextBody:      rendered.BodyText,
		CorrelationID: messaging.GetCorrelationID(event),
		EventType:     event.Type(),
	}

	job, err := worker.NewJobWithData(jobs.JobTypeSendEmail, payload)
	if err != nil {
		return fmt.Errorf("%s: failed to create job: %w", op, err)
	}

	if err := s.queue.Enqueue(ctx, job); err != nil {
		return fmt.Errorf("%s: failed to enqueue job: %w", op, err)
	}

	s.logger.Info("tenant welcome email job enqueued",
		logger.String("job_id", job.ID().String()),
		logger.String("email", ownerEmail),
		logger.String("tenant_slug", tenantSlug),
	)

	return nil
}

// handleTenantSuspended enqueues a suspension notification job.
func (s *TenantEventSubscriber) handleTenantSuspended(ctx context.Context, event messaging.Event) error {
	const op = "TenantEventSubscriber.handleTenantSuspended"

	data := event.Data()

	ownerEmail, _ := data["owner_email"].(string)
	if ownerEmail == "" {
		s.logger.Warn("tenant suspended event missing owner_email",
			logger.String("event_id", event.ID()),
		)
		return nil
	}

	tenantID, _ := data["tenant_id"].(string)
	reason, _ := data["reason"].(string)

	templateData := email.TemplateData{
		"tenant_id": tenantID,
		"reason":    reason,
		"email":     ownerEmail,
	}

	rendered, err := s.templateService.RenderTemplateWithFallback(
		ctx,
		nil,
		TemplateSlugTenantSuspended,
		DefaultLocale,
		templateData,
		"Your Workspace Has Been Suspended",
		buildTenantSuspendedHTMLBody(ownerEmail, reason),
		buildTenantSuspendedTextBody(ownerEmail, reason),
	)
	if err != nil {
		s.logger.Error("failed to render tenant suspended template",
			logger.String("event_id", event.ID()),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	payload := jobs.SendEmailPayload{
		TenantID:      tenantID,
		Recipient:     ownerEmail,
		Subject:       rendered.Subject,
		HTMLBody:      rendered.BodyHTML,
		TextBody:      rendered.BodyText,
		CorrelationID: messaging.GetCorrelationID(event),
		EventType:     event.Type(),
	}

	job, err := worker.NewJobWithData(jobs.JobTypeSendEmail, payload)
	if err != nil {
		return fmt.Errorf("%s: failed to create job: %w", op, err)
	}

	if err := s.queue.Enqueue(ctx, job); err != nil {
		return fmt.Errorf("%s: failed to enqueue job: %w", op, err)
	}

	s.logger.Info("tenant suspended email job enqueued",
		logger.String("job_id", job.ID().String()),
		logger.String("email", ownerEmail),
		logger.String("tenant_id", tenantID),
	)

	return nil
}

// handleTenantReactivated enqueues a reactivation notification job.
func (s *TenantEventSubscriber) handleTenantReactivated(ctx context.Context, event messaging.Event) error {
	const op = "TenantEventSubscriber.handleTenantReactivated"

	data := event.Data()

	ownerEmail, _ := data["owner_email"].(string)
	if ownerEmail == "" {
		s.logger.Warn("tenant reactivated event missing owner_email",
			logger.String("event_id", event.ID()),
		)
		return nil
	}

	tenantID, _ := data["tenant_id"].(string)

	templateData := email.TemplateData{
		"tenant_id": tenantID,
		"email":     ownerEmail,
	}

	rendered, err := s.templateService.RenderTemplateWithFallback(
		ctx,
		nil,
		TemplateSlugTenantReactivated,
		DefaultLocale,
		templateData,
		"Your Workspace Has Been Reactivated",
		buildTenantReactivatedHTMLBody(ownerEmail),
		buildTenantReactivatedTextBody(ownerEmail),
	)
	if err != nil {
		s.logger.Error("failed to render tenant reactivated template",
			logger.String("event_id", event.ID()),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	payload := jobs.SendEmailPayload{
		TenantID:      tenantID,
		Recipient:     ownerEmail,
		Subject:       rendered.Subject,
		HTMLBody:      rendered.BodyHTML,
		TextBody:      rendered.BodyText,
		CorrelationID: messaging.GetCorrelationID(event),
		EventType:     event.Type(),
	}

	job, err := worker.NewJobWithData(jobs.JobTypeSendEmail, payload)
	if err != nil {
		return fmt.Errorf("%s: failed to create job: %w", op, err)
	}

	if err := s.queue.Enqueue(ctx, job); err != nil {
		return fmt.Errorf("%s: failed to enqueue job: %w", op, err)
	}

	s.logger.Info("tenant reactivated email job enqueued",
		logger.String("job_id", job.ID().String()),
		logger.String("email", ownerEmail),
		logger.String("tenant_id", tenantID),
	)

	return nil
}

// handleTenantPlanChanged enqueues a plan change confirmation job.
func (s *TenantEventSubscriber) handleTenantPlanChanged(ctx context.Context, event messaging.Event) error {
	const op = "TenantEventSubscriber.handleTenantPlanChanged"

	data := event.Data()

	ownerEmail, _ := data["owner_email"].(string)
	if ownerEmail == "" {
		s.logger.Warn("tenant plan changed event missing owner_email",
			logger.String("event_id", event.ID()),
		)
		return nil
	}

	tenantID, _ := data["tenant_id"].(string)
	oldPlan, _ := data["old_plan"].(string)
	newPlan, _ := data["new_plan"].(string)

	templateData := email.TemplateData{
		"tenant_id": tenantID,
		"old_plan":  oldPlan,
		"new_plan":  newPlan,
		"email":     ownerEmail,
	}

	rendered, err := s.templateService.RenderTemplateWithFallback(
		ctx,
		nil,
		TemplateSlugTenantPlanChanged,
		DefaultLocale,
		templateData,
		"Your Plan Has Been Changed",
		buildTenantPlanChangedHTMLBody(ownerEmail, oldPlan, newPlan),
		buildTenantPlanChangedTextBody(ownerEmail, oldPlan, newPlan),
	)
	if err != nil {
		s.logger.Error("failed to render tenant plan changed template",
			logger.String("event_id", event.ID()),
			logger.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	payload := jobs.SendEmailPayload{
		TenantID:      tenantID,
		Recipient:     ownerEmail,
		Subject:       rendered.Subject,
		HTMLBody:      rendered.BodyHTML,
		TextBody:      rendered.BodyText,
		CorrelationID: messaging.GetCorrelationID(event),
		EventType:     event.Type(),
		Metadata: map[string]any{
			"old_plan": oldPlan,
			"new_plan": newPlan,
		},
	}

	job, err := worker.NewJobWithData(jobs.JobTypeSendEmail, payload)
	if err != nil {
		return fmt.Errorf("%s: failed to create job: %w", op, err)
	}

	if err := s.queue.Enqueue(ctx, job); err != nil {
		return fmt.Errorf("%s: failed to enqueue job: %w", op, err)
	}

	s.logger.Info("tenant plan changed email job enqueued",
		logger.String("job_id", job.ID().String()),
		logger.String("email", ownerEmail),
		logger.String("tenant_id", tenantID),
		logger.String("old_plan", oldPlan),
		logger.String("new_plan", newPlan),
	)

	return nil
}

// ============================================================================
// Fallback Templates
// ============================================================================

func buildTenantWelcomeTextBody(email, tenantName, tenantSlug string) string {
	return fmt.Sprintf(`Welcome to Your New Workspace!

Hello %s,

Your workspace "%s" has been created successfully.

Your workspace URL: https://app.example.com/%s

Best regards,
The Hexagonal Team
`, email, tenantName, tenantSlug)
}

func buildTenantWelcomeHTMLBody(email, tenantName, tenantSlug string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to Your New Workspace</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2563eb;">Welcome to Your New Workspace!</h1>
        <p>Hello %s,</p>
        <p>Your workspace <strong>"%s"</strong> has been created successfully.</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="https://app.example.com/%s" 
               style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px;">
                Go to Your Workspace
            </a>
        </p>
        <br>
        <p>Best regards,<br>The Hexagonal Team</p>
    </div>
</body>
</html>
`, email, tenantName, tenantSlug)
}

func buildTenantSuspendedTextBody(email, reason string) string {
	text := fmt.Sprintf(`Workspace Suspended

Hello %s,

Your workspace has been suspended.
`, email)

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

func buildTenantSuspendedHTMLBody(email, reason string) string {
	reasonHTML := ""
	if reason != "" {
		reasonHTML = fmt.Sprintf("<p><strong>Reason:</strong> %s</p>", reason)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Workspace Suspended</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #dc2626;">Workspace Suspended</h1>
        <p>Hello %s,</p>
        <p>Your workspace has been suspended.</p>
        %s
        <p>If you believe this is an error, please contact our support team.</p>
        <br>
        <p>Best regards,<br>The Hexagonal Team</p>
    </div>
</body>
</html>
`, email, reasonHTML)
}

func buildTenantReactivatedTextBody(email string) string {
	return fmt.Sprintf(`Workspace Reactivated

Hello %s,

Good news! Your workspace has been reactivated. You can now access all features.

Best regards,
The Hexagonal Team
`, email)
}

func buildTenantReactivatedHTMLBody(email string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Workspace Reactivated</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #16a34a;">Workspace Reactivated</h1>
        <p>Hello %s,</p>
        <p>Good news! Your workspace has been reactivated. You can now access all features.</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="https://app.example.com/login" 
               style="background-color: #16a34a; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px;">
                Go to Workspace
            </a>
        </p>
        <br>
        <p>Best regards,<br>The Hexagonal Team</p>
    </div>
</body>
</html>
`, email)
}

func buildTenantPlanChangedTextBody(email, oldPlan, newPlan string) string {
	return fmt.Sprintf(`Plan Changed

Hello %s,

Your workspace plan has been changed from %s to %s.

Best regards,
The Hexagonal Team
`, email, oldPlan, newPlan)
}

func buildTenantPlanChangedHTMLBody(email, oldPlan, newPlan string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Plan Changed</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2563eb;">Plan Changed</h1>
        <p>Hello %s,</p>
        <p>Your workspace plan has been changed:</p>
        <ul>
            <li><strong>Previous Plan:</strong> %s</li>
            <li><strong>New Plan:</strong> %s</li>
        </ul>
        <br>
        <p>Best regards,<br>The Hexagonal Team</p>
    </div>
</body>
</html>
`, email, oldPlan, newPlan)
}
