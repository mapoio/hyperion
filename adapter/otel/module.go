package otel

import (
	"context"
	"fmt"

	"github.com/mapoio/hyperion"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
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
	return &OtelTracer{tracer: tracer, provider: tp}, nil
}

// NewOtelTracerFromProvider creates a hyperion.Tracer from an external TracerProvider.
// This allows applications to fully control OTel SDK initialization and version,
// while still using Hyperion's tracing abstractions.
//
// Example usage:
//
//	fx.Provide(func(tp trace.TracerProvider) hyperion.Tracer {
//	    return otel.NewOtelTracerFromProvider(tp, "my-service")
//	})
func NewOtelTracerFromProvider(provider trace.TracerProvider, serviceName string) hyperion.Tracer {
	tracer := provider.Tracer(serviceName)
	return &OtelTracer{
		tracer:   tracer,
		provider: provider,
	}
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
	return &OtelMeter{meter: meter, provider: mp}, nil
}

// NewOtelMeterFromProvider creates a hyperion.Meter from an external MeterProvider.
// This allows applications to fully control OTel SDK initialization and version,
// while still using Hyperion's metrics abstractions.
//
// Example usage:
//
//	fx.Provide(func(mp metric.MeterProvider) hyperion.Meter {
//	    return otel.NewOtelMeterFromProvider(mp, "my-service")
//	})
func NewOtelMeterFromProvider(provider metric.MeterProvider, serviceName string) hyperion.Meter {
	meter := provider.Meter(serviceName)
	return &OtelMeter{
		meter:    meter,
		provider: provider,
	}
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
