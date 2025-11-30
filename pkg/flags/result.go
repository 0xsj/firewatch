package flags

// Result represents a flag evaluation result.
type Result struct {
	Key         string
	Enabled     bool
	Variant     string
	Reason      string
	MatchedRule string
}

// Reason constants for evaluation results.
const (
	ReasonDisabled = "disabled"
	ReasonDefault  = "default"
	ReasonRule     = "rule"
	ReasonOverride = "override"
)
