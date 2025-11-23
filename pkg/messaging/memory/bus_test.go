package memory_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/messaging/memory"
	"github.com/stretchr/testify/assert"
)

func TestBusPublishSubscribe(t *testing.T) {
	bus := memory.NewDefaultBus()
	defer bus.Close()

	var received messaging.Event
	var mu sync.Mutex

	handler := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		mu.Lock()
		received = event
		mu.Unlock()
		return nil
	})

	err := bus.Subscribe("test.event", handler)
	assert.NoError(t, err)

	event := messaging.NewEvent("test.event", "test", map[string]any{"key": "value"})
	err = bus.Publish(context.Background(), event)
	assert.NoError(t, err)

	// Wait for async delivery
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.NotNil(t, received)
	assert.Equal(t, event.ID(), received.ID())
	mu.Unlock()
}

func TestBusMultipleHandlers(t *testing.T) {
	bus := memory.NewDefaultBus()
	defer bus.Close()

	count := 0
	var mu sync.Mutex

	handler1 := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	})

	handler2 := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	})

	bus.Subscribe("test.event", handler1)
	bus.Subscribe("test.event", handler2)

	event := messaging.NewEvent("test.event", "test", nil)
	bus.Publish(context.Background(), event)

	// Wait for async delivery
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 2, count)
	mu.Unlock()
}

func TestBusSubscribeAll(t *testing.T) {
	bus := memory.NewDefaultBus()
	defer bus.Close()

	received := make([]messaging.Event, 0)
	var mu sync.Mutex

	handler := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		mu.Lock()
		received = append(received, event)
		mu.Unlock()
		return nil
	})

	bus.SubscribeAll(handler)

	event1 := messaging.NewEvent("event.type1", "test", nil)
	event2 := messaging.NewEvent("event.type2", "test", nil)

	bus.Publish(context.Background(), event1)
	bus.Publish(context.Background(), event2)

	// Wait for async delivery
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.Len(t, received, 2)
	mu.Unlock()
}

func TestBusHandlerError(t *testing.T) {
	var handlerErr error
	var mu sync.Mutex

	config := memory.Config{
		ErrorHandler: func(err error) {
			mu.Lock()
			handlerErr = err
			mu.Unlock()
		},
	}

	bus := memory.NewBus(config)
	defer bus.Close()

	expectedErr := errors.New("handler failed")
	handler := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		return expectedErr
	})

	bus.Subscribe("test.event", handler)

	event := messaging.NewEvent("test.event", "test", nil)
	bus.Publish(context.Background(), event)

	// Wait for async delivery
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.NotNil(t, handlerErr)
	assert.ErrorIs(t, handlerErr, expectedErr)
	mu.Unlock()
}

func TestBusUnsubscribeAll(t *testing.T) {
	bus := memory.NewDefaultBus()
	defer bus.Close()

	count := 0
	var mu sync.Mutex

	handler := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	})

	bus.Subscribe("test.event", handler)

	// Publish first event
	event1 := messaging.NewEvent("test.event", "test", nil)
	bus.Publish(context.Background(), event1)
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 1, count)
	mu.Unlock()

	// Unsubscribe all
	bus.UnsubscribeAll()

	// Publish second event
	event2 := messaging.NewEvent("test.event", "test", nil)
	bus.Publish(context.Background(), event2)
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 1, count) // Still 1, not incremented
	mu.Unlock()
}

func TestBusClose(t *testing.T) {
	bus := memory.NewDefaultBus()

	processed := false
	var mu sync.Mutex

	handler := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		time.Sleep(100 * time.Millisecond)
		mu.Lock()
		processed = true
		mu.Unlock()
		return nil
	})

	bus.Subscribe("test.event", handler)

	event := messaging.NewEvent("test.event", "test", nil)
	bus.Publish(context.Background(), event)

	// Close waits for handler to complete
	bus.Close()

	mu.Lock()
	assert.True(t, processed)
	mu.Unlock()
}

func TestBusStats(t *testing.T) {
	bus := memory.NewDefaultBus()
	defer bus.Close()

	handler1 := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		return nil
	})
	handler2 := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		return nil
	})

	bus.Subscribe("event.type1", handler1)
	bus.Subscribe("event.type2", handler1)
	bus.Subscribe("event.type2", handler2)
	bus.SubscribeAll(handler1)

	stats := bus.Stats()
	assert.Equal(t, 2, stats.EventTypes)
	assert.Equal(t, 4, stats.TotalHandlers) // 3 typed + 1 all
	assert.Equal(t, 1, stats.AllSubscribers)
	assert.False(t, stats.Closed)

	bus.Close()
	stats = bus.Stats()
	assert.True(t, stats.Closed)
}

func TestBusPublishBatch(t *testing.T) {
	bus := memory.NewDefaultBus()
	defer bus.Close()

	count := 0
	var mu sync.Mutex

	handler := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	})

	bus.Subscribe("test.event", handler)

	events := []messaging.Event{
		messaging.NewEvent("test.event", "test", nil),
		messaging.NewEvent("test.event", "test", nil),
		messaging.NewEvent("test.event", "test", nil),
	}

	err := bus.PublishBatch(context.Background(), events)
	assert.NoError(t, err)

	// Wait for async delivery
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 3, count)
	mu.Unlock()
}
