package messaging

import (
	"context"
)

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

const (
	// ContextKeyCorrelationID is the context key for correlation ID.
	ContextKeyCorrelationID ContextKey = "correlation_id"

	// ContextKeyTenantID is the context key for tenant ID.
	ContextKeyTenantID ContextKey = "tenant_id"

	// ContextKeyUserID is the context key for user ID.
	ContextKeyUserID ContextKey = "user_id"

	// ContextKeyRequestID is the context key for request ID.
	ContextKeyRequestID ContextKey = "request_id"

	// ContextKeyIPAddress is the context key for IP address.
	ContextKeyIPAddress ContextKey = "ip_address"

	// ContextKeyUserAgent is the context key for user agent.
	ContextKeyUserAgent ContextKey = "user_agent"
)

// Metadata helper constants
const (
	MetadataCorrelationID = "correlation_id"
	MetadataCausationID   = "causation_id"
	MetadataTenantID      = "tenant_id"
	MetadataUserID        = "user_id"
	MetadataIPAddress     = "ip_address"
	MetadataUserAgent     = "user_agent"
	MetadataSource        = "source"
	MetadataVersion       = "version"
)

// ============================================================================
// Context Extraction
// ============================================================================

// ExtractCorrelationID extracts the correlation ID from context.
// Returns empty string if not found.
func ExtractCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(ContextKeyCorrelationID).(string); ok {
		return id
	}
	return ""
}

// ExtractTenantID extracts the tenant ID from context.
// Returns empty string if not found.
func ExtractTenantID(ctx context.Context) string {
	if id, ok := ctx.Value(ContextKeyTenantID).(string); ok {
		return id
	}
	return ""
}

// ExtractUserID extracts the user ID from context.
// Returns empty string if not found.
func ExtractUserID(ctx context.Context) string {
	if id, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return id
	}
	return ""
}

// ExtractRequestID extracts the request ID from context.
// Returns empty string if not found.
func ExtractRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		return id
	}
	return ""
}

// ExtractIPAddress extracts the IP address from context.
// Returns empty string if not found.
func ExtractIPAddress(ctx context.Context) string {
	if ip, ok := ctx.Value(ContextKeyIPAddress).(string); ok {
		return ip
	}
	return ""
}

// ExtractUserAgent extracts the user agent from context.
// Returns empty string if not found.
func ExtractUserAgent(ctx context.Context) string {
	if ua, ok := ctx.Value(ContextKeyUserAgent).(string); ok {
		return ua
	}
	return ""
}

// ============================================================================
// Context Injection
// ============================================================================

// WithCorrelationID adds correlation ID to context.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, ContextKeyCorrelationID, correlationID)
}

// WithTenantID adds tenant ID to context.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, ContextKeyTenantID, tenantID)
}

// WithUserID adds user ID to context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// WithRequestID adds request ID to context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}

// WithIPAddress adds IP address to context.
func WithIPAddress(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ContextKeyIPAddress, ip)
}

// WithUserAgent adds user agent to context.
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, ContextKeyUserAgent, userAgent)
}

// ============================================================================
// Event Metadata Enrichment
// ============================================================================

// EnrichFromContext enriches an event with metadata from context.
// Extracts correlation ID, tenant ID, user ID, etc.
//
// Example:
//
//	event := messaging.NewEvent("user.registered", "identity", data)
//	event = messaging.EnrichFromContext(ctx, event)
func EnrichFromContext(ctx context.Context, event *BaseEvent) *BaseEvent {
	if correlationID := ExtractCorrelationID(ctx); correlationID != "" {
		event.WithCorrelationID(correlationID)
	}

	if tenantID := ExtractTenantID(ctx); tenantID != "" {
		event.WithTenantID(tenantID)
	}

	if userID := ExtractUserID(ctx); userID != "" {
		event.WithUserID(userID)
	}

	if requestID := ExtractRequestID(ctx); requestID != "" {
		event.WithMetadata("request_id", requestID)
	}

	if ip := ExtractIPAddress(ctx); ip != "" {
		event.WithIPAddress(ip)
	}

	if ua := ExtractUserAgent(ctx); ua != "" {
		event.WithMetadata("user_agent", ua)
	}

	return event
}

// NewEventFromContext creates a new event and enriches it with context metadata.
//
// Example:
//
//	event := messaging.NewEventFromContext(
//	    ctx,
//	    "user.registered",
//	    "identity",
//	    map[string]any{"user_id": "123"},
//	)
func NewEventFromContext(ctx context.Context, eventType, source string, data map[string]any) *BaseEvent {
	event := NewEvent(eventType, source, data)
	return EnrichFromContext(ctx, event)
}

// ============================================================================
// Context Propagation
// ============================================================================

// PropagateToContext creates a new context with metadata from event.
// Useful for passing event context to handlers.
//
// Example:
//
//	ctx := messaging.PropagateToContext(context.Background(), event)
//	// Now ctx contains correlation_id, tenant_id, etc. from event
func PropagateToContext(ctx context.Context, event Event) context.Context {
	metadata := event.Metadata()

	if correlationID, ok := metadata[MetadataCorrelationID].(string); ok {
		ctx = WithCorrelationID(ctx, correlationID)
	}

	if tenantID, ok := metadata[MetadataTenantID].(string); ok {
		ctx = WithTenantID(ctx, tenantID)
	}

	if userID, ok := metadata[MetadataUserID].(string); ok {
		ctx = WithUserID(ctx, userID)
	}

	if requestID, ok := metadata["request_id"].(string); ok {
		ctx = WithRequestID(ctx, requestID)
	}

	if ip, ok := metadata[MetadataIPAddress].(string); ok {
		ctx = WithIPAddress(ctx, ip)
	}

	if ua, ok := metadata["user_agent"].(string); ok {
		ctx = WithUserAgent(ctx, ua)
	}

	return ctx
}

// ============================================================================
// Metadata Extraction from Event
// ============================================================================

// GetCorrelationID extracts correlation ID from event metadata.
func GetCorrelationID(event Event) string {
	if id, ok := event.Metadata()[MetadataCorrelationID].(string); ok {
		return id
	}
	return ""
}

// GetTenantID extracts tenant ID from event metadata.
func GetTenantID(event Event) string {
	if id, ok := event.Metadata()[MetadataTenantID].(string); ok {
		return id
	}
	return ""
}

// GetUserID extracts user ID from event metadata.
func GetUserID(event Event) string {
	if id, ok := event.Metadata()[MetadataUserID].(string); ok {
		return id
	}
	return ""
}

// GetCausationID extracts causation ID from event metadata.
func GetCausationID(event Event) string {
	if id, ok := event.Metadata()[MetadataCausationID].(string); ok {
		return id
	}
	return ""
}

// GetIPAddress extracts IP address from event metadata.
func GetIPAddress(event Event) string {
	if ip, ok := event.Metadata()[MetadataIPAddress].(string); ok {
		return ip
	}
	return ""
}

// ============================================================================
// Metadata Helpers
// ============================================================================

// HasMetadata checks if event has a specific metadata key.
func HasMetadata(event Event, key string) bool {
	_, ok := event.Metadata()[key]
	return ok
}

// GetMetadata gets metadata value with type assertion.
func GetMetadata(event Event, key string) (any, bool) {
	val, ok := event.Metadata()[key]
	return val, ok
}

// GetMetadataString gets string metadata value.
func GetMetadataString(event Event, key string) (string, bool) {
	if val, ok := event.Metadata()[key].(string); ok {
		return val, true
	}
	return "", false
}

// GetMetadataInt gets int metadata value.
func GetMetadataInt(event Event, key string) (int, bool) {
	if val, ok := event.Metadata()[key].(int); ok {
		return val, true
	}
	return 0, false
}

// GetMetadataBool gets bool metadata value.
func GetMetadataBool(event Event, key string) (bool, bool) {
	if val, ok := event.Metadata()[key].(bool); ok {
		return val, true
	}
	return false, false
}
