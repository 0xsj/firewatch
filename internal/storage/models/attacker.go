package models

// Attacker is an aggregated profile built from observed events.
type Attacker struct {
	ID        string `json:"id"`         // UUID v4
	FirstSeen string `json:"first_seen"` // RFC3339 UTC
	LastSeen  string `json:"last_seen"`  // RFC3339 UTC

	// Identity
	IP         string     `json:"ip"`
	Hostname   string     `json:"hostname,omitempty"`
	GeoIP      *GeoIPInfo `json:"geoip,omitempty"`
	UserAgents []string   `json:"user_agents,omitempty"`

	// Behavior
	TotalEvents     int      `json:"total_events"`
	ModulesTargeted []string `json:"modules_targeted"`
	PathsProbed     []string `json:"paths_probed"`
	Severity        string   `json:"severity"` // Highest severity observed

	// Fingerprint
	JA3Hashes []string `json:"ja3_hashes,omitempty"`

	// Classification
	Tags []string `json:"tags,omitempty"` // e.g. "scanner", "brute-force", "researcher"
}
