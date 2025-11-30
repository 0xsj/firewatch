package flags

import "time"

// ClientOption configures the Client.
type ClientOption func(*DefaultClient)

// WithCache enables in-memory caching with the specified TTL.
func WithCache(ttl time.Duration) ClientOption {
	return func(c *DefaultClient) {
		c.cacheEnabled = true
		c.cacheTTL = ttl
		c.cache = make(map[string]*cachedFlag)
	}
}

// WithDefaultTenant sets a default tenant ID when none is in context.
func WithDefaultTenant(tenantID string) ClientOption {
	return func(c *DefaultClient) {
		c.defaultTenantID = tenantID
	}
}
