package messaging_test

import (
	"context"
	"errors"
	"testing"

	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/stretchr/testify/assert"
)

func TestWithName(t *testing.T) {
	baseHandler := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		return nil
	})

	handler := messaging.WithName("TestHandler", baseHandler)

	assert.Equal(t, "TestHandler", handler.Name())
}

func TestForType(t *testing.T) {
	var handled messaging.Event

	baseHandler := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		handled = event
		return nil
	})

	handler := messaging.ForType("user.registered", baseHandler)

	// Should handle matching type
	event1 := messaging.NewEvent("user.registered", "test", nil)
	err := handler.Handle(context.Background(), event1)
	assert.NoError(t, err)
	assert.Equal(t, event1.ID(), handled.ID())

	// Should skip non-matching type
	handled = nil
	event2 := messaging.NewEvent("user.deleted", "test", nil)
	err = handler.Handle(context.Background(), event2)
	assert.NoError(t, err)
	assert.Nil(t, handled)
}

func TestForTypes(t *testing.T) {
	count := 0

	baseHandler := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		count++
		return nil
	})

	handler := messaging.ForTypes(
		[]string{"user.registered", "user.verified"},
		baseHandler,
	)

	event1 := messaging.NewEvent("user.registered", "test", nil)
	event2 := messaging.NewEvent("user.verified", "test", nil)
	event3 := messaging.NewEvent("user.deleted", "test", nil)

	handler.Handle(context.Background(), event1)
	handler.Handle(context.Background(), event2)
	handler.Handle(context.Background(), event3)

	assert.Equal(t, 2, count) // Only first two
}

func TestWhen(t *testing.T) {
	var handled messaging.Event

	baseHandler := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		handled = event
		return nil
	})

	// Only process events from "identity" source
	handler := messaging.When(
		messaging.FromSource("identity"),
		baseHandler,
	)

	// Should handle
	event1 := messaging.NewEvent("user.registered", "identity", nil)
	handler.Handle(context.Background(), event1)
	assert.Equal(t, event1.ID(), handled.ID())

	// Should skip
	handled = nil
	event2 := messaging.NewEvent("order.placed", "orders", nil)
	handler.Handle(context.Background(), event2)
	assert.Nil(t, handled)
}

func TestAndPredicate(t *testing.T) {
	predicate := messaging.And(
		messaging.FromSource("identity"),
		messaging.HasMetadataKey("tenant_id"),
	)

	event1 := messaging.NewEvent("user.registered", "identity", nil).
		WithTenantID("acme")
	assert.True(t, predicate(event1))

	event2 := messaging.NewEvent("user.registered", "identity", nil)
	assert.False(t, predicate(event2))

	event3 := messaging.NewEvent("order.placed", "orders", nil).
		WithTenantID("acme")
	assert.False(t, predicate(event3))
}

func TestOrPredicate(t *testing.T) {
	predicate := messaging.Or(
		messaging.FromSource("identity"),
		messaging.FromSource("orders"),
	)

	event1 := messaging.NewEvent("user.registered", "identity", nil)
	assert.True(t, predicate(event1))

	event2 := messaging.NewEvent("order.placed", "orders", nil)
	assert.True(t, predicate(event2))

	event3 := messaging.NewEvent("payment.succeeded", "payments", nil)
	assert.False(t, predicate(event3))
}

func TestNotPredicate(t *testing.T) {
	predicate := messaging.Not(messaging.FromSource("identity"))

	event1 := messaging.NewEvent("user.registered", "identity", nil)
	assert.False(t, predicate(event1))

	event2 := messaging.NewEvent("order.placed", "orders", nil)
	assert.True(t, predicate(event2))
}

func TestMultiHandler(t *testing.T) {
	count := 0

	handler1 := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		count++
		return nil
	})

	handler2 := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		count++
		return errors.New("handler2 failed")
	})

	handler3 := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		count++
		return nil
	})

	multi := messaging.Multi(handler1, handler2, handler3)

	event := messaging.NewEvent("test.event", "test", nil)
	err := multi.Handle(context.Background(), event)

	assert.Error(t, err)      // handler2 error returned
	assert.Equal(t, 3, count) // All three handlers ran
}

func TestHandlerRegistry(t *testing.T) {
	registry := messaging.NewHandlerRegistry()

	registry.Register("user.registered", "EmailHandler")
	registry.Register("user.registered", "AnalyticsHandler")
	registry.Register("order.placed", "InventoryHandler")

	handlers := registry.GetHandlers("user.registered")
	assert.Len(t, handlers, 2)

	allHandlers := registry.GetAllHandlers()
	assert.Len(t, allHandlers, 2) // 2 event types
}
