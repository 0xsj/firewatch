package errors

import (
	"fmt"
	"maps"
	"sort"
	"sync"
)

// ============================================================================
// Error Registry
// ============================================================================

// Registry provides centralized error code registration and lookup.
// This enables:
//   - Stable error codes across service versions
//   - Documentation generation from registered errors
//   - Consistent error creation with ByCode()
//   - Validation that codes are registered before use
//
// Thread-safe for concurrent access.
type Registry struct {
	mu     sync.RWMutex
	errors map[Code]*RegisteredError
}

// RegisteredError represents an error code registered in the system.
// It contains all default values for errors of this code.
type RegisteredError struct {
	// Code is the unique error identifier
	Code Code

	// Kind is the error category
	Kind Kind

	// Template is the default message template.
	// Can contain placeholders for fmt.Sprintf if needed.
	Template string

	// Severity is the default severity level
	Severity Severity

	// Retryable indicates if errors of this type can be retried by default
	Retryable bool

	// HTTPStatus is the default HTTP status code (optional, 0 means use kind default)
	HTTPStatus int

	// GRPCCode is the default gRPC status code (optional, 0 means use kind default)
	GRPCCode int

	// Metadata contains additional registration metadata
	// (e.g., documentation, deprecation info, owner team)
	Metadata map[string]any
}

// Global registry instance
var globalRegistry = &Registry{
	errors: make(map[Code]*RegisteredError),
}

// ============================================================================
// Registration
// ============================================================================

// Register registers an error code with its default configuration.
// Panics if the code is already registered or invalid.
//
// Example:
//
//	errors.Register(
//	    "USER_EMAIL_TAKEN",
//	    errors.KindConflict,
//	    "email address is already registered",
//	    errors.WithSeverity(errors.SeverityError),
//	    errors.WithRetryable(false),
//	)
func Register(code Code, kind Kind, template string, opts ...RegisterOption) {
	if err := globalRegistry.Register(code, kind, template, opts...); err != nil {
		panic(fmt.Sprintf("failed to register error code %s: %v", code, err))
	}
}

// MustRegister is an alias for Register (for clarity).
func MustRegister(code Code, kind Kind, template string, opts ...RegisterOption) {
	Register(code, kind, template, opts...)
}

// TryRegister registers an error code and returns an error if it fails.
// Does not panic, useful for dynamic registration.
func TryRegister(code Code, kind Kind, template string, opts ...RegisterOption) error {
	return globalRegistry.Register(code, kind, template, opts...)
}

// Register registers an error code in this registry instance.
func (r *Registry) Register(code Code, kind Kind, template string, opts ...RegisterOption) error {
	// Validate code format
	if err := code.Validate(); err != nil {
		return fmt.Errorf("invalid code: %w", err)
	}

	// Check for duplicate registration
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.errors[code]; exists {
		return fmt.Errorf("error code %s is already registered", code)
	}

	// Create registered error with defaults
	registered := &RegisteredError{
		Code:       code,
		Kind:       kind,
		Template:   template,
		Severity:   SeverityError,
		Retryable:  kind.IsRetryableByDefault(),
		HTTPStatus: 0, // 0 means use kind default
		GRPCCode:   0, // 0 means use kind default
		Metadata:   make(map[string]any),
	}

	// Apply options
	for _, opt := range opts {
		opt(registered)
	}

	// Store in registry
	r.errors[code] = registered

	return nil
}

// ============================================================================
// Registration Options
// ============================================================================

// RegisterOption configures a RegisteredError during registration.
type RegisterOption func(*RegisteredError)

// WithSeverity sets the default severity for this error code.
func WithSeverity(severity Severity) RegisterOption {
	return func(r *RegisteredError) {
		r.Severity = severity
	}
}

// WithRetryable sets whether errors of this code are retryable by default.
func WithRetryable(retryable bool) RegisterOption {
	return func(r *RegisteredError) {
		r.Retryable = retryable
	}
}

// WithHTTPStatus sets a custom HTTP status code for this error.
// If not set, the status will be derived from Kind.
func WithHTTPStatus(status int) RegisterOption {
	return func(r *RegisteredError) {
		r.HTTPStatus = status
	}
}

// WithGRPCCode sets a custom gRPC status code for this error.
// If not set, the code will be derived from Kind.
func WithGRPCCode(code int) RegisterOption {
	return func(r *RegisteredError) {
		r.GRPCCode = code
	}
}

// WithMeta adds metadata to the registered error.
// Useful for documentation, deprecation notices, etc.
//
// Example:
//
//	errors.WithMeta("deprecated", true)
//	errors.WithMeta("replacement", "USER_NOT_FOUND_V2")
//	errors.WithMeta("docs", "https://docs.example.com/errors/USER_NOT_FOUND")
func WithMeta(key string, value any) RegisterOption {
	return func(r *RegisteredError) {
		if r.Metadata == nil {
			r.Metadata = make(map[string]any)
		}
		r.Metadata[key] = value
	}
}

// WithMetadata sets multiple metadata entries at once.
func WithMetadata(metadata map[string]any) RegisterOption {
	return func(r *RegisteredError) {
		if r.Metadata == nil {
			r.Metadata = make(map[string]any)
		}
		maps.Copy(r.Metadata, metadata)
	}
}

// ============================================================================
// Lookup
// ============================================================================

// Lookup finds a registered error by code.
// Returns nil if the code is not registered.
func Lookup(code Code) *RegisteredError {
	return globalRegistry.Lookup(code)
}

// Lookup finds a registered error by code in this registry instance.
func (r *Registry) Lookup(code Code) *RegisteredError {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.errors[code]
}

// IsRegistered checks if an error code is registered.
func IsRegistered(code Code) bool {
	return globalRegistry.IsRegistered(code)
}

// IsRegistered checks if a code is registered in this registry instance.
func (r *Registry) IsRegistered(code Code) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.errors[code]
	return exists
}

// MustLookup finds a registered error and panics if not found.
// Only use this when you're certain the code is registered.
func MustLookup(code Code) *RegisteredError {
	registered := Lookup(code)
	if registered == nil {
		panic(fmt.Sprintf("error code %s is not registered", code))
	}
	return registered
}

// ============================================================================
// Error Creation from Registry
// ============================================================================

// ByCode creates an error from a registered code.
// Returns an error if the code is not registered.
//
// Example:
//
//	return errors.ByCode("USER_EMAIL_TAKEN", "users.Service.Create").
//	    WithMeta("email", email)
func ByCode(code Code, operation string) *Error {
	registered := Lookup(code)
	if registered == nil {
		// Code not registered - create a generic error
		return &Error{
			Kind:      KindInternal,
			Code:      code,
			Operation: operation,
			Message:   fmt.Sprintf("unregistered error code: %s", code),
			Severity:  SeverityError,
			Retryable: false,
			Err:       ErrInternal,
		}
	}

	// Create error from registered definition
	return &Error{
		Kind:      registered.Kind,
		Code:      registered.Code,
		Operation: operation,
		Message:   registered.Template,
		Severity:  registered.Severity,
		Retryable: registered.Retryable,
		Err:       sentinelForKind(registered.Kind),
	}
}

// ByCodef creates an error from a registered code with formatted message.
// The format string and args are applied to the registered template.
//
// Example:
//
//	registered template: "user %s not found"
//	errors.ByCodef("USER_NOT_FOUND", "repo.Find", "123")
//	// Message: "user 123 not found"
func ByCodef(code Code, operation string, args ...any) *Error {
	registered := Lookup(code)
	if registered == nil {
		return ByCode(code, operation)
	}

	err := ByCode(code, operation)
	err.Message = fmt.Sprintf(registered.Template, args...)
	return err
}

// MustByCode creates an error from a registered code and panics if not registered.
// Only use this when you're certain the code is registered.
func MustByCode(code Code, operation string) *Error {
	MustLookup(code) // Panic if not registered
	return ByCode(code, operation)
}

// ============================================================================
// Registry Inspection
// ============================================================================

// List returns all registered errors.
// The returned slice is sorted by code for consistent output.
func List() []*RegisteredError {
	return globalRegistry.List()
}

// List returns all registered errors in this registry instance.
func (r *Registry) List() []*RegisteredError {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*RegisteredError, 0, len(r.errors))
	for _, registered := range r.errors {
		result = append(result, registered)
	}

	// Sort by code for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Code < result[j].Code
	})

	return result
}

// ListByKind returns all registered errors of a specific kind.
func ListByKind(kind Kind) []*RegisteredError {
	return globalRegistry.ListByKind(kind)
}

// ListByKind returns errors of a specific kind from this registry instance.
func (r *Registry) ListByKind(kind Kind) []*RegisteredError {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*RegisteredError
	for _, registered := range r.errors {
		if registered.Kind == kind {
			result = append(result, registered)
		}
	}

	// Sort by code
	sort.Slice(result, func(i, j int) bool {
		return result[i].Code < result[j].Code
	})

	return result
}

// ListByDomain returns all registered errors for a specific domain.
// Domain is extracted from the code prefix (e.g., "USER_" from "USER_NOT_FOUND").
func ListByDomain(domain string) []*RegisteredError {
	return globalRegistry.ListByDomain(domain)
}

// ListByDomain returns errors for a domain from this registry instance.
func (r *Registry) ListByDomain(domain string) []*RegisteredError {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*RegisteredError
	for _, registered := range r.errors {
		if registered.Code.Domain() == domain {
			result = append(result, registered)
		}
	}

	// Sort by code
	sort.Slice(result, func(i, j int) bool {
		return result[i].Code < result[j].Code
	})

	return result
}

// Count returns the total number of registered errors.
func Count() int {
	return globalRegistry.Count()
}

// Count returns the number of registered errors in this registry instance.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.errors)
}

// Domains returns all unique domain prefixes from registered codes.
// Sorted alphabetically.
func Domains() []string {
	return globalRegistry.Domains()
}

// Domains returns all unique domains from this registry instance.
func (r *Registry) Domains() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domainSet := make(map[string]struct{})
	for _, registered := range r.errors {
		domain := registered.Code.Domain()
		if domain != "" {
			domainSet[domain] = struct{}{}
		}
	}

	domains := make([]string, 0, len(domainSet))
	for domain := range domainSet {
		domains = append(domains, domain)
	}

	sort.Strings(domains)
	return domains
}

// ============================================================================
// Registry Management
// ============================================================================

// Clear removes all registered errors.
// Useful for testing, should not be used in production.
func Clear() {
	globalRegistry.Clear()
}

// Clear removes all errors from this registry instance.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.errors = make(map[Code]*RegisteredError)
}

// Unregister removes a specific error code from the registry.
// Returns true if the code was found and removed.
// Useful for testing or dynamic error code management.
func Unregister(code Code) bool {
	return globalRegistry.Unregister(code)
}

// Unregister removes a code from this registry instance.
func (r *Registry) Unregister(code Code) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.errors[code]; exists {
		delete(r.errors, code)
		return true
	}

	return false
}

// ============================================================================
// Multiple Registry Instances
// ============================================================================

// NewRegistry creates a new isolated registry instance.
// Useful for testing or when you need multiple independent registries.
//
// Example:
//
//	testRegistry := errors.NewRegistry()
//	testRegistry.Register("TEST_ERROR", errors.KindValidation, "test error")
func NewRegistry() *Registry {
	return &Registry{
		errors: make(map[Code]*RegisteredError),
	}
}
