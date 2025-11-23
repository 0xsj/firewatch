package messaging_test

import (
	"context"
	"errors"
	"testing"

	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/stretchr/testify/assert"
)

func TestEventHandlerFunc(t *testing.T) {
	var handled messaging.Event

	handler := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		handled = event
		return nil
	})

	event := messaging.NewEvent("test.event", "test", map[string]any{"key": "value"})
	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Equal(t, event.ID(), handled.ID())
}

func TestRecoverMiddleware(t *testing.T) {
	// Handler that panics
	panicHandler := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		panic("something went wrong")
	})

	// Wrap with recover middleware
	handler := messaging.RecoverMiddleware()(panicHandler)

	event := messaging.NewEvent("test.event", "test", nil)
	err := handler.Handle(context.Background(), event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panicked")
}

func TestChainMiddleware(t *testing.T) {
	calls := []string{}

	middleware1 := func(next messaging.EventHandler) messaging.EventHandler {
		return messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
			calls = append(calls, "before-1")
			err := next.Handle(ctx, event)
			calls = append(calls, "after-1")
			return err
		})
	}

	middleware2 := func(next messaging.EventHandler) messaging.EventHandler {
		return messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
			calls = append(calls, "before-2")
			err := next.Handle(ctx, event)
			calls = append(calls, "after-2")
			return err
		})
	}

	baseHandler := messaging.EventHandlerFunc(func(ctx context.Context, event messaging.Event) error {
		calls = append(calls, "handler")
		return nil
	})

	handler := messaging.Chain(middleware1, middleware2)(baseHandler)

	event := messaging.NewEvent("test.event", "test", nil)
	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Equal(t, []string{
		"before-1",
		"before-2",
		"handler",
		"after-2",
		"after-1",
	}, calls)
}

func TestSubscriptionError(t *testing.T) {
	originalErr := errors.New("connection failed")
	err := messaging.NewSubscriptionError("user.registered", "network error", originalErr)

	var subErr *messaging.ErrSubscriptionFailed
	assert.ErrorAs(t, err, &subErr)
	assert.Equal(t, "user.registered", subErr.EventType)
	assert.Equal(t, "network error", subErr.Reason)
	assert.ErrorIs(t, err, originalErr)
}

func TestHandlerError(t *testing.T) {
	event := messaging.NewEvent("test.event", "test", nil)
	originalErr := errors.New("database error")

	err := messaging.NewHandlerError(event, "TestHandler", originalErr)

	var handlerErr *messaging.ErrHandlerFailed
	assert.ErrorAs(t, err, &handlerErr)
	assert.Equal(t, event.ID(), handlerErr.EventID)
	assert.Equal(t, "test.event", handlerErr.EventType)
	assert.Equal(t, "TestHandler", handlerErr.Handler)
	assert.ErrorIs(t, err, originalErr)
}

func TestSubscriberConfig(t *testing.T) {
	config := messaging.DefaultSubscriberConfig()

	assert.Equal(t, 10, config.MaxConcurrency)
	assert.Equal(t, 3, config.RetryAttempts)
	assert.Len(t, config.Middlewares, 1) // RecoverMiddleware

	// Apply options
	config.ApplySubscriberOptions(
		messaging.WithMaxConcurrency(20),
		messaging.WithRetryAttempts(5),
	)

	assert.Equal(t, 20, config.MaxConcurrency)
	assert.Equal(t, 5, config.RetryAttempts)
}
