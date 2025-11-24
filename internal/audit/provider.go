package audit

import (
	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/internal/audit/application/subscriber"
	"github.com/0xsj/hexagonal-go/internal/audit/domain"
	"github.com/0xsj/hexagonal-go/internal/audit/infrastructure/repository"
)

// AuditSet provides all dependencies for the Audit domain.
var AuditSet = wire.NewSet(
	// Infrastructure - PostgreSQL Repository
	repository.NewPostgresRepository,
	wire.Bind(new(domain.Repository), new(*repository.PostgresRepository)),

	// Application - Event Subscriber
	subscriber.NewEventSubscriber,
)
