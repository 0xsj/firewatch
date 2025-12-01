// internal/audit/provider.go
package audit

import (
	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/internal/audit/application/query"
	"github.com/0xsj/hexagonal-go/internal/audit/application/subscriber"
	"github.com/0xsj/hexagonal-go/internal/audit/domain"
	"github.com/0xsj/hexagonal-go/internal/audit/infrastructure/repository"
	v1 "github.com/0xsj/hexagonal-go/internal/audit/interface/http/v1"
)

// AuditSet provides all dependencies for the Audit domain.
var AuditSet = wire.NewSet(
	// Infrastructure - PostgreSQL Repository
	repository.NewPostgresRepository,
	wire.Bind(new(domain.Repository), new(*repository.PostgresRepository)),

	// Application - Queries
	query.NewGetEntryQuery,
	query.NewListEntriesQuery,
	query.NewGetResourceTrailQuery,
	query.NewGetActorActivityQuery,

	// Application - Event Subscriber
	subscriber.NewEventSubscriber,

	// Interface - HTTP Handler
	v1.NewHandler,
)
