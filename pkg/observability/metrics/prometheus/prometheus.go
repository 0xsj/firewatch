package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/observability/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Provider is a Prometheus implementation of metrics.Provider.
type Provider struct {
	config   metrics.Config
	registry *prometheus.Registry

	// Cached metrics
	counters   map[string]*counterVec
	gauges     map[string]*gaugeVec
	histograms map[string]*histogramVec
	mu         sync.RWMutex

	server *http.Server
}

// New creates a new Prometheus metrics provider.
func New(config metrics.Config) *Provider {
	registry := prometheus.NewRegistry()

	// Register default collectors if enabled
	if config.CollectGoMetrics {
		registry.MustRegister(collectors.NewGoCollector())
	}
	if config.CollectProcessMetrics {
		registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}

	return &Provider{
		config:     config,
		registry:   registry,
		counters:   make(map[string]*counterVec),
		gauges:     make(map[string]*gaugeVec),
		histograms: make(map[string]*histogramVec),
	}
}

// Counter creates or retrieves a counter metric.
func (p *Provider) Counter(name, help string, labels ...string) metrics.Counter {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := p.metricKey(name)
	if c, ok := p.counters[key]; ok {
		return c
	}

	opts := prometheus.CounterOpts{
		Namespace: p.config.Namespace,
		Subsystem: p.config.Subsystem,
		Name:      name,
		Help:      help,
	}

	vec := prometheus.NewCounterVec(opts, labels)
	p.registry.MustRegister(vec)

	c := &counterVec{vec: vec, labels: labels}
	p.counters[key] = c
	return c
}

// Gauge creates or retrieves a gauge metric.
func (p *Provider) Gauge(name, help string, labels ...string) metrics.Gauge {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := p.metricKey(name)
	if g, ok := p.gauges[key]; ok {
		return g
	}

	opts := prometheus.GaugeOpts{
		Namespace: p.config.Namespace,
		Subsystem: p.config.Subsystem,
		Name:      name,
		Help:      help,
	}

	vec := prometheus.NewGaugeVec(opts, labels)
	p.registry.MustRegister(vec)

	g := &gaugeVec{vec: vec, labels: labels}
	p.gauges[key] = g
	return g
}

// Histogram creates or retrieves a histogram metric.
func (p *Provider) Histogram(name, help string, buckets []float64, labels ...string) metrics.Histogram {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := p.metricKey(name)
	if h, ok := p.histograms[key]; ok {
		return h
	}

	if buckets == nil {
		buckets = p.config.HistogramBuckets
	}

	opts := prometheus.HistogramOpts{
		Namespace: p.config.Namespace,
		Subsystem: p.config.Subsystem,
		Name:      name,
		Help:      help,
		Buckets:   buckets,
	}

	vec := prometheus.NewHistogramVec(opts, labels)
	p.registry.MustRegister(vec)

	h := &histogramVec{vec: vec, labels: labels}
	p.histograms[key] = h
	return h
}

// Handler returns an HTTP handler for exposing metrics.
func (p *Provider) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// Start starts a metrics server on the configured port.
func (p *Provider) Start(ctx context.Context) error {
	if !p.config.Enabled {
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle(p.config.Path, p.Handler())

	p.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", p.config.Port),
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Check for immediate startup errors
	select {
	case err := <-errCh:
		return fmt.Errorf("metrics server failed to start: %w", err)
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

// Close gracefully shuts down the metrics provider.
func (p *Provider) Close() error {
	if p.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return p.server.Shutdown(ctx)
	}
	return nil
}

// metricKey generates a unique key for a metric.
func (p *Provider) metricKey(name string) string {
	return fmt.Sprintf("%s_%s_%s", p.config.Namespace, p.config.Subsystem, name)
}

// ============================================================================
// Counter Implementation
// ============================================================================

type counterVec struct {
	vec    *prometheus.CounterVec
	labels []string
}

func (c *counterVec) Inc(labels ...string) {
	c.vec.WithLabelValues(labels...).Inc()
}

func (c *counterVec) Add(value float64, labels ...string) {
	c.vec.WithLabelValues(labels...).Add(value)
}

// ============================================================================
// Gauge Implementation
// ============================================================================

type gaugeVec struct {
	vec    *prometheus.GaugeVec
	labels []string
}

func (g *gaugeVec) Set(value float64, labels ...string) {
	g.vec.WithLabelValues(labels...).Set(value)
}

func (g *gaugeVec) Inc(labels ...string) {
	g.vec.WithLabelValues(labels...).Inc()
}

func (g *gaugeVec) Dec(labels ...string) {
	g.vec.WithLabelValues(labels...).Dec()
}

func (g *gaugeVec) Add(value float64, labels ...string) {
	g.vec.WithLabelValues(labels...).Add(value)
}

func (g *gaugeVec) Sub(value float64, labels ...string) {
	g.vec.WithLabelValues(labels...).Sub(value)
}

// ============================================================================
// Histogram Implementation
// ============================================================================

type histogramVec struct {
	vec    *prometheus.HistogramVec
	labels []string
}

func (h *histogramVec) Observe(value float64, labels ...string) {
	h.vec.WithLabelValues(labels...).Observe(value)
}
