package models

// Campaign is a correlated group of attacks from multiple sources
// sharing a common pattern or objective.
type Campaign struct {
	ID        string `json:"id"`         // UUID v4
	Name      string `json:"name"`       // Human-readable campaign name
	FirstSeen string `json:"first_seen"` // RFC3339 UTC
	LastSeen  string `json:"last_seen"`  // RFC3339 UTC

	// Scope
	AttackerIPs     []string `json:"attacker_ips"`
	AttackerCount   int      `json:"attacker_count"`
	EventCount      int      `json:"event_count"`
	ModulesTargeted []string `json:"modules_targeted"`

	// Analysis
	Pattern  string `json:"pattern"`  // Description of the shared pattern
	Severity string `json:"severity"` // Overall campaign severity

	// Classification
	Tags []string `json:"tags,omitempty"`
}
