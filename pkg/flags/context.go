package flags

import (
	"context"
	"crypto/sha256"
	"encoding/binary"

	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
)

// EvaluationContext holds the context for flag evaluation.
type EvaluationContext struct {
	tenantID   string
	userID     string
	attributes map[string]string
}

// NewEvaluationContext creates a new evaluation context.
func NewEvaluationContext(tenantID, userID string) *EvaluationContext {
	return &EvaluationContext{
		tenantID:   tenantID,
		userID:     userID,
		attributes: make(map[string]string),
	}
}

// FromContext extracts tenant and user from request context.
func FromContext(ctx context.Context) *EvaluationContext {
	tenantID := middleware.GetTenantID(ctx)
	userID := middleware.GetUserID(ctx)
	return NewEvaluationContext(tenantID, userID)
}

// TenantID returns the tenant ID.
func (c *EvaluationContext) TenantID() string {
	return c.tenantID
}

// UserID returns the user ID.
func (c *EvaluationContext) UserID() string {
	return c.userID
}

// WithTenantID sets the tenant ID.
func (c *EvaluationContext) WithTenantID(tenantID string) *EvaluationContext {
	c.tenantID = tenantID
	return c
}

// WithUserID sets the user ID.
func (c *EvaluationContext) WithUserID(userID string) *EvaluationContext {
	c.userID = userID
	return c
}

// WithAttribute sets an attribute.
func (c *EvaluationContext) WithAttribute(key, value string) *EvaluationContext {
	c.attributes[key] = value
	return c
}

// WithAttributes sets multiple attributes.
func (c *EvaluationContext) WithAttributes(attrs map[string]string) *EvaluationContext {
	for k, v := range attrs {
		c.attributes[k] = v
	}
	return c
}

// GetAttribute returns an attribute value.
func (c *EvaluationContext) GetAttribute(key string) (string, bool) {
	v, ok := c.attributes[key]
	return v, ok
}

// Attributes returns all attributes.
func (c *EvaluationContext) Attributes() map[string]string {
	return c.attributes
}

// HashBucket returns a deterministic bucket (0-99) for percentage-based rules.
func (c *EvaluationContext) HashBucket(flagKey string) int {
	identifier := c.userID
	if identifier == "" {
		identifier = c.tenantID
	}
	if identifier == "" {
		return 0
	}

	h := sha256.New()
	h.Write([]byte(flagKey + ":" + identifier))
	sum := h.Sum(nil)

	num := binary.BigEndian.Uint32(sum[:4])
	return int(num % 100)
}
