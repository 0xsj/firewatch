package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/email"
)

// Sender is an SMTP implementation of email.Sender.
type Sender struct {
	config email.Config
}

// New creates a new SMTP email sender.
func New(config email.Config) *Sender {
	return &Sender{
		config: config,
	}
}

// Send sends an email message via SMTP.
func (s *Sender) Send(ctx context.Context, msg email.Message) error {
	// Build the email
	from := msg.From
	if from == "" {
		from = s.config.FromAddress
	}

	// Build recipients list
	recipients := make([]string, 0, len(msg.To)+len(msg.CC)+len(msg.BCC))
	recipients = append(recipients, msg.To...)
	recipients = append(recipients, msg.CC...)
	recipients = append(recipients, msg.BCC...)

	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	// Build email content
	content := s.buildMessage(from, msg)

	// Send via SMTP
	return s.sendMail(ctx, from, recipients, content)
}

// SendBatch sends multiple email messages.
func (s *Sender) SendBatch(ctx context.Context, msgs []email.Message) error {
	for _, msg := range msgs {
		if err := s.Send(ctx, msg); err != nil {
			return fmt.Errorf("failed to send email to %v: %w", msg.To, err)
		}
	}
	return nil
}

// Ping verifies the connection to the SMTP server.
func (s *Sender) Ping(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	dialer := &net.Dialer{
		Timeout: s.config.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	return nil
}

// Close closes the sender (no-op for SMTP as connections are per-send).
func (s *Sender) Close() error {
	return nil
}

// buildMessage constructs the raw email message.
func (s *Sender) buildMessage(from string, msg email.Message) []byte {
	var b strings.Builder

	// From header
	fromName := s.config.FromName
	if fromName != "" {
		b.WriteString(fmt.Sprintf("From: %s <%s>\r\n", fromName, from))
	} else {
		b.WriteString(fmt.Sprintf("From: %s\r\n", from))
	}

	// To header
	b.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))

	// CC header
	if len(msg.CC) > 0 {
		b.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.CC, ", ")))
	}

	// Reply-To header
	if msg.ReplyTo != "" {
		b.WriteString(fmt.Sprintf("Reply-To: %s\r\n", msg.ReplyTo))
	}

	// Subject header
	b.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))

	// Custom headers
	for key, value := range msg.Headers {
		b.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	// MIME headers and body
	if msg.HTMLBody != "" && msg.TextBody != "" {
		// Multipart message
		boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
		b.WriteString("MIME-Version: 1.0\r\n")
		b.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		b.WriteString("\r\n")

		// Plain text part
		b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
		b.WriteString("\r\n")
		b.WriteString(msg.TextBody)
		b.WriteString("\r\n")

		// HTML part
		b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		b.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
		b.WriteString("\r\n")
		b.WriteString(msg.HTMLBody)
		b.WriteString("\r\n")

		// End boundary
		b.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if msg.HTMLBody != "" {
		// HTML only
		b.WriteString("MIME-Version: 1.0\r\n")
		b.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
		b.WriteString("\r\n")
		b.WriteString(msg.HTMLBody)
	} else {
		// Plain text only
		b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
		b.WriteString("\r\n")
		b.WriteString(msg.TextBody)
	}

	return []byte(b.String())
}

// sendMail sends the email via SMTP.
func (s *Sender) sendMail(ctx context.Context, from string, recipients []string, content []byte) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Create connection with timeout
	dialer := &net.Dialer{
		Timeout: s.config.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Start TLS if configured
	if s.config.TLS {
		tlsConfig := &tls.Config{
			ServerName:         s.config.Host,
			InsecureSkipVerify: s.config.InsecureTLS,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Authenticate if credentials provided
	if s.config.Username != "" && s.config.Password != "" {
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	if _, err := w.Write(content); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	// Quit gracefully
	if err := client.Quit(); err != nil {
		// Log but don't fail - email was sent
		return nil
	}

	return nil
}
