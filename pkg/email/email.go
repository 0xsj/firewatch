package email

import "context"

// Sender is the port for sending emails.
// Implementations include SMTP, SendGrid, SES, etc.
type Sender interface {
	// Send sends an email message.
	Send(ctx context.Context, msg Message) error

	// SendBatch sends multiple email messages.
	SendBatch(ctx context.Context, msgs []Message) error

	// Ping verifies the connection to the email service.
	Ping(ctx context.Context) error

	// Close closes the connection.
	Close() error
}
