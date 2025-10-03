package otel

import (
	"context"
	"fmt"

	"github.com/mapoio/hyperion"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.uber.org/fx"
)

// NewOtelTracer creates an OpenTelemetry Tracer from configuration.
func NewOtelTracer(config hyperion.Config) (hyperion.Tracer, error) {
	cfg, err := LoadTracingConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load tracing config: %w", err)
	}

	if !cfg.Enabled {
		return hyperion.NewNoOpTracer(), nil
	}

	// Create exporter based on config
	exporter, err := createTraceExporter(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create resource with service name
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRate)),
		sdktrace.WithResource(res),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	tracer := tp.Tracer(cfg.ServiceName)
	return &otelTracer{tracer: tracer, provider: tp}, nil
}

// NewOtelMeter creates an OpenTelemetry Meter from configuration.
func NewOtelMeter(config hyperion.Config) (hyperion.Meter, error) {
	cfg, err := LoadMetricsConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load metrics config: %w", err)
	}

	if !cfg.Enabled {
		return hyperion.NewNoOpMeter(), nil
	}

	// Create metrics reader based on config
	reader, err := createMetricsReader(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics reader: %w", err)
	}

	// Create resource with service name
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create MeterProvider
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)

	// Set global meter provider
	otel.SetMeterProvider(mp)

	meter := mp.Meter(cfg.ServiceName)
	return &otelMeter{meter: meter}, nil
}

// RegisterTracerShutdownHook registers a shutdown hook for the tracer.
func RegisterTracerShutdownHook(lc fx.Lifecycle, tracer hyperion.Tracer) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if otelTracer, ok := tracer.(*otelTracer); ok {
				return otelTracer.Shutdown(ctx)
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
	fx.Invoke(RegisterTracerShutdownHook),
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

// Module provides both Tracer and Meter.
var Module = fx.Options(TracerModule, MeterModule)
