package models

// IOCType identifies the kind of indicator.
type IOCType string

const (
	IOCTypeIP     IOCType = "ip"
	IOCTypeDomain IOCType = "domain"
	IOCTypeURL    IOCType = "url"
	IOCTypeHash   IOCType = "hash"
	IOCTypeEmail  IOCType = "email"
	IOCTypeCIDR   IOCType = "cidr"
)

// IOC is an Indicator of Compromise extracted from honeypot events.
type IOC struct {
	ID        string  `json:"id"`         // UUID v4
	Type      IOCType `json:"type"`       // ip, domain, url, hash, email, cidr
	Value     string  `json:"value"`      // The indicator value
	FirstSeen string  `json:"first_seen"` // RFC3339 UTC
	LastSeen  string  `json:"last_seen"`  // RFC3339 UTC

	// Context
	Source   string `json:"source"`   // How it was found (event ID, detection rule, etc.)
	Severity string `json:"severity"` // critical, high, medium, low, info

	// Enrichment
	GeoIP    *GeoIPInfo `json:"geoip,omitempty"`
	Hostname string     `json:"hostname,omitempty"`

	// Classification
	Tags []string `json:"tags,omitempty"`
}
