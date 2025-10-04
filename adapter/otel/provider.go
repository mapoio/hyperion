package otel

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// otelProvider manages shared OpenTelemetry providers (TracerProvider, MeterProvider, LoggerProvider).
// It ensures all telemetry signals share the same resource configuration.
type otelProvider struct {
	serviceName    string
	resource       *resource.Resource
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	mu             sync.Mutex
}

var (
	globalProvider *otelProvider
	providerOnce   sync.Once
	providerMu     sync.Mutex
)

// resetProviderForTesting resets the global provider. ONLY for testing.
func resetProviderForTesting() {
	providerMu.Lock()
	defer providerMu.Unlock()

	globalProvider = nil
	providerOnce = sync.Once{}
}

// initProvider initializes the shared OpenTelemetry provider with the given service name.
// It creates a resource with service.name attribute that is shared across all telemetry signals.
func initProvider(serviceName string, attrs map[string]string) (*otelProvider, error) {
	// Build resource attributes
	resAttrs := []resource.Option{
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	}

	// Add custom attributes from config
	if len(attrs) > 0 {
		// Note: Custom attributes handling will be implemented when needed
		// Currently focusing on service.name which is the most important attribute
		_ = attrs
	}

	// Create shared resource
	res, err := resource.New(context.Background(), resAttrs...)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	return &otelProvider{
		serviceName: serviceName,
		resource:    res,
	}, nil
}

// getOrCreateProvider returns the global provider, creating it if necessary.
// This ensures singleton behavior for the OTel provider.
func getOrCreateProvider(serviceName string, attrs map[string]string) (*otelProvider, error) {
	var err error
	providerOnce.Do(func() {
		globalProvider, err = initProvider(serviceName, attrs)
	})
	if err != nil {
		return nil, err
	}
	return globalProvider, nil
}

// initTracerProvider initializes the TracerProvider with the given configuration.
func (p *otelProvider) initTracerProvider(cfg TracingConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.tracerProvider != nil {
		return nil // Already initialized
	}

	// Create exporter based on config
	exporter, err := createTraceExporter(cfg)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create TracerProvider with shared resource
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRate)),
		sdktrace.WithResource(p.resource),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	p.tracerProvider = tp
	return nil
}

// initMeterProvider initializes the MeterProvider with the given configuration.
func (p *otelProvider) initMeterProvider(cfg MetricsConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.meterProvider != nil {
		return nil // Already initialized
	}

	// Create metrics reader based on config
	reader, err := createMetricsReader(cfg)
	if err != nil {
		return fmt.Errorf("failed to create metrics reader: %w", err)
	}

	// Create MeterProvider with shared resource
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(p.resource),
	)

	// Set global meter provider
	otel.SetMeterProvider(mp)

	p.meterProvider = mp
	return nil
}

// getTracer returns a Tracer from the TracerProvider.
func (p *otelProvider) getTracer() *sdktrace.TracerProvider {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.tracerProvider
}

// getMeter returns a Meter from the MeterProvider.
func (p *otelProvider) getMeter() *sdkmetric.MeterProvider {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.meterProvider
}

// shutdown shuts down all providers gracefully.
func (p *otelProvider) shutdown(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error

	if p.tracerProvider != nil {
		if err := p.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("tracer provider shutdown failed: %w", err))
		}
	}

	if p.meterProvider != nil {
		if err := p.meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("meter provider shutdown failed: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}
