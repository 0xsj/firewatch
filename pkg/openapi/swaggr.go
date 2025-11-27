package openapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// SwaggerConfig holds configuration for Swagger UI.
type SwaggerConfig struct {
	// BasePath is the base path for the Swagger UI (default: "/swagger")
	BasePath string

	// Title is the page title shown in browser tab
	Title string

	// DeepLinking enables deep linking for tags and operations
	DeepLinking bool

	// DocExpansion controls how operations are displayed: "list", "full", or "none"
	DocExpansion string

	// PersistAuthorization persists authorization data in localStorage
	PersistAuthorization bool
}

// DefaultSwaggerConfig returns sensible defaults for Swagger UI.
func DefaultSwaggerConfig() SwaggerConfig {
	return SwaggerConfig{
		BasePath:             "/swagger",
		Title:                "API Documentation",
		DeepLinking:          true,
		DocExpansion:         "list",
		PersistAuthorization: true,
	}
}

// RegisterSwaggerRoutes registers Swagger UI routes on the given router.
func RegisterSwaggerRoutes(r chi.Router, cfg SwaggerConfig) {
	if cfg.BasePath == "" {
		cfg.BasePath = "/swagger"
	}

	r.Get(cfg.BasePath+"/*", httpSwagger.Handler(
		httpSwagger.URL(cfg.BasePath+"/doc.json"),
		httpSwagger.DeepLinking(cfg.DeepLinking),
		httpSwagger.DocExpansion(cfg.DocExpansion),
		httpSwagger.PersistAuthorization(cfg.PersistAuthorization),
	))
}

// Handler returns a http.Handler for Swagger UI.
// Use this if you need more control over mounting.
func Handler(cfg SwaggerConfig) http.Handler {
	return httpSwagger.Handler(
		httpSwagger.URL(cfg.BasePath+"/doc.json"),
		httpSwagger.DeepLinking(cfg.DeepLinking),
		httpSwagger.DocExpansion(cfg.DocExpansion),
		httpSwagger.PersistAuthorization(cfg.PersistAuthorization),
	)
}