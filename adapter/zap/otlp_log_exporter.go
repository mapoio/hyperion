package zap

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// createOtlpLogBridge creates an OTLP log bridge with the provided configuration.
func createOtlpLogBridge(config *OtlpLogConfig) (*otlpLogBridge, error) {
	if config == nil {
		return nil, fmt.Errorf("OTLP log config is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create OTLP gRPC exporter
	exporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(config.Endpoint),
		otlploggrpc.WithInsecure(), // Use insecure for local HyperDX
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	// Create resource with service name
	serviceName := config.ServiceName
	if serviceName == "" {
		serviceName = "hyperion-service"
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create logger provider
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)

	// Create and return the bridge
	return newOtlpLogBridge(provider, serviceName), nil
}
