package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

// createTraceExporter creates a trace exporter based on the configuration.
func createTraceExporter(cfg TracingConfig) (trace.SpanExporter, error) {
	switch cfg.Exporter {
	case exporterOTLP:
		return otlptracegrpc.New(context.Background(),
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
			otlptracegrpc.WithInsecure(), // TODO: Add TLS configuration
		)
	case exporterJaeger:
		// Jaeger exporter is deprecated in newer OTel versions
		// We'll use OTLP instead and users can configure Jaeger to accept OTLP
		return otlptracegrpc.New(context.Background(),
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
			otlptracegrpc.WithInsecure(),
		)
	default:
		return nil, fmt.Errorf("unsupported trace exporter: %s", cfg.Exporter)
	}
}

// createMetricsReader creates a metrics reader based on the configuration.
func createMetricsReader(cfg MetricsConfig) (metric.Reader, error) {
	switch cfg.Exporter {
	case exporterPrometheus:
		return prometheus.New()
	case exporterOTLP:
		// Create OTLP gRPC metrics exporter
		exporter, err := otlpmetricgrpc.New(context.Background(),
			otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
			otlpmetricgrpc.WithInsecure(), // TODO: Add TLS configuration
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP metrics exporter: %w", err)
		}

		// Create periodic reader with the specified interval
		return metric.NewPeriodicReader(exporter,
			metric.WithInterval(cfg.Interval),
		), nil
	default:
		return nil, fmt.Errorf("unsupported metrics exporter: %s", cfg.Exporter)
	}
}
