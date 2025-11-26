package main

import (
	auditsubscriber "github.com/0xsj/hexagonal-go/internal/audit/application/subscriber"
	emailv1 "github.com/0xsj/hexagonal-go/internal/email/interface/http/v1"
	identityv1 "github.com/0xsj/hexagonal-go/internal/identity/interface/http/v1"
	notificationsubscriber "github.com/0xsj/hexagonal-go/internal/notifications/application/subscriber"
	"github.com/0xsj/hexagonal-go/pkg/cache"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/observability/metrics"
	"github.com/0xsj/hexagonal-go/pkg/observability/tracing"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
	"github.com/0xsj/hexagonal-go/pkg/storage"
)

// App holds all application dependencies.
type App struct {
	Logger                 logger.Logger
	DB                     database.DB
	EventBus               messaging.Publisher
	IdentityHandler        *identityv1.Handler
	EmailHandler           *emailv1.Handler
	AuditSubscriber        *auditsubscriber.EventSubscriber
	NotificationSubscriber *notificationsubscriber.UserEventSubscriber

	// Security
	JWTService jwt.Service
	Cache      cache.Cache

	// Storage
	Storage storage.Storage

	// Observability
	MetricsProvider metrics.Provider
	TracingProvider tracing.Provider
	HTTPMetrics     *middleware.HTTPMetrics
}
