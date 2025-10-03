package otel

import (
	"context"
	"fmt"

	"github.com/mapoio/hyperion"

	"go.uber.org/fx"
)

// NewOtelTracer creates an OpenTelemetry Tracer from configuration.
// It uses a shared provider to ensure consistent resource attributes across telemetry signals.
func NewOtelTracer(config hyperion.Config) (hyperion.Tracer, error) {
	cfg, err := LoadTracingConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load tracing config: %w", err)
	}

	if !cfg.Enabled {
		return hyperion.NewNoOpTracer(), nil
	}

	// Get or create shared provider
	provider, err := getOrCreateProvider(cfg.ServiceName, cfg.Attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider: %w", err)
	}

	// Initialize TracerProvider with config
	if err := provider.initTracerProvider(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize tracer provider: %w", err)
	}

	// Get tracer from provider
	tp := provider.getTracer()
	tracer := tp.Tracer(cfg.ServiceName)
	return &otelTracer{tracer: tracer, provider: tp}, nil
}

// NewOtelMeter creates an OpenTelemetry Meter from configuration.
// It uses a shared provider to ensure consistent resource attributes across telemetry signals.
func NewOtelMeter(config hyperion.Config) (hyperion.Meter, error) {
	cfg, err := LoadMetricsConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load metrics config: %w", err)
	}

	if !cfg.Enabled {
		return hyperion.NewNoOpMeter(), nil
	}

	// Get or create shared provider
	provider, err := getOrCreateProvider(cfg.ServiceName, cfg.Attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider: %w", err)
	}

	// Initialize MeterProvider with config
	if err := provider.initMeterProvider(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize meter provider: %w", err)
	}

	// Get meter from provider
	mp := provider.getMeter()
	meter := mp.Meter(cfg.ServiceName)
	return &otelMeter{meter: meter}, nil
}

// RegisterShutdownHook registers a shutdown hook for the global OTel provider.
// This ensures graceful shutdown of both TracerProvider and MeterProvider.
func RegisterShutdownHook(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if globalProvider != nil {
				return globalProvider.shutdown(ctx)
			}
			return nil
		},
	})
}

// TracerModule provides OpenTelemetry Tracer implementation.
var TracerModule = fx.Module("hyperion.adapter.otel.tracer",
	fx.Provide(
		fx.Annotate(
			NewOtelTracer,
			fx.As(new(hyperion.Tracer)),
		),
	),
)

// MeterModule provides OpenTelemetry Meter implementation.
var MeterModule = fx.Module("hyperion.adapter.otel.meter",
	fx.Provide(
		fx.Annotate(
			NewOtelMeter,
			fx.As(new(hyperion.Meter)),
		),
	),
)

// Module provides both Tracer and Meter with unified shutdown.
var Module = fx.Options(
	TracerModule,
	MeterModule,
	fx.Invoke(RegisterShutdownHook),
)
