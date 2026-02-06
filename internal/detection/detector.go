package detection

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/0xsj/firewatch/internal/fingerprint"
)

// DetectionResult holds everything the detector found for a request.
type DetectionResult struct {
	Signatures []*Signature // Matched signatures
	Patterns   []*Pattern   // Matched patterns
	Severity   string       // Highest severity across all matches
}

// Matched reports whether any signatures or patterns matched.
func (r *DetectionResult) Matched() bool {
	return len(r.Signatures) > 0 || len(r.Patterns) > 0
}

// SignatureIDs returns the IDs of all matched signatures.
func (r *DetectionResult) SignatureIDs() []string {
	ids := make([]string, len(r.Signatures))
	for i, s := range r.Signatures {
		ids[i] = s.ID
	}
	return ids
}

// Detector evaluates requests against signatures and patterns.
type Detector struct {
	signatures []*Signature
	patterns   []*Pattern
	compiled   map[string]*compiledRegex
	mu         sync.RWMutex
	logger     *slog.Logger
}

// compiledRegex caches compiled regexps for signature/pattern matchers.
type compiledRegex struct{}

// New creates a Detector with the given signatures and patterns.
func New(sigs []*Signature, pats []*Pattern, logger *slog.Logger) *Detector {
	return &Detector{
		signatures: sigs,
		patterns:   pats,
		logger:     logger.With("component", "detector"),
	}
}

// NewDefault creates a Detector loaded with all built-in
// signatures and patterns.
func NewDefault(logger *slog.Logger) *Detector {
	return New(DefaultSignatures(), DefaultPatterns(), logger)
}

// Detect evaluates all signatures and patterns against the request
// and returns the combined result.
func (d *Detector) Detect(r *http.Request, body string) *DetectionResult {
	result := &DetectionResult{}
	fields := d.extractFields(r, body)

	// Evaluate signatures (all matchers must match)
	for _, sig := range d.signatures {
		if d.matchSignature(sig, fields) {
			result.Signatures = append(result.Signatures, sig)
			result.Severity = highestSeverity(result.Severity, sig.Severity)
		}
	}

	// Evaluate patterns (any rule must match)
	for _, pat := range d.patterns {
		if d.matchPattern(pat, fields) {
			result.Patterns = append(result.Patterns, pat)
			result.Severity = highestSeverity(result.Severity, pat.Severity)
		}
	}

	if result.Matched() {
		d.logger.Debug("detection match",
			"signatures", result.SignatureIDs(),
			"severity", result.Severity,
			"ip", fields[FieldPath],
		)
	}

	return result
}

// requestFields maps MatchFields to their values for a request.
type requestFields map[MatchField]string

// extractFields pulls all matchable values from the request.
func (d *Detector) extractFields(r *http.Request, body string) requestFields {
	fields := requestFields{
		FieldPath:      r.URL.Path,
		FieldMethod:    r.Method,
		FieldBody:      body,
		FieldUserAgent: r.UserAgent(),
		FieldQuery:     r.URL.RawQuery,
	}

	// Add all headers as matchable fields
	for key, vals := range r.Header {
		if len(vals) > 0 {
			fields[HeaderField(key)] = vals[0]
		}
	}

	// Add fingerprint data if available
	fp := fingerprint.GetResult(r.Context())
	if fp.KnownClient != "" {
		fields[MatchField("known_client")] = fp.KnownClient
	}

	return fields
}

// matchSignature checks if all matchers in a signature match.
func (d *Detector) matchSignature(sig *Signature, fields requestFields) bool {
	for _, m := range sig.Matchers {
		value := fields[m.Field]

		// For header fields, also try canonical form
		if m.Field.IsHeaderField() && value == "" {
			value = fields[HeaderField(http.CanonicalHeaderKey(m.Field.HeaderName()))]
		}

		if !m.Match(value) {
			return false
		}
	}
	return len(sig.Matchers) > 0
}

// matchPattern checks if any rule in a pattern matches.
func (d *Detector) matchPattern(pat *Pattern, fields requestFields) bool {
	for _, rule := range pat.Rules {
		if d.matchRule(rule, fields) {
			return true
		}
	}
	return false
}

// matchRule checks if all matchers in a rule match.
func (d *Detector) matchRule(rule Rule, fields requestFields) bool {
	for _, m := range rule.Matchers {
		value := fields[m.Field]
		if m.Field.IsHeaderField() && value == "" {
			value = fields[HeaderField(http.CanonicalHeaderKey(m.Field.HeaderName()))]
		}
		if !m.Match(value) {
			return false
		}
	}
	return len(rule.Matchers) > 0
}

// Severity ordering from lowest to highest.
var severityRank = map[string]int{
	"info":     0,
	"low":      1,
	"medium":   2,
	"high":     3,
	"critical": 4,
}

// highestSeverity returns whichever severity is more severe.
func highestSeverity(a, b string) string {
	if severityRank[b] > severityRank[a] {
		return b
	}
	return a
}
