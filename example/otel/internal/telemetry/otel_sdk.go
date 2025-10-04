package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/fx"
)

// SDKConfig holds configuration for the OpenTelemetry SDK.
type SDKConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	EnablePrometheus bool
}

// NewSDKConfig creates SDK configuration from environment or defaults.
func NewSDKConfig() *SDKConfig {
	return &SDKConfig{
		ServiceName:    "hyperion-otel-example",
		ServiceVersion: "v1.0.0",
		Environment:    "development",
		OTLPEndpoint:   "localhost:4317",
		EnablePrometheus: true,
	}
}

// NewResource creates an OpenTelemetry Resource with service metadata.
// This resource is shared across all telemetry signals (traces, metrics, logs).
func NewResource(cfg *SDKConfig) (*resource.Resource, error) {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.ServiceVersion),
		semconv.DeploymentEnvironment(cfg.Environment),
	), nil
}

// NewTracerProvider creates and configures an OpenTelemetry TracerProvider.
// Applications have full control over the configuration.
func NewTracerProvider(cfg *SDKConfig, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// Create OTLP trace exporter
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create TracerProvider with custom configuration
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Customize sampling strategy
	)

	// Set as global provider
	otel.SetTracerProvider(tp)

	return tp, nil
}

// NewMeterProvider creates and configures an OpenTelemetry MeterProvider.
// Supports both OTLP and Prometheus exporters.
func NewMeterProvider(cfg *SDKConfig, res *resource.Resource) (*metric.MeterProvider, error) {
	ctx := context.Background()

	var readers []metric.Reader

	// OTLP Metrics Exporter
	otlpExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}
	readers = append(readers, metric.NewPeriodicReader(otlpExporter,
		metric.WithInterval(10*time.Second),
	))

	// Prometheus Exporter (optional)
	if cfg.EnablePrometheus {
		promExporter, err := prometheus.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
		}
		readers = append(readers, promExporter)
	}

	// Create MeterProvider with all readers
	opts := []metric.Option{
		metric.WithResource(res),
	}
	for _, reader := range readers {
		opts = append(opts, metric.WithReader(reader))
	}
	mp := metric.NewMeterProvider(opts...)

	// Set as global provider
	otel.SetMeterProvider(mp)

	return mp, nil
}

// RegisterShutdown registers graceful shutdown hooks for OpenTelemetry SDK.
// This ensures all telemetry data is flushed before application exit.
func RegisterShutdown(
	lc fx.Lifecycle,
	tp *sdktrace.TracerProvider,
	mp *metric.MeterProvider,
) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			// Shutdown with timeout
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			// Shutdown TracerProvider
			if err := tp.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("failed to shutdown TracerProvider: %w", err)
			}

			// Shutdown MeterProvider
			if err := mp.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("failed to shutdown MeterProvider: %w", err)
			}

			return nil
		},
	})
}

// Module provides a fully configured OpenTelemetry SDK.
// This is the single source of truth for OTel configuration in the application.
var Module = fx.Module("telemetry",
	fx.Provide(
		NewSDKConfig,
		NewResource,
		NewTracerProvider,
		NewMeterProvider,
	),
	fx.Invoke(RegisterShutdown),
)
