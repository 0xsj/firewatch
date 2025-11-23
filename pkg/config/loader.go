package config

import (
	"context"
	"fmt"
)

// Loader loads configuration from sources into a target struct.
// It is stateless and can be used multiple times.
type Loader struct {
	sources []Source
	options *LoadOptions
}

// Source represents a configuration source (file, env, etc.)
type Source interface {
	// Load loads configuration into the target struct
	Load(ctx context.Context, target any) error

	// Name returns the source name for debugging
	Name() string
}

// LoadOptions configures the loading behavior.
type LoadOptions struct {
	// Validator validates the loaded config
	Validator Validator

	// FailOnMissingSource causes loading to fail if a source doesn't exist
	// Default: false (missing sources are skipped)
	FailOnMissingSource bool

	// MergeStrategy determines how multiple sources are merged
	// Default: SourcesOverride (later sources override earlier)
	MergeStrategy MergeStrategy
}

// MergeStrategy determines how configuration from multiple sources is combined.
type MergeStrategy string

const (
	// SourcesOverride means later sources override earlier sources
	SourcesOverride MergeStrategy = "override"

	// SourcesMergeDeep means sources are deep-merged (nested structs combined)
	SourcesMergeDeep MergeStrategy = "deep_merge"
)

// Validator validates configuration.
type Validator interface {
	Validate(ctx context.Context, v any) error
}

// NewLoader creates a new configuration loader.
func NewLoader(sources []Source, opts ...LoadOption) *Loader {
	options := &LoadOptions{
		MergeStrategy:       SourcesOverride,
		FailOnMissingSource: false,
	}

	for _, opt := range opts {
		opt(options)
	}

	return &Loader{
		sources: sources,
		options: options,
	}
}

// Load loads configuration from all sources into target.
// target must be a pointer to a struct.
func (l *Loader) Load(ctx context.Context, target any) error {
	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	// Validate target is a pointer to struct
	if err := validateTarget(target); err != nil {
		return err
	}

	// Load from each source in order
	for _, source := range l.sources {
		if err := source.Load(ctx, target); err != nil {
			// Check if error is because source doesn't exist
			if isMissingSourceError(err) && !l.options.FailOnMissingSource {
				// Skip missing sources silently
				continue
			}
			return fmt.Errorf("failed to load from source %s: %w", source.Name(), err)
		}
	}

	// Validate if validator is configured
	if l.options.Validator != nil {
		if err := l.options.Validator.Validate(ctx, target); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	return nil
}

// LoadOption configures a Loader.
type LoadOption func(*LoadOptions)

// WithValidator sets a validator for the loaded configuration.
func WithValidator(validator Validator) LoadOption {
	return func(o *LoadOptions) {
		o.Validator = validator
	}
}

// WithFailOnMissingSource causes loading to fail if a source doesn't exist.
func WithFailOnMissingSource() LoadOption {
	return func(o *LoadOptions) {
		o.FailOnMissingSource = true
	}
}

// WithMergeStrategy sets the merge strategy for multiple sources.
func WithMergeStrategy(strategy MergeStrategy) LoadOption {
	return func(o *LoadOptions) {
		o.MergeStrategy = strategy
	}
}
