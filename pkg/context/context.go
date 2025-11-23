package context

import (
	"context"
	"fmt"
)

// contextKey is a private type for context keys to avoid collisions.
// This prevents conflicts with other packages using context values.
type contextKey string

// Context keys for storing values in context.
const (
	// tenantIDKey stores the tenant ID for multi-tenancy.
	tenantIDKey contextKey = "tenant_id"

	// userIDKey stores the authenticated user's ID.
	userIDKey contextKey = "user_id"

	// requestIDKey stores a unique request identifier for tracing.
	requestIDKey contextKey = "request_id"

	// correlationIDKey stores a correlation ID for distributed tracing.
	correlationIDKey contextKey = "correlation_id"

	// sessionIDKey stores the session ID for authenticated requests.
	sessionIDKey contextKey = "session_id"

	// permissionsKey stores the user's permissions for authorization.
	permissionsKey contextKey = "permissions"

	// rolesKey stores the user's roles.
	rolesKey contextKey = "roles"

	// ipAddressKey stores the client's IP address.
	ipAddressKey contextKey = "ip_address"

	// userAgentKey stores the client's user agent.
	userAgentKey contextKey = "user_agent"
)

// ============================================================================
// Tenant Context (Multi-Tenancy)
// ============================================================================

// WithTenantID adds a tenant ID to the context.
// This is the primary mechanism for multi-tenancy - all data access
// should be scoped by the tenant ID extracted from context.
//
// Example:
//
//	ctx := WithTenantID(ctx, tenantID)
//	users, err := repo.List(ctx) // Automatically filtered by tenant
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// TenantID extracts the tenant ID from context.
// Returns empty string if no tenant ID is set.
//
// Example:
//
//	tenantID := TenantID(ctx)
//	if tenantID == "" {
//	    return errors.Unauthorized(op, "missing tenant context")
//	}
func TenantID(ctx context.Context) string {
	if v := ctx.Value(tenantIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// HasTenantID returns true if the context contains a tenant ID.
func HasTenantID(ctx context.Context) bool {
	return TenantID(ctx) != ""
}

// RequireTenantID extracts the tenant ID and returns an error if missing.
// Use this in repositories or services that require tenant context.
//
// Example:
//
//	tenantID, err := RequireTenantID(ctx)
//	if err != nil {
//	    return err
//	}
func RequireTenantID(ctx context.Context) (string, error) {
	tenantID := TenantID(ctx)
	if tenantID == "" {
		return "", fmt.Errorf("missing tenant context")
	}
	return tenantID, nil
}

// ============================================================================
// User Context (Authentication)
// ============================================================================

// WithUserID adds a user ID to the context.
// Set by authentication middleware after validating JWT/session.
//
// Example:
//
//	ctx := WithUserID(ctx, userID)
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserID extracts the user ID from context.
// Returns empty string if no user ID is set (unauthenticated request).
func UserID(ctx context.Context) string {
	if v := ctx.Value(userIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// HasUserID returns true if the context contains a user ID.
// Useful for checking if a request is authenticated.
func HasUserID(ctx context.Context) bool {
	return UserID(ctx) != ""
}

// IsAuthenticated returns true if the context has a user ID.
// Alias for HasUserID for better readability.
func IsAuthenticated(ctx context.Context) bool {
	return HasUserID(ctx)
}

// RequireUserID extracts the user ID and returns an error if missing.
// Use this in handlers/services that require authentication.
//
// Example:
//
//	userID, err := RequireUserID(ctx)
//	if err != nil {
//	    return errors.Unauthorized(op, "authentication required")
//	}
func RequireUserID(ctx context.Context) (string, error) {
	userID := UserID(ctx)
	if userID == "" {
		return "", fmt.Errorf("authentication required")
	}
	return userID, nil
}

// ============================================================================
// Session Context
// ============================================================================

// WithSessionID adds a session ID to the context.
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// SessionID extracts the session ID from context.
func SessionID(ctx context.Context) string {
	if v := ctx.Value(sessionIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// HasSessionID returns true if the context contains a session ID.
func HasSessionID(ctx context.Context) bool {
	return SessionID(ctx) != ""
}

// ============================================================================
// Request Tracking (Observability)
// ============================================================================

// WithRequestID adds a request ID to the context for tracing.
// The request ID should be unique per HTTP request.
//
// Example:
//
//	ctx := WithRequestID(ctx, uuid.New().String())
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestID extracts the request ID from context.
// Returns empty string if no request ID is set.
func RequestID(ctx context.Context) string {
	if v := ctx.Value(requestIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// HasRequestID returns true if the context contains a request ID.
func HasRequestID(ctx context.Context) bool {
	return RequestID(ctx) != ""
}

// WithCorrelationID adds a correlation ID for distributed tracing.
// Used to trace requests across multiple services.
//
// Example:
//
//	ctx := WithCorrelationID(ctx, correlationID)
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

// CorrelationID extracts the correlation ID from context.
func CorrelationID(ctx context.Context) string {
	if v := ctx.Value(correlationIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// HasCorrelationID returns true if the context contains a correlation ID.
func HasCorrelationID(ctx context.Context) bool {
	return CorrelationID(ctx) != ""
}

// ============================================================================
// Authorization Context
// ============================================================================

// WithPermissions adds user permissions to the context.
// Permissions are typically loaded after authentication.
//
// Example:
//
//	ctx := WithPermissions(ctx, []string{"users:read", "users:write"})
func WithPermissions(ctx context.Context, permissions []string) context.Context {
	return context.WithValue(ctx, permissionsKey, permissions)
}

// Permissions extracts permissions from context.
// Returns nil if no permissions are set.
func Permissions(ctx context.Context) []string {
	if v := ctx.Value(permissionsKey); v != nil {
		if perms, ok := v.([]string); ok {
			return perms
		}
	}
	return nil
}

// HasPermission checks if the context has a specific permission.
//
// Example:
//
//	if !HasPermission(ctx, "users:write") {
//	    return errors.Forbidden(op, "insufficient permissions")
//	}
func HasPermission(ctx context.Context, permission string) bool {
	perms := Permissions(ctx)
	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if the context has any of the specified permissions.
func HasAnyPermission(ctx context.Context, permissions ...string) bool {
	for _, permission := range permissions {
		if HasPermission(ctx, permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if the context has all of the specified permissions.
func HasAllPermissions(ctx context.Context, permissions ...string) bool {
	for _, permission := range permissions {
		if !HasPermission(ctx, permission) {
			return false
		}
	}
	return true
}

// WithRoles adds user roles to the context.
//
// Example:
//
//	ctx := WithRoles(ctx, []string{"admin", "user"})
func WithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, rolesKey, roles)
}

// Roles extracts roles from context.
func Roles(ctx context.Context) []string {
	if v := ctx.Value(rolesKey); v != nil {
		if roles, ok := v.([]string); ok {
			return roles
		}
	}
	return nil
}

// HasRole checks if the context has a specific role.
//
// Example:
//
//	if !HasRole(ctx, "admin") {
//	    return errors.Forbidden(op, "admin access required")
//	}
func HasRole(ctx context.Context, role string) bool {
	roles := Roles(ctx)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the context has any of the specified roles.
func HasAnyRole(ctx context.Context, roles ...string) bool {
	for _, role := range roles {
		if HasRole(ctx, role) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the context has all of the specified roles.
func HasAllRoles(ctx context.Context, roles ...string) bool {
	for _, role := range roles {
		if !HasRole(ctx, role) {
			return false
		}
	}
	return true
}

// ============================================================================
// Client Context (Request Metadata)
// ============================================================================

// WithIPAddress adds the client IP address to the context.
//
// Example:
//
//	ctx := WithIPAddress(ctx, "192.168.1.100")
func WithIPAddress(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ipAddressKey, ip)
}

// IPAddress extracts the client IP address from context.
func IPAddress(ctx context.Context) string {
	if v := ctx.Value(ipAddressKey); v != nil {
		if ip, ok := v.(string); ok {
			return ip
		}
	}
	return ""
}

// WithUserAgent adds the client user agent to the context.
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, userAgentKey, userAgent)
}

// UserAgent extracts the client user agent from context.
func UserAgent(ctx context.Context) string {
	if v := ctx.Value(userAgentKey); v != nil {
		if ua, ok := v.(string); ok {
			return ua
		}
	}
	return ""
}

// ============================================================================
// Composite Helpers
// ============================================================================

// WithAuth adds both user ID and session ID to the context.
// Convenience function for authentication middleware.
//
// Example:
//
//	ctx := WithAuth(ctx, userID, sessionID)
func WithAuth(ctx context.Context, userID, sessionID string) context.Context {
	ctx = WithUserID(ctx, userID)
	ctx = WithSessionID(ctx, sessionID)
	return ctx
}

// WithTenant adds tenant context with all tenant-related data.
// Convenience function for tenant middleware.
//
// Example:
//
//	ctx := WithTenant(ctx, tenantID)
func WithTenant(ctx context.Context, tenantID string) context.Context {
	return WithTenantID(ctx, tenantID)
}

// WithTracing adds request ID and correlation ID to the context.
// Convenience function for tracing middleware.
//
// Example:
//
//	ctx := WithTracing(ctx, requestID, correlationID)
func WithTracing(ctx context.Context, requestID, correlationID string) context.Context {
	ctx = WithRequestID(ctx, requestID)
	if correlationID != "" {
		ctx = WithCorrelationID(ctx, correlationID)
	}
	return ctx
}

// ============================================================================
// Validation Helpers
// ============================================================================

// ValidateAuthContext checks if the context has required authentication.
// Returns true if both tenant ID and user ID are present.
func ValidateAuthContext(ctx context.Context) bool {
	return HasTenantID(ctx) && HasUserID(ctx)
}

// ValidateTenantContext checks if the context has required tenant information.
func ValidateTenantContext(ctx context.Context) bool {
	return HasTenantID(ctx)
}

// ============================================================================
// Debug Helpers
// ============================================================================

// ContextValues returns all context values as a map for debugging/logging.
// DO NOT use this for sensitive production logging - may contain auth tokens.
//
// Example:
//
//	logger.Debug("request context", "values", ContextValues(ctx))
func ContextValues(ctx context.Context) map[string]string {
	values := make(map[string]string)

	if tid := TenantID(ctx); tid != "" {
		values["tenant_id"] = tid
	}
	if uid := UserID(ctx); uid != "" {
		values["user_id"] = uid
	}
	if sid := SessionID(ctx); sid != "" {
		values["session_id"] = sid
	}
	if rid := RequestID(ctx); rid != "" {
		values["request_id"] = rid
	}
	if cid := CorrelationID(ctx); cid != "" {
		values["correlation_id"] = cid
	}
	if ip := IPAddress(ctx); ip != "" {
		values["ip_address"] = ip
	}

	return values
}
