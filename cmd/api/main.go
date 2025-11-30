// @title           Hexagonal Go API
// @version         1.0
// @description     Multi-tenant API built with hexagonal architecture. Provides identity management, tenant management, email templates, and more.

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your bearer token in the format: Bearer <token>

// @tag.name identity
// @tag.description User registration, authentication, and session management

// @tag.name tenants
// @tag.description Tenant lifecycle management

// @tag.name email
// @tag.description Email template management

// @tag.name permissions
// @tag.description Role-based access control and permission management

// @tag.name system
// @tag.description System endpoints (health checks, etc.)

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/hexagonal-go/cmd/api/config"
	"github.com/0xsj/hexagonal-go/internal/demo"
	emailv1 "github.com/0xsj/hexagonal-go/internal/email/interface/http/v1"
	"github.com/0xsj/hexagonal-go/internal/flags/interface/http/admin"
	flagsv1 "github.com/0xsj/hexagonal-go/internal/flags/interface/http/v1"
	tenantv1 "github.com/0xsj/hexagonal-go/internal/tenant/interface/http/v1"
	pkghttp "github.com/0xsj/hexagonal-go/pkg/http"
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/openapi"

	_ "github.com/0xsj/hexagonal-go/docs/swagger" // swagger docs
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	// ========================================================================
	// Load Configuration
	// ========================================================================
	cfg, err := config.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// ========================================================================
	// Initialize Application (Wire handles all dependency injection)
	// ========================================================================
	app, cleanup, err := InitializeApp(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}
	defer cleanup()

	app.Logger.Info("starting identity service")

	// ========================================================================
	// Start Metrics Server
	// ========================================================================
	if err := app.MetricsProvider.Start(ctx); err != nil {
		return fmt.Errorf("failed to start metrics server: %w", err)
	}
	defer app.MetricsProvider.Close()
	app.Logger.Info("metrics server started",
		logger.Int("port", cfg.Metrics.Port),
		logger.String("path", cfg.Metrics.Path),
	)

	// ========================================================================
	// Register Event Subscribers
	// ========================================================================
	if err := registerSubscribers(app); err != nil {
		return fmt.Errorf("failed to register subscribers: %w", err)
	}

	// ========================================================================
	// Configure CORS
	// ========================================================================
	corsConfig := middleware.DefaultCORSConfig()
	corsConfig.AllowedOrigins = []string{
		"http://localhost:3000",
		"http://localhost:5173",
		"http://localhost:8080",
	}
	app.Logger.Info("configured cors")

	// ========================================================================
	// Create Router with Observability Middleware
	// ========================================================================
	root := chi.NewRouter()

	root.Use(middleware.Tracing(app.TracingProvider.Tracer()))
	root.Use(middleware.Metrics(app.HTTPMetrics))

	// ========================================================================
	// Mount Swagger UI
	// ========================================================================
	swaggerConfig := openapi.DefaultSwaggerConfig()
	swaggerConfig.Title = "Hexagonal Go API"
	openapi.RegisterSwaggerRoutes(root, swaggerConfig)
	app.Logger.Info("swagger ui available at /swagger/index.html")

	// ========================================================================
	// Mount Module Routes
	// ========================================================================

	// Mount email routes FIRST (more specific path)
	emailRouter := emailv1.NewRouter(app.EmailHandler, app.JWTService, app.Logger, corsConfig)
	root.Mount("/api/v1/email", emailRouter)

	tenantRouter := tenantv1.NewRouter(app.TenantHandler, app.JWTService, app.Logger, corsConfig)
	root.Mount("/api/v1/tenants", tenantRouter)

	// Mount flags routes
	flagsRouter := flagsv1.NewRouter(app.FlagsHandler, app.JWTService, app.Logger, corsConfig)
	root.Mount("/api/v1/flags", flagsRouter)

	// Mount permissions routes
	permissionsRouter := app.PermissionsHandler.Routes(app.Logger, corsConfig, app.JWTService)
	root.Mount("", permissionsRouter)

	// Mount demo routes
	demoRouter := demo.NewRouter(app.DemoHandler, app.Logger)
	root.Mount("/demo", demoRouter)

	// Mount flags admin dashboard (public for development)
	flagsAdminRouter := admin.NewPublicRouter(app.FlagsAdminHandler, app.Logger)
	root.Mount("/admin/flags", flagsAdminRouter)

	// Mount identity routes (includes /health and /api/v1/users, /api/v1/auth, etc.)
	identityRouter := app.IdentityHandler.Routes(app.Logger, corsConfig, app.JWTService)
	root.Mount("/", identityRouter)

	app.Logger.Info("configured routes with observability and authentication")

	// ========================================================================
	// Configure and Start Server
	// ========================================================================
	serverConfig := pkghttp.DefaultConfig()
	serverConfig.Host = cfg.Server.Host
	serverConfig.Port = cfg.Server.Port

	printEndpoints(serverConfig.Port, cfg.Metrics.Port)

	defer func() {
		if err := app.TracingProvider.Shutdown(ctx); err != nil {
			app.Logger.Error("failed to shutdown tracing", logger.Err(err))
		}
	}()

	server := pkghttp.NewServer(root, serverConfig, app.Logger)
	app.Logger.Info("starting http server")

	return server.Start()
}

// registerSubscribers registers all event subscribers.
func registerSubscribers(app *App) error {
	subscriber, ok := app.EventBus.(messaging.Subscriber)
	if !ok {
		return fmt.Errorf("event bus does not implement Subscriber interface")
	}

	if err := app.AuditSubscriber.Register(subscriber); err != nil {
		return fmt.Errorf("failed to register audit subscriber: %w", err)
	}
	app.Logger.Info("registered audit subscriber")

	if err := app.UserNotificationSubscriber.Register(subscriber); err != nil {
		return fmt.Errorf("failed to register user notification subscriber: %w", err)
	}
	app.Logger.Info("registered user notification subscriber")

	if err := app.TenantNotificationSubscriber.Register(subscriber); err != nil {
		return fmt.Errorf("failed to register tenant notification subscriber: %w", err)
	}
	app.Logger.Info("registered tenant notification subscriber")

	return nil
}

// printEndpoints prints available API endpoints on startup.
func printEndpoints(port, metricsPort int) {
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	metricsURL := fmt.Sprintf("http://localhost:%d", metricsPort)

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  Identity Service - Available Endpoints")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Documentation:")
	fmt.Printf("  GET  %s/swagger/index.html\n", baseURL)
	fmt.Println()
	fmt.Println("Health Check:")
	fmt.Printf("  GET  %s/health\n", baseURL)
	fmt.Println()
	fmt.Println("Public - User Registration & Auth:")
	fmt.Printf("  POST %s/api/v1/users/register\n", baseURL)
	fmt.Printf("  POST %s/api/v1/auth/login\n", baseURL)
	fmt.Printf("  POST %s/api/v1/auth/refresh\n", baseURL)
	fmt.Println()
	fmt.Println("Public - Email Verification:")
	fmt.Printf("  GET  %s/api/v1/users/verify-email?token=...\n", baseURL)
	fmt.Printf("  POST %s/api/v1/users/verify-email\n", baseURL)
	fmt.Println()
	fmt.Println("Protected - Auth (requires JWT):")
	fmt.Printf("  POST %s/api/v1/auth/logout\n", baseURL)
	fmt.Println()
	fmt.Println("Protected - User Queries (requires JWT):")
	fmt.Printf("  GET  %s/api/v1/users/me\n", baseURL)
	fmt.Printf("  GET  %s/api/v1/users/{id}\n", baseURL)
	fmt.Printf("  GET  %s/api/v1/users\n", baseURL)
	fmt.Println()
	fmt.Println("Protected - Sessions (requires JWT):")
	fmt.Printf("  GET  %s/api/v1/sessions\n", baseURL)
	fmt.Println()
	fmt.Println("Protected - Tenants (requires JWT + Admin):")
	fmt.Printf("  GET  %s/api/v1/tenants\n", baseURL)
	fmt.Printf("  POST %s/api/v1/tenants\n", baseURL)
	fmt.Printf("  GET  %s/api/v1/tenants/{id}\n", baseURL)
	fmt.Printf("  GET  %s/api/v1/tenants/slug/{slug}\n", baseURL)
	fmt.Printf("  PATCH %s/api/v1/tenants/{id}\n", baseURL)
	fmt.Printf("  DELETE %s/api/v1/tenants/{id}\n", baseURL)
	fmt.Printf("  PUT  %s/api/v1/tenants/{id}/settings\n", baseURL)
	fmt.Printf("  POST %s/api/v1/tenants/{id}/plan\n", baseURL)
	fmt.Printf("  POST %s/api/v1/tenants/{id}/suspend\n", baseURL)
	fmt.Printf("  POST %s/api/v1/tenants/{id}/reactivate\n", baseURL)
	fmt.Println()
	fmt.Println("Protected - Email Templates (requires JWT):")
	fmt.Printf("  GET  %s/api/v1/email/templates\n", baseURL)
	fmt.Printf("  POST %s/api/v1/email/templates\n", baseURL)
	fmt.Printf("  GET  %s/api/v1/email/templates/{id}\n", baseURL)
	fmt.Printf("  PUT  %s/api/v1/email/templates/{id}\n", baseURL)
	fmt.Printf("  DELETE %s/api/v1/email/templates/{id}\n", baseURL)
	fmt.Printf("  POST %s/api/v1/email/templates/{id}/activate\n", baseURL)
	fmt.Printf("  POST %s/api/v1/email/templates/{id}/archive\n", baseURL)
	fmt.Printf("  POST %s/api/v1/email/templates/{id}/preview\n", baseURL)
	fmt.Printf("  GET  %s/api/v1/email/templates/by-slug?slug=...&locale=...\n", baseURL)
	fmt.Printf("  POST %s/api/v1/email/templates/preview-by-slug\n", baseURL)
	fmt.Println()
	fmt.Println("Protected - Feature Flags (requires JWT):")
	fmt.Printf("  GET  %s/api/v1/flags\n", baseURL)
	fmt.Printf("  POST %s/api/v1/flags\n", baseURL)
	fmt.Printf("  GET  %s/api/v1/flags/{id}\n", baseURL)
	fmt.Printf("  PUT  %s/api/v1/flags/{id}\n", baseURL)
	fmt.Printf("  DELETE %s/api/v1/flags/{id}\n", baseURL)
	fmt.Printf("  POST %s/api/v1/flags/{id}/enable\n", baseURL)
	fmt.Printf("  POST %s/api/v1/flags/{id}/disable\n", baseURL)
	fmt.Printf("  GET  %s/api/v1/flags/by-key?key=...\n", baseURL)
	fmt.Printf("  POST %s/api/v1/flags/{id}/rules\n", baseURL)
	fmt.Printf("  DELETE %s/api/v1/flags/{id}/rules/{ruleId}\n", baseURL)
	fmt.Printf("  POST %s/api/v1/flags/{id}/overrides\n", baseURL)
	fmt.Printf("  DELETE %s/api/v1/flags/{id}/overrides\n", baseURL)
	fmt.Printf("  POST %s/api/v1/flags/{key}/evaluate\n", baseURL)
	fmt.Println()
	fmt.Println("Protected - Permissions (requires JWT):")
	fmt.Printf("  GET  %s/api/v1/permissions/me\n", baseURL)
	fmt.Printf("  GET  %s/api/v1/permissions/check?action=...&resource=...\n", baseURL)
	fmt.Printf("  GET  %s/api/v1/roles\n", baseURL)
	fmt.Printf("  GET  %s/api/v1/roles/{id}\n", baseURL)
	fmt.Println()
	fmt.Println("Protected - Permissions Admin (requires JWT + Admin):")
	fmt.Printf("  POST %s/api/v1/roles\n", baseURL)
	fmt.Printf("  PUT  %s/api/v1/roles/{id}\n", baseURL)
	fmt.Printf("  DELETE %s/api/v1/roles/{id}\n", baseURL)
	fmt.Printf("  POST %s/api/v1/assignments\n", baseURL)
	fmt.Printf("  POST %s/api/v1/assignments/revoke\n", baseURL)
	fmt.Println()
	fmt.Println("Admin Dashboard:")
	fmt.Printf("  GET  %s/admin/flags\n", baseURL)
	fmt.Printf("  GET  %s/admin/flags/new\n", baseURL)
	fmt.Printf("  GET  %s/admin/flags/{id}\n", baseURL)
	fmt.Printf("  GET  %s/admin/flags/{id}/edit\n", baseURL)
	fmt.Println()
	fmt.Println()
	fmt.Println("Observability:")
	fmt.Printf("  GET  %s/metrics\n", metricsURL)
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println()
}