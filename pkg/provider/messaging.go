package provider

import (
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/messaging/memory"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// ProvideEventBus creates an in-memory event bus.
// This acts as both Publisher and Subscriber.
//
// Backward compatible: existing commands continue to receive messaging.Publisher.
func ProvideEventBus(log logger.Logger) messaging.Publisher {
	config := memory.Config{
		Logger: log,
		ErrorHandler: func(err error) {
			log.Error("event handler error", logger.Err(err))
		},
	}

	bus := memory.NewBus(config)
	return bus
}

// ProvideEventSubscriber creates an event subscriber (same bus instance).
//
// Note: In production, you may want to share the same bus instance.
// For now, this creates a separate instance for simplicity.
func ProvideEventSubscriber(log logger.Logger) messaging.Subscriber {
	config := memory.Config{
		Logger: log,
		ErrorHandler: func(err error) {
			log.Error("event handler error", logger.Err(err))
		},
	}

	bus := memory.NewBus(config)
	return bus
}

// ProvideDomainEventPublisher creates a DomainEventPublisher.
// This is the shared service that commands can use to publish domain events
// with consistent metadata handling.
//
// Commands can gradually migrate from using messaging.Publisher directly
// to using DomainEventPublisher for cleaner, deduplicated event publishing.
func ProvideDomainEventPublisher(
	publisher messaging.Publisher,
	log logger.Logger,
) *messaging.DomainEventPublisher {
	return messaging.NewDomainEventPublisher(publisher, log)
}
