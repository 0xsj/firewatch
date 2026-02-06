package detection

// Pattern describes a broader attack behavior class. Unlike
// signatures which match individual requests, patterns may
// describe behaviors that span multiple requests or match
// on request context (fingerprint anomalies, known clients).
type Pattern struct {
	ID          string   // Unique ID, e.g., "recon-sweep-001"
	Name        string   // Human-readable name
	Description string   // What this pattern detects
	Category    Category // Attack category
	Severity    string   // Severity when matched
	Rules       []Rule   // At least one rule must match (OR logic)
}

// Rule is a set of matchers that must all match (AND logic).
// A pattern fires when any of its rules matches.
type Rule struct {
	Matchers []Matcher
}

// Category classifies the type of attack a pattern detects.
type Category string

const (
	CategoryRecon       Category = "reconnaissance"
	CategoryEnumeration Category = "enumeration"
	CategoryBruteForce  Category = "brute_force"
	CategoryExploit     Category = "exploit"
	CategoryExfil       Category = "exfiltration"
)

// DefaultPatterns returns the built-in attack patterns.
func DefaultPatterns() []*Pattern {
	return []*Pattern{
		{
			ID:          "recon-sweep-001",
			Name:        "Sensitive File Sweep",
			Description: "Request probing for common sensitive files and directories",
			Category:    CategoryRecon,
			Severity:    "medium",
			Rules: []Rule{
				{Matchers: []Matcher{
					{Field: FieldPath, Operator: OpRegex, Value: `(?i)^/\.(env|git|svn|htaccess|htpasswd|DS_Store)`},
				}},
				{Matchers: []Matcher{
					{Field: FieldPath, Operator: OpRegex, Value: `(?i)^/(backup|dump|export|database|db)\b`},
				}},
				{Matchers: []Matcher{
					{Field: FieldPath, Operator: OpRegex, Value: `(?i)\.(sql|bak|old|orig|save|swp|swo|tmp)$`},
				}},
			},
		},
		{
			ID:          "recon-api-001",
			Name:        "API Discovery",
			Description: "Request probing for API documentation and debug endpoints",
			Category:    CategoryRecon,
			Severity:    "medium",
			Rules: []Rule{
				{Matchers: []Matcher{
					{Field: FieldPath, Operator: OpRegex, Value: `(?i)^/(swagger|api-docs|openapi|graphql|graphiql)`},
				}},
				{Matchers: []Matcher{
					{Field: FieldPath, Operator: OpRegex, Value: `(?i)^/(debug|trace|metrics|health|status|info|actuator)`},
				}},
			},
		},
		{
			ID:          "exploit-ssrf-001",
			Name:        "SSRF / Cloud Metadata",
			Description: "Request targeting cloud metadata endpoints for SSRF exploitation",
			Category:    CategoryExploit,
			Severity:    "critical",
			Rules: []Rule{
				{Matchers: []Matcher{
					{Field: FieldPath, Operator: OpPrefix, Value: "/latest/meta-data"},
				}},
				{Matchers: []Matcher{
					{Field: FieldPath, Operator: OpPrefix, Value: "/metadata/v1"},
				}},
				{Matchers: []Matcher{
					{Field: HeaderField("Metadata-Flavor"), Operator: OpExists},
				}},
			},
		},
		{
			ID:          "exploit-log4j-001",
			Name:        "Log4Shell Probe",
			Description: "Request containing JNDI lookup strings indicating Log4Shell exploitation",
			Category:    CategoryExploit,
			Severity:    "critical",
			Rules: []Rule{
				{Matchers: []Matcher{
					{Field: FieldPath, Operator: OpContains, Value: "${jndi:"},
				}},
				{Matchers: []Matcher{
					{Field: FieldBody, Operator: OpContains, Value: "${jndi:"},
				}},
				{Matchers: []Matcher{
					{Field: HeaderField("User-Agent"), Operator: OpContains, Value: "${jndi:"},
				}},
				{Matchers: []Matcher{
					{Field: HeaderField("X-Forwarded-For"), Operator: OpContains, Value: "${jndi:"},
				}},
			},
		},
		{
			ID:          "enum-tech-001",
			Name:        "Technology Stack Enumeration",
			Description: "Request probing for specific framework files to identify the tech stack",
			Category:    CategoryEnumeration,
			Severity:    "low",
			Rules: []Rule{
				{Matchers: []Matcher{
					{Field: FieldPath, Operator: OpRegex, Value: `(?i)^/(wp-content|wp-includes|wp-json)`},
				}},
				{Matchers: []Matcher{
					{Field: FieldPath, Operator: OpRegex, Value: `(?i)^/(_next|__next|_nuxt)`},
				}},
				{Matchers: []Matcher{
					{Field: FieldPath, Operator: OpRegex, Value: `(?i)^/(rails|django|laravel|symfony|spring)`},
				}},
			},
		},
	}
}
