package main

import (
	auditsubscriber "github.com/0xsj/hexagonal-go/internal/audit/application/subscriber"
	"github.com/0xsj/hexagonal-go/internal/demo"
	emailv1 "github.com/0xsj/hexagonal-go/internal/email/interface/http/v1"
	"github.com/0xsj/hexagonal-go/internal/flags/interface/http/admin"
	flagsv1 "github.com/0xsj/hexagonal-go/internal/flags/interface/http/v1"
	identityv1 "github.com/0xsj/hexagonal-go/internal/identity/interface/http/v1"
	notificationjobs "github.com/0xsj/hexagonal-go/internal/notifications/application/jobs"
	notificationsubscriber "github.com/0xsj/hexagonal-go/internal/notifications/application/subscriber"
	permissionsv1 "github.com/0xsj/hexagonal-go/internal/permissions/interface/http/v1"
	tenantv1 "github.com/0xsj/hexagonal-go/internal/tenant/interface/http/v1"
	"github.com/0xsj/hexagonal-go/pkg/cache"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/flags"
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/observability/metrics"
	"github.com/0xsj/hexagonal-go/pkg/observability/tracing"
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
	"github.com/0xsj/hexagonal-go/pkg/storage"
	"github.com/0xsj/hexagonal-go/pkg/worker"
)

// App holds all application dependencies.
type App struct {
	Logger   logger.Logger
	DB       database.DB
	EventBus messaging.Publisher

	// Domain Handlers
	IdentityHandler    *identityv1.Handler
	TenantHandler      *tenantv1.Handler
	EmailHandler       *emailv1.Handler
	FlagsHandler       *flagsv1.Handler
	FlagsAdminHandler  *admin.Handler
	PermissionsHandler *permissionsv1.Handler

	// Demo Handler
	DemoHandler *demo.Handler

	// Feature Flags Client (SDK)
	FlagsClient flags.Client

	// Subscribers
	AuditSubscriber              *auditsubscriber.EventSubscriber
	UserNotificationSubscriber   *notificationsubscriber.UserEventSubscriber
	TenantNotificationSubscriber *notificationsubscriber.TenantEventSubscriber

	// Job Queue & Handlers
	JobQueue            worker.Queue
	SendEmailJobHandler *notificationjobs.SendEmailHandler

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