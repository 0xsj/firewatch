package provider

import (
	"github.com/0xsj/hexagonal-go/pkg/observability/metrics"
	"github.com/0xsj/hexagonal-go/pkg/observability/metrics/prometheus"
)

// ProvideMetricsProvider creates a Prometheus metrics provider.
func ProvideMetricsProvider(config metrics.Config) metrics.Provider {
	if !config.Enabled {
		return &metrics.NoopProvider{}
	}
	return prometheus.New(config)
}
