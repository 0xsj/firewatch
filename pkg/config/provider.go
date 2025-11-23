package config

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Provider provides access to configuration.
// This is the main interface for dependency injection.
type Provider[T any] interface {
	// Get returns the current configuration
	Get(ctx context.Context) (*T, error)

	// Watch starts watching for configuration changes
	// Returns a channel that receives new configurations
	Watch(ctx context.Context) (<-chan *T, error)

	// Close stops watching and releases resources
	Close() error
}

// ============================================================================
// Static Provider (no reloading)
// ============================================================================

// StaticProvider provides immutable configuration.
// This is the simplest provider - configuration is loaded once and never changes.
type StaticProvider[T any] struct {
	config *T
}

// NewStaticProvider creates a provider with static configuration.
func NewStaticProvider[T any](config *T) *StaticProvider[T] {
	return &StaticProvider[T]{
		config: config,
	}
}

// Get returns the static configuration.
func (p *StaticProvider[T]) Get(ctx context.Context) (*T, error) {
	return p.config, nil
}

// Watch is not supported for static providers.
func (p *StaticProvider[T]) Watch(ctx context.Context) (<-chan *T, error) {
	return nil, fmt.Errorf("static provider does not support watching")
}

// Close is a no-op for static providers.
func (p *StaticProvider[T]) Close() error {
	return nil
}

// ============================================================================
// Dynamic Provider (with hot-reload)
// ============================================================================

// DynamicProvider provides configuration with hot-reload support.
// Thread-safe for concurrent access.
type DynamicProvider[T any] struct {
	mu      sync.RWMutex
	current *T
	loader  *Loader

	// For watching
	watchCtx    context.Context
	watchCancel context.CancelFunc
	watchCh     chan *T
	watchOnce   sync.Once

	options *WatcherOptions
}

// NewDynamicProvider creates a provider with hot-reload support.
func NewDynamicProvider[T any](loader *Loader, opts ...WatcherOption) (*DynamicProvider[T], error) {
	options := DefaultWatcherOptions()
	for _, opt := range opts {
		opt(options)
	}

	provider := &DynamicProvider[T]{
		loader:  loader,
		options: options,
	}

	// Initial load
	if err := provider.reload(context.Background()); err != nil {
		return nil, fmt.Errorf("initial load failed: %w", err)
	}

	return provider, nil
}

// Get returns the current configuration (thread-safe).
func (p *DynamicProvider[T]) Get(ctx context.Context) (*T, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.current == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}

	return p.current, nil
}

// Watch starts watching for configuration changes.
// Can only be called once. Returns a channel that receives new configurations.
func (p *DynamicProvider[T]) Watch(ctx context.Context) (<-chan *T, error) {
	var err error

	p.watchOnce.Do(func() {
		p.watchCtx, p.watchCancel = context.WithCancel(context.Background())
		p.watchCh = make(chan *T, 1)

		go p.watch(p.watchCtx)
	})

	if p.watchCh == nil {
		err = fmt.Errorf("watch already started")
	}

	return p.watchCh, err
}

// watch runs the watch loop.
func (p *DynamicProvider[T]) watch(ctx context.Context) {
	ticker := time.NewTicker(p.options.Interval)
	defer ticker.Stop()

	var lastModified time.Time
	debounceTimer := time.NewTimer(0)
	if !debounceTimer.Stop() {
		<-debounceTimer.C
	}

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			// Check if sources have changed
			// This is a simplified version - in production, you'd use file watchers
			modified := p.checkSourcesModified()

			if modified.After(lastModified) {
				lastModified = modified

				// Debounce rapid changes
				debounceTimer.Reset(p.options.Debounce)

				go func() {
					<-debounceTimer.C

					if err := p.reload(ctx); err != nil {
						p.options.OnError(ctx, err)
					} else {
						p.mu.RLock()
						config := p.current
						p.mu.RUnlock()

						p.options.OnChange(ctx, config)

						select {
						case p.watchCh <- config:
						default:
							// Channel full, skip
						}
					}
				}()
			}
		}
	}
}

// reload reloads configuration from all sources.
func (p *DynamicProvider[T]) reload(ctx context.Context) error {
	var newConfig T

	if err := p.loader.Load(ctx, &newConfig); err != nil {
		return err
	}

	p.mu.Lock()
	p.current = &newConfig
	p.mu.Unlock()

	return nil
}

// checkSourcesModified checks if any source has been modified.
// This is a placeholder - real implementation would track file modification times.
func (p *DynamicProvider[T]) checkSourcesModified() time.Time {
	// TODO: Implement actual file modification tracking
	return time.Now()
}

// Close stops watching and releases resources.
func (p *DynamicProvider[T]) Close() error {
	if p.watchCancel != nil {
		p.watchCancel()
	}

	if p.watchCh != nil {
		close(p.watchCh)
	}

	return nil
}

// ============================================================================
// Provider Factory
// ============================================================================

// ProviderFactory creates configuration providers.
type ProviderFactory[T any] struct {
	loader  *Loader
	options []WatcherOption
}

// NewProviderFactory creates a new provider factory.
func NewProviderFactory[T any](loader *Loader, opts ...WatcherOption) *ProviderFactory[T] {
	return &ProviderFactory[T]{
		loader:  loader,
		options: opts,
	}
}

// Static creates a static provider (load once, no reload).
func (f *ProviderFactory[T]) Static(ctx context.Context) (Provider[T], error) {
	var config T
	if err := f.loader.Load(ctx, &config); err != nil {
		return nil, err
	}

	return NewStaticProvider(&config), nil
}

// Dynamic creates a dynamic provider (with hot-reload).
func (f *ProviderFactory[T]) Dynamic() (Provider[T], error) {
	return NewDynamicProvider[T](f.loader, f.options...)
}

// MustStatic creates a static provider, panicking on error.
func (f *ProviderFactory[T]) MustStatic(ctx context.Context) Provider[T] {
	provider, err := f.Static(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to create static provider: %v", err))
	}
	return provider
}

// MustDynamic creates a dynamic provider, panicking on error.
func (f *ProviderFactory[T]) MustDynamic() Provider[T] {
	provider, err := f.Dynamic()
	if err != nil {
		panic(fmt.Sprintf("failed to create dynamic provider: %v", err))
	}
	return provider
}
