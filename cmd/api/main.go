package main

import (
	"context"
	"fmt"
	"os"

	"github.com/0xsj/hexagonal-go/cmd/api/config"
	pkghttp "github.com/0xsj/hexagonal-go/pkg/http"
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
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
	app, cleanup, err := InitializeApp(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}
	defer cleanup()

	app.Logger.Info("starting identity service")

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
		"http://localhost:3000", // React dev server
		"http://localhost:5173", // Vite dev server
		"http://localhost:8080", // Same origin
	}
	app.Logger.Info("configured cors")

	// ========================================================================
	// Create Router
	// ========================================================================
	router := app.IdentityHandler.Routes(app.Logger, corsConfig)
	app.Logger.Info("configured routes")

	// ========================================================================
	// Configure and Start Server
	// ========================================================================
	serverConfig := pkghttp.DefaultConfig()
	serverConfig.Host = cfg.Server.Host
	serverConfig.Port = cfg.Server.Port

	// Print available endpoints
	printEndpoints(serverConfig.Port)

	// Start server (blocks until shutdown)
	server := pkghttp.NewServer(router, serverConfig, app.Logger)
	app.Logger.Info("starting http server")

	return server.Start()
}

// registerSubscribers registers all event subscribers.
func registerSubscribers(app *App) error {
	// EventBus implements both Publisher and Subscriber
	subscriber, ok := app.EventBus.(messaging.Subscriber)
	if !ok {
		return fmt.Errorf("event bus does not implement Subscriber interface")
	}

	// Register audit subscriber (listens to all events)
	if err := app.AuditSubscriber.Register(subscriber); err != nil {
		return fmt.Errorf("failed to register audit subscriber: %w", err)
	}
	app.Logger.Info("registered audit subscriber")

	// Register notification subscriber (listens to user events)
	if err := app.NotificationSubscriber.Register(subscriber); err != nil {
		return fmt.Errorf("failed to register notification subscriber: %w", err)
	}
	app.Logger.Info("registered notification subscriber")

	return nil
}

// printEndpoints prints available API endpoints on startup.
func printEndpoints(port int) {
	baseURL := fmt.Sprintf("http://localhost:%d", port)

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  Identity Service - Available Endpoints")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Health Check:")
	fmt.Printf("  GET  %s/health\n", baseURL)
	fmt.Println()
	fmt.Println("User Registration & Auth:")
	fmt.Printf("  POST %s/api/v1/users/register\n", baseURL)
	fmt.Printf("  POST %s/api/v1/auth/login\n", baseURL)
	fmt.Println()
	fmt.Println("Email Verification:")
	fmt.Printf("  GET  %s/api/v1/users/verify-email?token=...\n", baseURL)
	fmt.Printf("  POST %s/api/v1/users/verify-email\n", baseURL)
	fmt.Println()
	fmt.Println("User Queries:")
	fmt.Printf("  GET  %s/api/v1/users/{id}\n", baseURL)
	fmt.Printf("  GET  %s/api/v1/users\n", baseURL)
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println()
}
