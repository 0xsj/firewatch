package domain

import (
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Notification represents an outbound notification.
type Notification struct {
	id        types.ID
	tenantID  string
	channel   Channel
	recipient string
	subject   string
	body      string
	status    Status

	// Tracking
	attempts  int
	lastError string
	sentAt    *time.Time

	// Context
	userID        string
	correlationID string
	eventType     string

	// Timestamps
	createdAt time.Time
	updatedAt time.Time
}

// NewNotification creates a new pending notification.
func NewNotification(
	tenantID string,
	channel Channel,
	recipient string,
	subject string,
	body string,
	userID string,
	correlationID string,
	eventType string,
) *Notification {
	now := time.Now().UTC()
	return &Notification{
		id:            types.NewID(),
		tenantID:      tenantID,
		channel:       channel,
		recipient:     recipient,
		subject:       subject,
		body:          body,
		status:        StatusPending,
		attempts:      0,
		userID:        userID,
		correlationID: correlationID,
		eventType:     eventType,
		createdAt:     now,
		updatedAt:     now,
	}
}

// Getters

func (n *Notification) ID() types.ID          { return n.id }
func (n *Notification) TenantID() string      { return n.tenantID }
func (n *Notification) Channel() Channel      { return n.channel }
func (n *Notification) Recipient() string     { return n.recipient }
func (n *Notification) Subject() string       { return n.subject }
func (n *Notification) Body() string          { return n.body }
func (n *Notification) Status() Status        { return n.status }
func (n *Notification) Attempts() int         { return n.attempts }
func (n *Notification) LastError() string     { return n.lastError }
func (n *Notification) SentAt() *time.Time    { return n.sentAt }
func (n *Notification) UserID() string        { return n.userID }
func (n *Notification) CorrelationID() string { return n.correlationID }
func (n *Notification) EventType() string     { return n.eventType }
func (n *Notification) CreatedAt() time.Time  { return n.createdAt }
func (n *Notification) UpdatedAt() time.Time  { return n.updatedAt }

// MarkSent marks the notification as successfully sent.
func (n *Notification) MarkSent() error {
	if n.status.IsTerminal() {
		return fmt.Errorf("notification already in terminal state: %s", n.status)
	}

	now := time.Now().UTC()
	n.status = StatusSent
	n.sentAt = &now
	n.attempts++
	n.updatedAt = now
	return nil
}

// MarkFailed marks the notification as failed.
func (n *Notification) MarkFailed(err string) error {
	if n.status == StatusSent {
		return fmt.Errorf("cannot mark sent notification as failed")
	}

	now := time.Now().UTC()
	n.status = StatusFailed
	n.lastError = err
	n.attempts++
	n.updatedAt = now
	return nil
}

// IncrementAttempt increments the attempt counter without changing status.
func (n *Notification) IncrementAttempt(err string) {
	n.attempts++
	n.lastError = err
	n.updatedAt = time.Now().UTC()
}

// CanRetry returns true if the notification can be retried.
func (n *Notification) CanRetry(maxAttempts int) bool {
	return n.status == StatusPending && n.attempts < maxAttempts
}

// Snapshot for persistence.
type Snapshot struct {
	ID            string
	TenantID      string
	Channel       string
	Recipient     string
	Subject       string
	Body          string
	Status        string
	Attempts      int
	LastError     string
	SentAt        *time.Time
	UserID        string
	CorrelationID string
	EventType     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ToSnapshot converts to a snapshot for persistence.
func (n *Notification) ToSnapshot() Snapshot {
	return Snapshot{
		ID:            n.id.String(),
		TenantID:      n.tenantID,
		Channel:       n.channel.String(),
		Recipient:     n.recipient,
		Subject:       n.subject,
		Body:          n.body,
		Status:        n.status.String(),
		Attempts:      n.attempts,
		LastError:     n.lastError,
		SentAt:        n.sentAt,
		UserID:        n.userID,
		CorrelationID: n.correlationID,
		EventType:     n.eventType,
		CreatedAt:     n.createdAt,
		UpdatedAt:     n.updatedAt,
	}
}

// FromSnapshot reconstitutes from a snapshot.
func FromSnapshot(s Snapshot) (*Notification, error) {
	id, err := types.ParseID(s.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid notification id: %w", err)
	}

	return &Notification{
		id:            id,
		tenantID:      s.TenantID,
		channel:       Channel(s.Channel),
		recipient:     s.Recipient,
		subject:       s.Subject,
		body:          s.Body,
		status:        Status(s.Status),
		attempts:      s.Attempts,
		lastError:     s.LastError,
		sentAt:        s.SentAt,
		userID:        s.UserID,
		correlationID: s.CorrelationID,
		eventType:     s.EventType,
		createdAt:     s.CreatedAt,
		updatedAt:     s.UpdatedAt,
	}, nil
}
