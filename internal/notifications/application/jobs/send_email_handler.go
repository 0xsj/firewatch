package jobs

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/pkg/email"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/worker"
)

// SendEmailHandler processes send_email jobs.
type SendEmailHandler struct {
	sender email.Sender
	logger logger.Logger
}

// NewSendEmailHandler creates a new SendEmailHandler.
func NewSendEmailHandler(sender email.Sender, logger logger.Logger) *SendEmailHandler {
	return &SendEmailHandler{
		sender: sender,
		logger: logger,
	}
}

// Handle processes a send_email job.
func (h *SendEmailHandler) Handle(ctx context.Context, job *worker.Job) error {
	const op = "SendEmailHandler.Handle"

	var payload SendEmailPayload
	if err := job.UnmarshalPayload(&payload); err != nil {
		return fmt.Errorf("%s: failed to unmarshal payload: %w", op, err)
	}

	h.logger.Debug("processing send_email job",
		logger.String("job_id", job.ID().String()),
		logger.String("recipient", payload.Recipient),
		logger.String("subject", payload.Subject),
	)

	// Build email message
	msg := email.NewMessage(
		[]string{payload.Recipient},
		payload.Subject,
		payload.TextBody,
	).WithHTML(payload.HTMLBody)

	// Send email
	if err := h.sender.Send(ctx, msg); err != nil {
		h.logger.Error("failed to send email",
			logger.String("job_id", job.ID().String()),
			logger.String("recipient", payload.Recipient),
			logger.Err(err),
		)
		return fmt.Errorf("%s: failed to send email: %w", op, err)
	}

	h.logger.Info("email sent successfully",
		logger.String("job_id", job.ID().String()),
		logger.String("recipient", payload.Recipient),
		logger.String("event_type", payload.EventType),
	)

	return nil
}

// Type returns the job type this handler processes.
func (h *SendEmailHandler) Type() string {
	return JobTypeSendEmail
}
