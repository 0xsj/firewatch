package models

// Event represents a single captured honeypot interaction.
type Event struct {
	ID        string `json:"id"`         // UUID v4
	Timestamp string `json:"timestamp"`  // RFC3339 UTC
	RequestID string `json:"request_id"` // Correlation ID

	// Source
	SourceIP   string `json:"source_ip"` // Normalized IPv4/IPv6
	SourcePort int    `json:"source_port"`

	// Request
	Module    string            `json:"module"`  // Honeypot module name
	Method    string            `json:"method"`  // HTTP method
	Path      string            `json:"path"`    // Request path
	Query     string            `json:"query"`   // Raw query string
	Headers   map[string]string `json:"headers"` // Flattened headers
	Body      string            `json:"body"`    // Request body (truncated)
	UserAgent string            `json:"user_agent"`

	// Analysis
	Severity   string   `json:"severity"`   // critical, high, medium, low, info
	Signatures []string `json:"signatures"` // Matched signature IDs

	// Fingerprint
	Fingerprint Fingerprint `json:"fingerprint"`

	// Enrichment
	GeoIP      *GeoIPInfo `json:"geoip,omitempty"`
	ReverseDNS string     `json:"reverse_dns,omitempty"`

	// Tracking
	AttackerID string `json:"attacker_id,omitempty"` // Linked attacker profile
	CampaignID string `json:"campaign_id,omitempty"` // Linked campaign
}

// Fingerprint holds request fingerprinting data.
type Fingerprint struct {
	JA3         string   `json:"ja3,omitempty"`
	JA3Hash     string   `json:"ja3_hash,omitempty"`
	JA4         string   `json:"ja4,omitempty"`
	HeaderOrder []string `json:"header_order,omitempty"`
}

// GeoIPInfo holds geolocation data for an IP address.
type GeoIPInfo struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	City        string  `json:"city"`
	ASN         int     `json:"asn"`
	ASNOrg      string  `json:"asn_org"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}
