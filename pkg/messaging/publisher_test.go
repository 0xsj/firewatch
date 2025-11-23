package messaging_test

import (
	"context"
	"errors"
	"testing"

	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/stretchr/testify/assert"
)

func TestNoOpPublisher(t *testing.T) {
	pub := &messaging.NoOpPublisher{}

	event := messaging.NewEvent("test.event", "test", nil)
	err := pub.Publish(context.Background(), event)

	assert.NoError(t, err)
}

func TestPublisherFunc(t *testing.T) {
	var published messaging.Event

	pub := messaging.PublisherFunc(func(ctx context.Context, event messaging.Event) error {
		published = event
		return nil
	})

	event := messaging.NewEvent("test.event", "test", map[string]any{"key": "value"})
	err := pub.Publish(context.Background(), event)

	assert.NoError(t, err)
	assert.Equal(t, event.ID(), published.ID())
}

func TestPublisherFuncError(t *testing.T) {
	expectedErr := errors.New("publish failed")

	pub := messaging.PublisherFunc(func(ctx context.Context, event messaging.Event) error {
		return expectedErr
	})

	event := messaging.NewEvent("test.event", "test", nil)
	err := pub.Publish(context.Background(), event)

	assert.ErrorIs(t, err, expectedErr)
}

func TestPublisherConfig(t *testing.T) {
	config := messaging.DefaultPublisherConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 100, config.BatchSize)
	assert.Equal(t, 1000, config.BufferSize)
	assert.True(t, config.Async)

	// Apply options
	config.ApplyOptions(
		messaging.WithMaxRetries(5),
		messaging.WithBatchSize(50),
		messaging.WithAsync(false),
	)

	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 50, config.BatchSize)
	assert.False(t, config.Async)
}

func TestPublishError(t *testing.T) {
	event := messaging.NewEvent("test.event", "test", nil)
	originalErr := errors.New("connection failed")

	err := messaging.NewPublishError(event, "network error", originalErr)

	var publishErr *messaging.ErrPublishFailed
	assert.ErrorAs(t, err, &publishErr)
	assert.Equal(t, event.ID(), publishErr.EventID)
	assert.Equal(t, "test.event", publishErr.EventType)
	assert.Equal(t, "network error", publishErr.Reason)
	assert.ErrorIs(t, err, originalErr)
}
