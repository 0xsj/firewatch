package models

// HoneyToken tracks a unique fake credential issued to a specific
// request. When the same value appears in a later request, the
// honeypot can correlate attacker activity across sessions.
type HoneyToken struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`       // e.g. "aws_access_key", "aws_secret_key", "session_token", "imds_token", "api_key", "db_password"
	Value     string `json:"value"`      // the fake credential value
	IssuedAt  string `json:"issued_at"`  // RFC3339
	SourceIP  string `json:"source_ip"`  // IP that received this token
	Module    string `json:"module"`     // honeypot module that issued it
	Path      string `json:"path"`       // request path
	RequestID string `json:"request_id"` // correlates with event
}
