package detection

import (
	"regexp"
	"strings"
)

// Signature is a detection rule that matches specific scanner behavior.
// All matchers must match for the signature to fire (AND logic).
type Signature struct {
	ID          string    // Unique ID, e.g., "nextjs-action-001"
	Name        string    // Human-readable name
	Description string    // What this signature detects
	Module      string    // Restrict to a module ("" = any module)
	Severity    string    // Severity when matched
	Matchers    []Matcher // All must match
}

// Matcher is a single condition in a signature or pattern rule.
type Matcher struct {
	Field    MatchField // What to match against
	Operator MatchOp    // How to compare
	Value    string     // Expected value
	Negate   bool       // Invert the result
}

// MatchField identifies which part of the request to inspect.
type MatchField string

const (
	FieldPath      MatchField = "path"
	FieldMethod    MatchField = "method"
	FieldBody      MatchField = "body"
	FieldUserAgent MatchField = "user_agent"
	FieldQuery     MatchField = "query"
)

// HeaderField returns a MatchField for a specific header.
// Example: HeaderField("Next-Action") targets that header's value.
func HeaderField(name string) MatchField {
	return MatchField("header:" + name)
}

// IsHeaderField reports whether the field targets a specific header.
func (f MatchField) IsHeaderField() bool {
	return strings.HasPrefix(string(f), "header:")
}

// HeaderName extracts the header name from a header field.
func (f MatchField) HeaderName() string {
	return strings.TrimPrefix(string(f), "header:")
}

// MatchOp defines how a matcher compares the field value.
type MatchOp string

const (
	OpEquals   MatchOp = "equals"
	OpContains MatchOp = "contains"
	OpPrefix   MatchOp = "prefix"
	OpSuffix   MatchOp = "suffix"
	OpRegex    MatchOp = "regex"
	OpExists   MatchOp = "exists" // Field is present (value ignored)
)

// Match evaluates a single matcher against a field value.
func (m Matcher) Match(value string) bool {
	var result bool

	switch m.Operator {
	case OpEquals:
		result = value == m.Value
	case OpContains:
		result = strings.Contains(value, m.Value)
	case OpPrefix:
		result = strings.HasPrefix(value, m.Value)
	case OpSuffix:
		result = strings.HasSuffix(value, m.Value)
	case OpRegex:
		re, err := regexp.Compile(m.Value)
		if err != nil {
			return false
		}
		result = re.MatchString(value)
	case OpExists:
		result = value != ""
	default:
		return false
	}

	if m.Negate {
		return !result
	}
	return result
}

// DefaultSignatures returns the built-in detection signatures.
func DefaultSignatures() []*Signature {
	return []*Signature{
		// --- Next.js ---
		{
			ID:          "nextjs-action-001",
			Name:        "Next.js Server Action Probe",
			Description: "POST request with Next-Action header indicating server action vulnerability scanning",
			Module:      "nextjs",
			Severity:    "high",
			Matchers: []Matcher{
				{Field: FieldMethod, Operator: OpEquals, Value: "POST"},
				{Field: HeaderField("Next-Action"), Operator: OpExists},
			},
		},
		{
			ID:          "nextjs-rsc-001",
			Name:        "Next.js RSC Probe",
			Description: "Request with RSC header probing for React Server Components",
			Module:      "nextjs",
			Severity:    "medium",
			Matchers: []Matcher{
				{Field: HeaderField("Rsc"), Operator: OpExists},
			},
		},
		{
			ID:          "nextjs-debug-001",
			Name:        "Next.js Debug Endpoint",
			Description: "Access to Next.js debug/development endpoint",
			Module:      "nextjs",
			Severity:    "high",
			Matchers: []Matcher{
				{Field: FieldPath, Operator: OpPrefix, Value: "/__nextjs_original-stack-frame"},
			},
		},
		{
			ID:          "nextjs-sourcemap-001",
			Name:        "Next.js Source Map Probe",
			Description: "Request for JavaScript source maps to extract source code",
			Module:      "nextjs",
			Severity:    "medium",
			Matchers: []Matcher{
				{Field: FieldPath, Operator: OpPrefix, Value: "/_next/"},
				{Field: FieldPath, Operator: OpSuffix, Value: ".map"},
			},
		},

		// --- WordPress ---
		{
			ID:          "wp-login-001",
			Name:        "WordPress Login Probe",
			Description: "Access to WordPress login page",
			Module:      "wordpress",
			Severity:    "medium",
			Matchers: []Matcher{
				{Field: FieldPath, Operator: OpEquals, Value: "/wp-login.php"},
			},
		},
		{
			ID:          "wp-bruteforce-001",
			Name:        "WordPress Brute Force",
			Description: "POST to WordPress login with credentials",
			Module:      "wordpress",
			Severity:    "high",
			Matchers: []Matcher{
				{Field: FieldMethod, Operator: OpEquals, Value: "POST"},
				{Field: FieldPath, Operator: OpEquals, Value: "/wp-login.php"},
			},
		},
		{
			ID:          "wp-xmlrpc-001",
			Name:        "WordPress XML-RPC Probe",
			Description: "Access to WordPress XML-RPC endpoint, often used for brute force or DDoS",
			Module:      "wordpress",
			Severity:    "high",
			Matchers: []Matcher{
				{Field: FieldPath, Operator: OpEquals, Value: "/xmlrpc.php"},
			},
		},

		// --- Exposure ---
		{
			ID:          "exposure-env-001",
			Name:        "Environment File Probe",
			Description: "Access to .env file containing secrets",
			Module:      "exposure",
			Severity:    "high",
			Matchers: []Matcher{
				{Field: FieldPath, Operator: OpRegex, Value: `^/\.env`},
			},
		},
		{
			ID:          "exposure-git-001",
			Name:        "Git Repository Probe",
			Description: "Access to .git directory to extract source code",
			Module:      "exposure",
			Severity:    "high",
			Matchers: []Matcher{
				{Field: FieldPath, Operator: OpPrefix, Value: "/.git"},
			},
		},
		{
			ID:          "exposure-config-001",
			Name:        "Config File Probe",
			Description: "Access to application configuration files",
			Module:      "exposure",
			Severity:    "medium",
			Matchers: []Matcher{
				{Field: FieldPath, Operator: OpRegex, Value: `(?i)\.(cfg|conf|config|ini|yml|yaml|toml|xml|properties)$`},
			},
		},

		// --- Generic ---
		{
			ID:          "generic-scanner-001",
			Name:        "Known Scanner User-Agent",
			Description: "Request from a known scanning tool",
			Severity:    "medium",
			Matchers: []Matcher{
				{Field: FieldUserAgent, Operator: OpRegex, Value: `(?i)(nuclei|sqlmap|nikto|nmap|masscan|zgrab|dirsearch|gobuster|ffuf|wfuzz)`},
			},
		},
		{
			ID:          "generic-traversal-001",
			Name:        "Path Traversal Attempt",
			Description: "Request containing directory traversal sequences",
			Severity:    "high",
			Matchers: []Matcher{
				{Field: FieldPath, Operator: OpContains, Value: ".."},
			},
		},
		{
			ID:          "generic-admin-001",
			Name:        "Admin Panel Probe",
			Description: "Access to common admin panel paths",
			Severity:    "medium",
			Matchers: []Matcher{
				{Field: FieldPath, Operator: OpRegex, Value: `(?i)^/(admin|administrator|phpmyadmin|adminer|manager|cpanel)`},
			},
		},
	}
}
