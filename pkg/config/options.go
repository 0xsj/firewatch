package config

import (
	"context"
	"time"
)

// ============================================================================
// Source Options
// ============================================================================

// SourceOption configures a configuration source.
type SourceOption func(any)

// FileSourceOptions configures a FileSource.
type FileSourceOptions struct {
	// Format explicitly sets the file format (yaml, json, toml)
	// If empty, format is detected from file extension
	Format string

	// Required causes an error if the file doesn't exist
	// Default: false (missing files are skipped)
	Required bool
}

// WithFormat sets the file format explicitly.
func WithFormat(format string) SourceOption {
	return func(s any) {
		if fs, ok := s.(*FileSourceOptions); ok {
			fs.Format = format
		}
	}
}

// WithRequired marks a source as required.
func WithRequired() SourceOption {
	return func(s any) {
		if fs, ok := s.(*FileSourceOptions); ok {
			fs.Required = true
		}
	}
}

// EnvSourceOptions configures an EnvSource.
type EnvSourceOptions struct {
	// Prefix is prepended to all environment variable names
	Prefix string

	// CaseSensitive determines if env var matching is case-sensitive
	// Default: false (case-insensitive)
	CaseSensitive bool

	// StripPrefix removes the prefix from field matching
	// Useful when prefix is just for namespacing
	StripPrefix bool
}

// WithPrefix sets the environment variable prefix.
func WithPrefix(prefix string) SourceOption {
	return func(s any) {
		if es, ok := s.(*EnvSourceOptions); ok {
			es.Prefix = prefix
		}
	}
}

// WithCaseSensitive enables case-sensitive env var matching.
func WithCaseSensitive() SourceOption {
	return func(s any) {
		if es, ok := s.(*EnvSourceOptions); ok {
			es.CaseSensitive = true
		}
	}
}

// ============================================================================
// Watcher Options
// ============================================================================

// WatcherOptions configures configuration watching/hot-reload.
type WatcherOptions struct {
	// Interval is how often to check for changes
	// Default: 5s
	Interval time.Duration

	// Debounce prevents rapid reloads
	// Multiple changes within this window trigger a single reload
	// Default: 100ms
	Debounce time.Duration

	// OnChange is called when configuration changes
	// Receives the new configuration
	OnChange func(ctx context.Context, config any)

	// OnError is called when reload fails
	OnError func(ctx context.Context, err error)
}

// WatcherOption configures a configuration watcher.
type WatcherOption func(*WatcherOptions)

// WithInterval sets the watch interval.
func WithInterval(interval time.Duration) WatcherOption {
	return func(o *WatcherOptions) {
		o.Interval = interval
	}
}

// WithDebounce sets the debounce duration.
func WithDebounce(debounce time.Duration) WatcherOption {
	return func(o *WatcherOptions) {
		o.Debounce = debounce
	}
}

// WithOnChange sets the change callback.
func WithOnChange(fn func(ctx context.Context, config any)) WatcherOption {
	return func(o *WatcherOptions) {
		o.OnChange = fn
	}
}

// WithOnError sets the error callback.
func WithOnError(fn func(ctx context.Context, err error)) WatcherOption {
	return func(o *WatcherOptions) {
		o.OnError = fn
	}
}

// ============================================================================
// Default Options
// ============================================================================

// DefaultFileSourceOptions returns default file source options.
func DefaultFileSourceOptions() *FileSourceOptions {
	return &FileSourceOptions{
		Format:   "",
		Required: false,
	}
}

// DefaultEnvSourceOptions returns default env source options.
func DefaultEnvSourceOptions() *EnvSourceOptions {
	return &EnvSourceOptions{
		Prefix:        "",
		CaseSensitive: false,
		StripPrefix:   false,
	}
}

// DefaultWatcherOptions returns default watcher options.
func DefaultWatcherOptions() *WatcherOptions {
	return &WatcherOptions{
		Interval: 5 * time.Second,
		Debounce: 100 * time.Millisecond,
		OnChange: func(ctx context.Context, config any) {},
		OnError:  func(ctx context.Context, err error) {},
	}
}

// ============================================================================
// Builder Options
// ============================================================================

// BuilderOption configures a configuration builder.
type BuilderOption func(*Builder)

// Builder provides a fluent interface for loading configuration.
type Builder struct {
	sources   []Source
	validator Validator
	options   *LoadOptions
}

// NewBuilder creates a new configuration builder.
func NewBuilder(opts ...BuilderOption) *Builder {
	b := &Builder{
		sources: make([]Source, 0),
		options: &LoadOptions{
			MergeStrategy:       SourcesOverride,
			FailOnMissingSource: false,
		},
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// WithFileSource adds a file source to the builder.
func (b *Builder) WithFileSource(path string, opts ...SourceOption) *Builder {
	fileOpts := DefaultFileSourceOptions()
	for _, opt := range opts {
		opt(fileOpts)
	}

	b.sources = append(b.sources, NewFileSourceWithOptions(path, fileOpts))
	return b
}

// WithEnvSource adds an environment source to the builder.
func (b *Builder) WithEnvSource(opts ...SourceOption) *Builder {
	envOpts := DefaultEnvSourceOptions()
	for _, opt := range opts {
		opt(envOpts)
	}

	b.sources = append(b.sources, NewEnvSourceWithOptions(envOpts))
	return b
}

// WithStaticSource adds a static source to the builder.
func (b *Builder) WithStaticSource(name string, config any) *Builder {
	b.sources = append(b.sources, NewStaticSource(name, config))
	return b
}

// WithSource adds a custom source to the builder.
func (b *Builder) WithSource(source Source) *Builder {
	b.sources = append(b.sources, source)
	return b
}

// WithValidation adds validation to the builder.
func (b *Builder) WithValidation(validator Validator) *Builder {
	b.validator = validator
	b.options.Validator = validator
	return b
}

// WithValidationFunc adds a validation function to the builder.
func (b *Builder) WithValidationFunc(fn func(context.Context, any) error) *Builder {
	validator := NewFuncValidator(fn)
	return b.WithValidation(validator)
}

// WithFailOnMissing causes the loader to fail if any source is missing.
func (b *Builder) WithFailOnMissing() *Builder {
	b.options.FailOnMissingSource = true
	return b
}

// Build creates a Loader from the builder configuration.
func (b *Builder) Build() *Loader {
	return NewLoader(b.sources,
		WithValidator(b.validator),
		func(o *LoadOptions) {
			o.FailOnMissingSource = b.options.FailOnMissingSource
			o.MergeStrategy = b.options.MergeStrategy
		},
	)
}

// Load builds the loader and loads configuration.
func (b *Builder) Load(ctx context.Context, target any) error {
	loader := b.Build()
	return loader.Load(ctx, target)
}

// MustLoad builds the loader and loads configuration, panicking on error.
func (b *Builder) MustLoad(ctx context.Context, target any) {
	if err := b.Load(ctx, target); err != nil {
		panic(err)
	}
}
