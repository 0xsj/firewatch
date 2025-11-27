package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/0xsj/hexagonal-go/pkg/email"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/worker"
)

// Job type constants.
const (
	JobTypeSendEmail = "notification.send_email"
)

// SendEmailPayload is the payload for email jobs.
type SendEmailPayload struct {
	To            []string `json:"to"`
	Subject       string   `json:"subject"`
	TextBody      string   `json:"text_body"`
	HTMLBody      string   `json:"html_body,omitempty"`
	TenantID      string   `json:"tenant_id,omitempty"`
	UserID        string   `json:"user_id,omitempty"`
	CorrelationID string   `json:"correlation_id,omitempty"`
}

// SendEmailHandler handles email sending jobs.
type SendEmailHandler struct {
	sender email.Sender
	logger logger.Logger
}

// NewSendEmailHandler creates a new email handler.
func NewSendEmailHandler(sender email.Sender, logger logger.Logger) *SendEmailHandler {
	return &SendEmailHandler{
		sender: sender,
		logger: logger,
	}
}

// Handle processes an email job.
func (h *SendEmailHandler) Handle(ctx context.Context, job *worker.Job) error {
	var payload SendEmailPayload
	if err := job.UnmarshalPayload(&payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	h.logger.Debug("processing email job",
		logger.String("job_id", job.ID().String()),
		logger.String("to", strings.Join(payload.To, ",")),
		logger.String("subject", payload.Subject),
	)

	// Build email message
	msg := email.NewMessage(payload.To, payload.Subject, payload.TextBody)
	if payload.HTMLBody != "" {
		msg = msg.WithHTML(payload.HTMLBody)
	}

	// Send email
	if err := h.sender.Send(ctx, msg); err != nil {
		h.logger.Error("failed to send email",
			logger.String("job_id", job.ID().String()),
			logger.Err(err),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	h.logger.Info("email sent successfully",
		logger.String("job_id", job.ID().String()),
		logger.String("to", strings.Join(payload.To, ",")),
	)

	return nil
}

// Registry holds all job handlers.
type Registry struct {
	handlers map[string]worker.Handler
}

// NewRegistry creates a new handler registry.
func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]worker.Handler),
	}
}

// Register adds a handler for a job type.
func (r *Registry) Register(jobType string, handler worker.Handler) {
	r.handlers[jobType] = handler
}

// RegisterAll registers all handlers with the worker.
func (r *Registry) RegisterAll(w *worker.Worker) {
	for jobType, handler := range r.handlers {
		w.Register(jobType, handler)
	}
}

// SetupHandlers creates and registers all job handlers.
func SetupHandlers(
	emailSender email.Sender,
	log logger.Logger,
) *Registry {
	registry := NewRegistry()

	// Register email handler
	emailHandler := NewSendEmailHandler(emailSender, log)
	registry.Register(JobTypeSendEmail, emailHandler)

	return registry
}
