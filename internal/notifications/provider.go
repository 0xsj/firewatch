package notifications

import (
	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/internal/notifications/application/jobs"
	"github.com/0xsj/hexagonal-go/internal/notifications/application/subscriber"
	"github.com/0xsj/hexagonal-go/internal/notifications/domain"
	"github.com/0xsj/hexagonal-go/internal/notifications/infrastructure/repository"
)

// NotificationsSet provides all dependencies for the Notifications domain.
var NotificationsSet = wire.NewSet(
	// Infrastructure - PostgreSQL Repository
	repository.NewPostgresRepository,
	wire.Bind(new(domain.Repository), new(*repository.PostgresRepository)),

	// Application - Job Handlers
	jobs.NewSendEmailHandler,

	// Application - Subscribers
	subscriber.NewUserEventSubscriber,
	subscriber.NewTenantEventSubscriber,
)
