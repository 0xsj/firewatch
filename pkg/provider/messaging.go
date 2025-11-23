package provider

import (
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/messaging/memory"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// ProvideEventBus creates an in-memory event bus.
// This acts as both Publisher and Subscriber.
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

// ProvideEventSubscriber creates an event subscriber (same as event bus).
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
