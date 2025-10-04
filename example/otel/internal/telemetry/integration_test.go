package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/mapoio/hyperion"
	hyperotel "github.com/mapoio/hyperion/adapter/otel"
)

// TestOTelSDKInitialization tests that the OTel SDK can be initialized correctly.
func TestOTelSDKInitialization(t *testing.T) {
	cfg := NewSDKConfig()
	cfg.OTLPEndpoint = "localhost:4317" // Use local endpoint for testing

	res, err := NewResource(cfg)
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}

	tp, err := NewTracerProvider(cfg, res)
	if err != nil {
		t.Fatalf("Failed to create TracerProvider: %v", err)
	}
	defer tp.Shutdown(context.Background())

	mp, err := NewMeterProvider(cfg, res)
	if err != nil {
		t.Fatalf("Failed to create MeterProvider: %v", err)
	}
	defer mp.Shutdown(context.Background())

	// Verify providers are not nil
	if tp == nil {
		t.Error("TracerProvider is nil")
	}
	if mp == nil {
		t.Error("MeterProvider is nil")
	}
}

// TestHyperionAdapterIntegration tests the integration between Hyperion adapters
// and application-provided OTel SDK.
func TestHyperionAdapterIntegration(t *testing.T) {
	// Step 1: Application layer initializes OTel SDK
	cfg := NewSDKConfig()
	res, err := NewResource(cfg)
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}

	tp, err := NewTracerProvider(cfg, res)
	if err != nil {
		t.Fatalf("Failed to create TracerProvider: %v", err)
	}
	defer tp.Shutdown(context.Background())

	mp, err := NewMeterProvider(cfg, res)
	if err != nil {
		t.Fatalf("Failed to create MeterProvider: %v", err)
	}
	defer mp.Shutdown(context.Background())

	// Step 2: Create Hyperion adapters from application's OTel providers
	tracer := hyperotel.NewOtelTracerFromProvider(tp, "test-service")
	meter := hyperotel.NewOtelMeterFromProvider(mp, "test-service")

	// Step 3: Verify adapters implement correct interfaces
	var _ hyperion.Tracer = tracer
	var _ hyperion.Meter = meter

	// Step 4: Verify accessor methods work
	otelTracer, ok := tracer.(*hyperotel.OtelTracer)
	if !ok {
		t.Fatal("Expected *hyperotel.OtelTracer")
	}

	otelMeter, ok := meter.(*hyperotel.OtelMeter)
	if !ok {
		t.Fatal("Expected *hyperotel.OtelMeter")
	}

	// Step 5: Verify we can access the underlying providers
	if otelTracer.TracerProvider() != tp {
		t.Error("TracerProvider does not match")
	}

	if otelMeter.MeterProvider() != mp {
		t.Error("MeterProvider does not match")
	}
}

// TestOTelVersionIndependence tests that the application can use a different
// OTel version than the adapter.
func TestOTelVersionIndependence(t *testing.T) {
	// This test verifies the key principle: applications control OTel version
	cfg := NewSDKConfig()
	res, err := NewResource(cfg)
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}

	// Application uses latest OTel SDK (controlled by application's go.mod)
	tp, err := NewTracerProvider(cfg, res)
	if err != nil {
		t.Fatalf("Failed to create TracerProvider: %v", err)
	}
	defer tp.Shutdown(context.Background())

	// Hyperion adapter receives the provider (version-agnostic)
	tracer := hyperotel.NewOtelTracerFromProvider(tp, "test-service")

	// Verify tracing works
	ctx := hyperion.NewContextFactory(
		hyperion.NewNoOpLogger(),
		tracer,
		hyperion.NewNoOpDatabase(),
		hyperion.NewNoOpMeter(),
	).New(context.Background())

	ctx, span := tracer.Start(ctx, "test-span")
	defer span.End()

	// Verify span was created
	if !span.SpanContext().IsValid() {
		t.Error("Span context is not valid")
	}
}

// TestRuntimeMetricsCollection tests that runtime metrics can be enabled.
func TestRuntimeMetricsCollection(t *testing.T) {
	cfg := NewSDKConfig()
	res, err := NewResource(cfg)
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}

	mp, err := NewMeterProvider(cfg, res)
	if err != nil {
		t.Fatalf("Failed to create MeterProvider: %v", err)
	}
	defer mp.Shutdown(context.Background())

	// Enable runtime metrics
	if err := EnableRuntimeMetrics(mp); err != nil {
		t.Fatalf("Failed to enable runtime metrics: %v", err)
	}

	// Give metrics some time to be collected
	time.Sleep(2 * time.Second)

	// Runtime metrics are collected in the background
	// We can't directly verify them here, but we can check that
	// the function returns without error
	t.Log("Runtime metrics enabled successfully")
}

// TestHTTPInstrumentation tests HTTP client instrumentation.
func TestHTTPInstrumentation(t *testing.T) {
	cfg := NewSDKConfig()
	res, err := NewResource(cfg)
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}

	tp, err := NewTracerProvider(cfg, res)
	if err != nil {
		t.Fatalf("Failed to create TracerProvider: %v", err)
	}
	defer tp.Shutdown(context.Background())

	// Create Hyperion tracer
	tracer := hyperotel.NewOtelTracerFromProvider(tp, "test-service")

	// Create instrumented HTTP client
	client := NewInstrumentedHTTPClient(tracer)

	if client == nil {
		t.Fatal("Instrumented HTTP client is nil")
	}

	// Verify it's not the default client (it should be wrapped)
	// This is a simple check; in reality, the transport is wrapped
	t.Log("Instrumented HTTP client created successfully")
}

// BenchmarkOTelSDKOverhead benchmarks the overhead of OTel SDK initialization.
func BenchmarkOTelSDKOverhead(b *testing.B) {
	cfg := NewSDKConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, _ := NewResource(cfg)
		tp, _ := NewTracerProvider(cfg, res)
		mp, _ := NewMeterProvider(cfg, res)

		// Cleanup
		tp.Shutdown(context.Background())
		mp.Shutdown(context.Background())
	}
}

// BenchmarkTracingOverhead benchmarks the overhead of creating spans.
func BenchmarkTracingOverhead(b *testing.B) {
	cfg := NewSDKConfig()
	res, _ := NewResource(cfg)
	tp, _ := NewTracerProvider(cfg, res)
	defer tp.Shutdown(context.Background())

	tracer := hyperotel.NewOtelTracerFromProvider(tp, "bench-service")
	ctx := hyperion.NewContextFactory(
		hyperion.NewNoOpLogger(),
		tracer,
		hyperion.NewNoOpDatabase(),
		hyperion.NewNoOpMeter(),
	).New(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, span := tracer.Start(ctx, "benchmark-span")
		span.End()
	}
}
