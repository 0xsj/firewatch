package main

import (
	"fmt"
	"os"

	pkghttp "github.com/0xsj/hexagonal-go/pkg/http"
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// ========================================================================
	// Initialize Application (Wire handles all dependency injection)
	// ========================================================================

	app, cleanup, err := InitializeApp()
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}
	defer cleanup()

	app.Logger.Info("starting identity service")

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
	serverConfig.Host = "0.0.0.0"
	serverConfig.Port = getPort()

	// Print available endpoints
	printEndpoints(serverConfig.Port)

	// Start server (blocks until shutdown)
	server := pkghttp.NewServer(router, serverConfig, app.Logger)

	app.Logger.Info("starting http server")
	return server.Start()
}

// ============================================================================
// Configuration Helpers
// ============================================================================

// getPort returns the HTTP port from environment or defaults to 8080.
func getPort() int {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		return 8080
	}

	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		return 8080
	}

	return port
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
