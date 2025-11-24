package notifications

import (
	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/internal/notifications/application/command"
	"github.com/0xsj/hexagonal-go/internal/notifications/application/subscriber"
	"github.com/0xsj/hexagonal-go/internal/notifications/domain"
	"github.com/0xsj/hexagonal-go/internal/notifications/infrastructure/repository"
)

// NotificationsSet provides all dependencies for the Notifications domain.
var NotificationsSet = wire.NewSet(
	// Infrastructure - PostgreSQL Repository
	repository.NewPostgresRepository,
	wire.Bind(new(domain.Repository), new(*repository.PostgresRepository)),

	// Application - Commands
	command.NewSendNotificationCommand,

	// Application - Subscribers
	subscriber.NewUserEventSubscriber,
)
