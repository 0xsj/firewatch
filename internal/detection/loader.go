package detection

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// SignatureFile is the top-level YAML structure for custom signatures.
type SignatureFile struct {
	Signatures []SignatureYAML `yaml:"signatures"`
	Patterns   []PatternYAML   `yaml:"patterns"`
}

// SignatureYAML is the YAML representation of a Signature.
type SignatureYAML struct {
	ID          string        `yaml:"id"`
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Module      string        `yaml:"module"`
	Severity    string        `yaml:"severity"`
	Matchers    []MatcherYAML `yaml:"matchers"`
}

// MatcherYAML is the YAML representation of a Matcher.
type MatcherYAML struct {
	Field    string `yaml:"field"`
	Operator string `yaml:"operator"`
	Value    string `yaml:"value"`
	Negate   bool   `yaml:"negate"`
}

// PatternYAML is the YAML representation of a Pattern.
type PatternYAML struct {
	ID          string     `yaml:"id"`
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Category    string     `yaml:"category"`
	Severity    string     `yaml:"severity"`
	Rules       []RuleYAML `yaml:"rules"`
}

// RuleYAML is the YAML representation of a Rule.
type RuleYAML struct {
	Matchers []MatcherYAML `yaml:"matchers"`
}

var validOperators = map[string]MatchOp{
	"equals":   OpEquals,
	"contains": OpContains,
	"prefix":   OpPrefix,
	"suffix":   OpSuffix,
	"regex":    OpRegex,
	"exists":   OpExists,
}

// LoadSignatures loads signatures and patterns from a single YAML file.
func LoadSignatures(path string) ([]*Signature, []*Pattern, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("reading signatures file %s: %w", path, err)
	}

	var sf SignatureFile
	if err := yaml.Unmarshal(data, &sf); err != nil {
		return nil, nil, fmt.Errorf("parsing signatures file %s: %w", path, err)
	}

	sigs, err := convertSignatures(sf.Signatures)
	if err != nil {
		return nil, nil, fmt.Errorf("in file %s: %w", path, err)
	}

	pats, err := convertPatterns(sf.Patterns)
	if err != nil {
		return nil, nil, fmt.Errorf("in file %s: %w", path, err)
	}

	return sigs, pats, nil
}

// LoadSignaturesDir loads all YAML files from a directory.
func LoadSignaturesDir(dir string) ([]*Signature, []*Pattern, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("reading signatures directory %s: %w", dir, err)
	}

	var allSigs []*Signature
	var allPats []*Pattern

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		sigs, pats, err := LoadSignatures(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, nil, err
		}
		allSigs = append(allSigs, sigs...)
		allPats = append(allPats, pats...)
	}

	return allSigs, allPats, nil
}

// NewWithCustom creates a Detector that merges built-in signatures/patterns
// with custom ones. Custom signatures with the same ID override built-in ones.
func NewWithCustom(customSigs []*Signature, customPats []*Pattern, logger *slog.Logger) *Detector {
	sigs := mergeSigs(DefaultSignatures(), customSigs)
	pats := mergePats(DefaultPatterns(), customPats)
	return New(sigs, pats, logger)
}

func convertSignatures(yamlSigs []SignatureYAML) ([]*Signature, error) {
	var sigs []*Signature
	for _, ys := range yamlSigs {
		if ys.ID == "" {
			return nil, fmt.Errorf("signature missing required field: id")
		}
		if ys.Name == "" {
			return nil, fmt.Errorf("signature %q missing required field: name", ys.ID)
		}
		if ys.Severity == "" {
			return nil, fmt.Errorf("signature %q missing required field: severity", ys.ID)
		}
		if len(ys.Matchers) == 0 {
			return nil, fmt.Errorf("signature %q must have at least one matcher", ys.ID)
		}

		matchers, err := convertMatchers(ys.Matchers)
		if err != nil {
			return nil, fmt.Errorf("signature %q: %w", ys.ID, err)
		}

		sigs = append(sigs, &Signature{
			ID:          ys.ID,
			Name:        ys.Name,
			Description: ys.Description,
			Module:      ys.Module,
			Severity:    ys.Severity,
			Matchers:    matchers,
		})
	}
	return sigs, nil
}

func convertPatterns(yamlPats []PatternYAML) ([]*Pattern, error) {
	var pats []*Pattern
	for _, yp := range yamlPats {
		if yp.ID == "" {
			return nil, fmt.Errorf("pattern missing required field: id")
		}
		if yp.Name == "" {
			return nil, fmt.Errorf("pattern %q missing required field: name", yp.ID)
		}
		if yp.Severity == "" {
			return nil, fmt.Errorf("pattern %q missing required field: severity", yp.ID)
		}
		if len(yp.Rules) == 0 {
			return nil, fmt.Errorf("pattern %q must have at least one rule", yp.ID)
		}

		var rules []Rule
		for i, yr := range yp.Rules {
			matchers, err := convertMatchers(yr.Matchers)
			if err != nil {
				return nil, fmt.Errorf("pattern %q rule %d: %w", yp.ID, i, err)
			}
			rules = append(rules, Rule{Matchers: matchers})
		}

		pats = append(pats, &Pattern{
			ID:          yp.ID,
			Name:        yp.Name,
			Description: yp.Description,
			Category:    Category(yp.Category),
			Severity:    yp.Severity,
			Rules:       rules,
		})
	}
	return pats, nil
}

func convertMatchers(yamlMatchers []MatcherYAML) ([]Matcher, error) {
	var matchers []Matcher
	for _, ym := range yamlMatchers {
		op, ok := validOperators[ym.Operator]
		if !ok {
			return nil, fmt.Errorf("unknown operator %q", ym.Operator)
		}

		// Validate regex at load time
		if op == OpRegex {
			if _, err := regexp.Compile(ym.Value); err != nil {
				return nil, fmt.Errorf("invalid regex %q: %w", ym.Value, err)
			}
		}

		field := MatchField(ym.Field)
		if hdr, ok := strings.CutPrefix(ym.Field, "header:"); ok {
			field = HeaderField(hdr)
		}

		matchers = append(matchers, Matcher{
			Field:    field,
			Operator: op,
			Value:    ym.Value,
			Negate:   ym.Negate,
		})
	}
	return matchers, nil
}

func mergeSigs(builtIn, custom []*Signature) []*Signature {
	byID := make(map[string]*Signature)
	for _, s := range builtIn {
		byID[s.ID] = s
	}
	for _, s := range custom {
		byID[s.ID] = s // custom overrides built-in
	}

	merged := make([]*Signature, 0, len(byID))

	// Preserve order: built-in first, then new custom
	seen := make(map[string]bool)
	for _, s := range builtIn {
		merged = append(merged, byID[s.ID])
		seen[s.ID] = true
	}
	for _, s := range custom {
		if !seen[s.ID] {
			merged = append(merged, s)
			seen[s.ID] = true
		}
	}
	return merged
}

func mergePats(builtIn, custom []*Pattern) []*Pattern {
	byID := make(map[string]*Pattern)
	for _, p := range builtIn {
		byID[p.ID] = p
	}
	for _, p := range custom {
		byID[p.ID] = p
	}

	seen := make(map[string]bool)
	var merged []*Pattern
	for _, p := range builtIn {
		merged = append(merged, byID[p.ID])
		seen[p.ID] = true
	}
	for _, p := range custom {
		if !seen[p.ID] {
			merged = append(merged, p)
			seen[p.ID] = true
		}
	}
	return merged
}
