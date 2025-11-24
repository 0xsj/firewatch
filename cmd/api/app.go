package main

import (
	auditsubscriber "github.com/0xsj/hexagonal-go/internal/audit/application/subscriber"
	v1 "github.com/0xsj/hexagonal-go/internal/identity/interface/http/v1"
	notificationsubscriber "github.com/0xsj/hexagonal-go/internal/notifications/application/subscriber"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/observability/metrics"
	"github.com/0xsj/hexagonal-go/pkg/observability/tracing"
)

// App holds all application dependencies.
type App struct {
	Logger                 logger.Logger
	DB                     database.DB
	EventBus               messaging.Publisher
	IdentityHandler        *v1.Handler
	AuditSubscriber        *auditsubscriber.EventSubscriber
	NotificationSubscriber *notificationsubscriber.UserEventSubscriber

	// Observability
	MetricsProvider metrics.Provider
	TracingProvider tracing.Provider
	HTTPMetrics     *middleware.HTTPMetrics
}
