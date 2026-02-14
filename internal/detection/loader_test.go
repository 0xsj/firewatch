package detection

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSignatures_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sigs.yaml")
	content := `
signatures:
  - id: "custom-test-001"
    name: "Test Signature"
    severity: "high"
    matchers:
      - field: "path"
        operator: "prefix"
        value: "/test"
patterns:
  - id: "custom-pat-001"
    name: "Test Pattern"
    category: "reconnaissance"
    severity: "medium"
    rules:
      - matchers:
          - field: "path"
            operator: "contains"
            value: "admin"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	sigs, pats, err := LoadSignatures(path)
	if err != nil {
		t.Fatalf("LoadSignatures: %v", err)
	}
	if len(sigs) != 1 {
		t.Fatalf("sigs = %d, want 1", len(sigs))
	}
	if sigs[0].ID != "custom-test-001" {
		t.Errorf("sig ID = %q, want custom-test-001", sigs[0].ID)
	}
	if len(pats) != 1 {
		t.Fatalf("pats = %d, want 1", len(pats))
	}
	if pats[0].ID != "custom-pat-001" {
		t.Errorf("pat ID = %q, want custom-pat-001", pats[0].ID)
	}
}

func TestLoadSignatures_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, err := LoadSignatures(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadSignatures_MissingFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.yaml")
	content := `
signatures:
  - name: "No ID"
    severity: "high"
    matchers:
      - field: "path"
        operator: "prefix"
        value: "/test"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, err := LoadSignatures(path)
	if err == nil {
		t.Error("expected error for missing ID")
	}
}

func TestLoadSignatures_BadRegex(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "badregex.yaml")
	content := `
signatures:
  - id: "bad-regex-001"
    name: "Bad Regex"
    severity: "high"
    matchers:
      - field: "path"
        operator: "regex"
        value: "[invalid("
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, err := LoadSignatures(path)
	if err == nil {
		t.Error("expected error for bad regex")
	}
}

func TestLoadSignaturesDir(t *testing.T) {
	dir := t.TempDir()

	file1 := `
signatures:
  - id: "dir-sig-001"
    name: "Dir Sig 1"
    severity: "low"
    matchers:
      - field: "path"
        operator: "equals"
        value: "/a"
`
	file2 := `
patterns:
  - id: "dir-pat-001"
    name: "Dir Pat 1"
    category: "exploit"
    severity: "critical"
    rules:
      - matchers:
          - field: "body"
            operator: "contains"
            value: "evil"
`
	if err := os.WriteFile(filepath.Join(dir, "a.yaml"), []byte(file1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.yml"), []byte(file2), 0644); err != nil {
		t.Fatal(err)
	}
	// Non-YAML file should be skipped
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("skip me"), 0644); err != nil {
		t.Fatal(err)
	}

	sigs, pats, err := LoadSignaturesDir(dir)
	if err != nil {
		t.Fatalf("LoadSignaturesDir: %v", err)
	}
	if len(sigs) != 1 {
		t.Errorf("sigs = %d, want 1", len(sigs))
	}
	if len(pats) != 1 {
		t.Errorf("pats = %d, want 1", len(pats))
	}
}

func TestNewWithCustom_MergeOverride(t *testing.T) {
	custom := []*Signature{
		{
			ID:       "generic-scanner-001", // Override built-in
			Name:     "Custom Scanner Override",
			Severity: "critical",
			Matchers: []Matcher{
				{Field: FieldUserAgent, Operator: OpContains, Value: "evil-bot"},
			},
		},
		{
			ID:       "brand-new-001", // New custom
			Name:     "Brand New",
			Severity: "low",
			Matchers: []Matcher{
				{Field: FieldPath, Operator: OpEquals, Value: "/custom"},
			},
		},
	}

	logger := testLogger()
	det := NewWithCustom(custom, nil, logger)

	// Find the overridden signature
	var found *Signature
	for _, s := range det.signatures {
		if s.ID == "generic-scanner-001" {
			found = s
			break
		}
	}
	if found == nil {
		t.Fatal("generic-scanner-001 not found")
	}
	if found.Name != "Custom Scanner Override" {
		t.Errorf("name = %q, want Custom Scanner Override", found.Name)
	}

	// Find the new custom signature
	var foundNew bool
	for _, s := range det.signatures {
		if s.ID == "brand-new-001" {
			foundNew = true
			break
		}
	}
	if !foundNew {
		t.Error("brand-new-001 not found in merged signatures")
	}
}

func TestLoadSignatures_UnknownOperator(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "unknown_op.yaml")
	content := `
signatures:
  - id: "bad-op-001"
    name: "Bad Op"
    severity: "high"
    matchers:
      - field: "path"
        operator: "fuzzy_match"
        value: "/test"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, err := LoadSignatures(path)
	if err == nil {
		t.Error("expected error for unknown operator")
	}
}

func TestLoadSignatures_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	sigs, pats, err := LoadSignatures(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sigs) != 0 {
		t.Errorf("sigs = %d, want 0", len(sigs))
	}
	if len(pats) != 0 {
		t.Errorf("pats = %d, want 0", len(pats))
	}
}
