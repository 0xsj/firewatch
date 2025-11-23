package messaging_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {
	event := messaging.NewEvent(
		"identity.user.registered",
		"identity",
		map[string]any{
			"user_id": "123",
			"email":   "test@example.com",
		},
	)

	assert.NotEmpty(t, event.ID())
	assert.Equal(t, "identity.user.registered", event.Type())
	assert.Equal(t, "identity", event.Source())
	assert.Equal(t, "test@example.com", event.Data()["email"])
	assert.WithinDuration(t, time.Now(), event.Timestamp(), time.Second)
}

func TestEventMetadata(t *testing.T) {
	event := messaging.NewEvent("test.event", "test", nil).
		WithCorrelationID("corr-123").
		WithTenantID("tenant-1").
		WithUserID("user-1")

	assert.Equal(t, "corr-123", event.Metadata()["correlation_id"])
	assert.Equal(t, "tenant-1", event.Metadata()["tenant_id"])
	assert.Equal(t, "user-1", event.Metadata()["user_id"])
}

func TestEventSerialization(t *testing.T) {
	original := messaging.NewEvent(
		"test.event",
		"test",
		map[string]any{"key": "value"},
	).WithCorrelationID("abc-123")

	// Marshal
	data, err := json.Marshal(original)
	assert.NoError(t, err)

	// Unmarshal
	parsed, err := messaging.ParseEvent(data)
	assert.NoError(t, err)

	assert.Equal(t, original.ID(), parsed.ID())
	assert.Equal(t, original.Type(), parsed.Type())
	assert.Equal(t, original.Source(), parsed.Source())
	assert.Equal(t, "value", parsed.Data()["key"])
	assert.Equal(t, "abc-123", parsed.Metadata()["correlation_id"])
}
