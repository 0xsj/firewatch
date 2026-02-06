package detection

import "testing"

func TestMatcherOperators(t *testing.T) {
	tests := []struct {
		name    string
		matcher Matcher
		value   string
		want    bool
	}{
		// Equals
		{"equals match", Matcher{Operator: OpEquals, Value: "POST"}, "POST", true},
		{"equals no match", Matcher{Operator: OpEquals, Value: "POST"}, "GET", false},

		// Contains
		{"contains match", Matcher{Operator: OpContains, Value: ".."}, "/etc/../passwd", true},
		{"contains no match", Matcher{Operator: OpContains, Value: ".."}, "/normal/path", false},

		// Prefix
		{"prefix match", Matcher{Operator: OpPrefix, Value: "/_next/"}, "/_next/static/chunk.js", true},
		{"prefix no match", Matcher{Operator: OpPrefix, Value: "/_next/"}, "/api/users", false},

		// Suffix
		{"suffix match", Matcher{Operator: OpSuffix, Value: ".map"}, "/_next/static/main.js.map", true},
		{"suffix no match", Matcher{Operator: OpSuffix, Value: ".map"}, "/style.css", false},

		// Regex
		{"regex match", Matcher{Operator: OpRegex, Value: `^/\.env`}, "/.env", true},
		{"regex match variant", Matcher{Operator: OpRegex, Value: `^/\.env`}, "/.env.local", true},
		{"regex no match", Matcher{Operator: OpRegex, Value: `^/\.env`}, "/environment", false},
		{"regex invalid pattern", Matcher{Operator: OpRegex, Value: `[invalid`}, "anything", false},

		// Exists
		{"exists match", Matcher{Operator: OpExists}, "any-value", true},
		{"exists empty", Matcher{Operator: OpExists}, "", false},

		// Negate
		{"negate equals", Matcher{Operator: OpEquals, Value: "GET", Negate: true}, "POST", true},
		{"negate equals same", Matcher{Operator: OpEquals, Value: "GET", Negate: true}, "GET", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.matcher.Match(tt.value)
			if got != tt.want {
				t.Errorf("Match(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestDefaultSignatures(t *testing.T) {
	sigs := DefaultSignatures()
	if len(sigs) == 0 {
		t.Fatal("DefaultSignatures() returned empty slice")
	}

	seen := make(map[string]bool)
	for _, sig := range sigs {
		if sig.ID == "" {
			t.Error("signature has empty ID")
		}
		if seen[sig.ID] {
			t.Errorf("duplicate signature ID: %s", sig.ID)
		}
		seen[sig.ID] = true

		if len(sig.Matchers) == 0 {
			t.Errorf("signature %s has no matchers", sig.ID)
		}
		if sig.Severity == "" {
			t.Errorf("signature %s has no severity", sig.ID)
		}
	}
}

func TestDefaultPatterns(t *testing.T) {
	pats := DefaultPatterns()
	if len(pats) == 0 {
		t.Fatal("DefaultPatterns() returned empty slice")
	}

	seen := make(map[string]bool)
	for _, pat := range pats {
		if pat.ID == "" {
			t.Error("pattern has empty ID")
		}
		if seen[pat.ID] {
			t.Errorf("duplicate pattern ID: %s", pat.ID)
		}
		seen[pat.ID] = true

		if len(pat.Rules) == 0 {
			t.Errorf("pattern %s has no rules", pat.ID)
		}
	}
}
