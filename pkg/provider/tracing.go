package provider

import (
	"context"

	"github.com/0xsj/hexagonal-go/pkg/observability/tracing"
	"github.com/0xsj/hexagonal-go/pkg/observability/tracing/otel"
)

// ProvideTracingProvider creates an OpenTelemetry tracing provider.
func ProvideTracingProvider(ctx context.Context, config tracing.Config) (tracing.Provider, error) {
	if !config.Enabled {
		return &tracing.NoopProvider{}, nil
	}
	return otel.New(ctx, config)
}
