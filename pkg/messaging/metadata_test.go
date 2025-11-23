package messaging_test

import (
	"context"
	"testing"

	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/stretchr/testify/assert"
)

func TestContextExtraction(t *testing.T) {
	ctx := context.Background()
	ctx = messaging.WithCorrelationID(ctx, "corr-123")
	ctx = messaging.WithTenantID(ctx, "tenant-1")
	ctx = messaging.WithUserID(ctx, "user-1")

	assert.Equal(t, "corr-123", messaging.ExtractCorrelationID(ctx))
	assert.Equal(t, "tenant-1", messaging.ExtractTenantID(ctx))
	assert.Equal(t, "user-1", messaging.ExtractUserID(ctx))
}

func TestContextExtractionMissing(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "", messaging.ExtractCorrelationID(ctx))
	assert.Equal(t, "", messaging.ExtractTenantID(ctx))
	assert.Equal(t, "", messaging.ExtractUserID(ctx))
}

func TestEnrichFromContext(t *testing.T) {
	ctx := context.Background()
	ctx = messaging.WithCorrelationID(ctx, "corr-123")
	ctx = messaging.WithTenantID(ctx, "tenant-1")
	ctx = messaging.WithUserID(ctx, "user-1")
	ctx = messaging.WithIPAddress(ctx, "192.168.1.1")

	event := messaging.NewEvent("test.event", "test", nil)
	event = messaging.EnrichFromContext(ctx, event)

	assert.Equal(t, "corr-123", messaging.GetCorrelationID(event))
	assert.Equal(t, "tenant-1", messaging.GetTenantID(event))
	assert.Equal(t, "user-1", messaging.GetUserID(event))
	assert.Equal(t, "192.168.1.1", messaging.GetIPAddress(event))
}

func TestNewEventFromContext(t *testing.T) {
	ctx := context.Background()
	ctx = messaging.WithCorrelationID(ctx, "corr-123")
	ctx = messaging.WithTenantID(ctx, "tenant-1")

	event := messaging.NewEventFromContext(
		ctx,
		"test.event",
		"test",
		map[string]any{"key": "value"},
	)

	assert.Equal(t, "test.event", event.Type())
	assert.Equal(t, "test", event.Source())
	assert.Equal(t, "value", event.Data()["key"])
	assert.Equal(t, "corr-123", messaging.GetCorrelationID(event))
	assert.Equal(t, "tenant-1", messaging.GetTenantID(event))
}

func TestPropagateToContext(t *testing.T) {
	event := messaging.NewEvent("test.event", "test", nil).
		WithCorrelationID("corr-123").
		WithTenantID("tenant-1").
		WithUserID("user-1")

	ctx := messaging.PropagateToContext(context.Background(), event)

	assert.Equal(t, "corr-123", messaging.ExtractCorrelationID(ctx))
	assert.Equal(t, "tenant-1", messaging.ExtractTenantID(ctx))
	assert.Equal(t, "user-1", messaging.ExtractUserID(ctx))
}

func TestGetMetadataTyped(t *testing.T) {
	event := messaging.NewEvent("test.event", "test", nil).
		WithMetadata("string_key", "string_value").
		WithMetadata("int_key", 42).
		WithMetadata("bool_key", true)

	strVal, ok := messaging.GetMetadataString(event, "string_key")
	assert.True(t, ok)
	assert.Equal(t, "string_value", strVal)

	intVal, ok := messaging.GetMetadataInt(event, "int_key")
	assert.True(t, ok)
	assert.Equal(t, 42, intVal)

	boolVal, ok := messaging.GetMetadataBool(event, "bool_key")
	assert.True(t, ok)
	assert.True(t, boolVal)

	// Wrong type
	_, ok = messaging.GetMetadataString(event, "int_key")
	assert.False(t, ok)
}

func TestHasMetadata(t *testing.T) {
	event := messaging.NewEvent("test.event", "test", nil).
		WithTenantID("tenant-1")

	assert.True(t, messaging.HasMetadata(event, "tenant_id"))
	assert.False(t, messaging.HasMetadata(event, "missing_key"))
}
