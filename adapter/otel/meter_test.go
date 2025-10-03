package otel

import (
	"context"
	"testing"

	"github.com/mapoio/hyperion"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestOtelMeter_Counter(t *testing.T) {
	// Create a manual reader for testing
	reader := metric.NewManualReader()

	// Create meter provider
	mp := metric.NewMeterProvider(
		metric.WithReader(reader),
	)

	meter := &otelMeter{
		meter: mp.Meter("test"),
	}

	ctx := context.Background()

	t.Run("creates and uses counter", func(t *testing.T) {
		counter := meter.Counter("test-counter")

		if counter == nil {
			t.Fatal("expected non-nil counter")
		}

		// Add some values
		counter.Add(ctx, 5, hyperion.String("method", "test"))
		counter.Add(ctx, 3, hyperion.String("method", "test"))

		// Collect metrics
		var rm metricdata.ResourceMetrics
		if err := reader.Collect(ctx, &rm); err != nil {
			t.Fatalf("failed to collect metrics: %v", err)
		}

		// Verify metrics were collected
		if len(rm.ScopeMetrics) == 0 {
			t.Fatal("expected scope metrics")
		}

		if len(rm.ScopeMetrics[0].Metrics) == 0 {
			t.Fatal("expected at least one metric")
		}

		// Shutdown
		if err := mp.Shutdown(ctx); err != nil {
			t.Fatalf("failed to shutdown meter provider: %v", err)
		}
	})
}

func TestOtelMeter_Histogram(t *testing.T) {
	reader := metric.NewManualReader()
	mp := metric.NewMeterProvider(
		metric.WithReader(reader),
	)

	meter := &otelMeter{
		meter: mp.Meter("test"),
	}

	ctx := context.Background()

	t.Run("creates and uses histogram", func(t *testing.T) {
		histogram := meter.Histogram("test-histogram")

		if histogram == nil {
			t.Fatal("expected non-nil histogram")
		}

		// Record some values
		histogram.Record(ctx, 10.5, hyperion.String("endpoint", "/api"))
		histogram.Record(ctx, 25.3, hyperion.String("endpoint", "/api"))
		histogram.Record(ctx, 15.7, hyperion.String("endpoint", "/api"))

		// Collect metrics
		var rm metricdata.ResourceMetrics
		if err := reader.Collect(ctx, &rm); err != nil {
			t.Fatalf("failed to collect metrics: %v", err)
		}

		// Verify metrics were collected
		if len(rm.ScopeMetrics) == 0 {
			t.Fatal("expected scope metrics")
		}

		if len(rm.ScopeMetrics[0].Metrics) == 0 {
			t.Fatal("expected at least one metric")
		}

		// Shutdown
		if err := mp.Shutdown(ctx); err != nil {
			t.Fatalf("failed to shutdown meter provider: %v", err)
		}
	})
}

func TestOtelMeter_Gauge(t *testing.T) {
	reader := metric.NewManualReader()
	mp := metric.NewMeterProvider(
		metric.WithReader(reader),
	)

	meter := &otelMeter{
		meter: mp.Meter("test"),
	}

	ctx := context.Background()

	t.Run("creates and uses gauge", func(t *testing.T) {
		gauge := meter.Gauge("test-gauge")

		if gauge == nil {
			t.Fatal("expected non-nil gauge")
		}

		// Record some values (using histogram internally)
		gauge.Record(ctx, 42.0, hyperion.String("resource", "memory"))
		gauge.Record(ctx, 45.5, hyperion.String("resource", "memory"))

		// Collect metrics
		var rm metricdata.ResourceMetrics
		if err := reader.Collect(ctx, &rm); err != nil {
			t.Fatalf("failed to collect metrics: %v", err)
		}

		// Verify metrics were collected
		if len(rm.ScopeMetrics) == 0 {
			t.Fatal("expected scope metrics")
		}

		if len(rm.ScopeMetrics[0].Metrics) == 0 {
			t.Fatal("expected at least one metric")
		}

		// Shutdown
		if err := mp.Shutdown(ctx); err != nil {
			t.Fatalf("failed to shutdown meter provider: %v", err)
		}
	})
}

func TestOtelMeter_UpDownCounter(t *testing.T) {
	reader := metric.NewManualReader()
	mp := metric.NewMeterProvider(
		metric.WithReader(reader),
	)

	meter := &otelMeter{
		meter: mp.Meter("test"),
	}

	ctx := context.Background()

	t.Run("creates and uses up-down counter", func(t *testing.T) {
		upDownCounter := meter.UpDownCounter("test-updown")

		if upDownCounter == nil {
			t.Fatal("expected non-nil up-down counter")
		}

		// Add and subtract values
		upDownCounter.Add(ctx, 10, hyperion.String("pool", "connections"))
		upDownCounter.Add(ctx, 5, hyperion.String("pool", "connections"))
		upDownCounter.Add(ctx, -3, hyperion.String("pool", "connections"))

		// Collect metrics
		var rm metricdata.ResourceMetrics
		if err := reader.Collect(ctx, &rm); err != nil {
			t.Fatalf("failed to collect metrics: %v", err)
		}

		// Verify metrics were collected
		if len(rm.ScopeMetrics) == 0 {
			t.Fatal("expected scope metrics")
		}

		if len(rm.ScopeMetrics[0].Metrics) == 0 {
			t.Fatal("expected at least one metric")
		}

		// Shutdown
		if err := mp.Shutdown(ctx); err != nil {
			t.Fatalf("failed to shutdown meter provider: %v", err)
		}
	})
}

func TestOtelMeter_MultipleMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	mp := metric.NewMeterProvider(
		metric.WithReader(reader),
	)

	meter := &otelMeter{
		meter: mp.Meter("test"),
	}

	ctx := context.Background()

	t.Run("handles multiple metric types", func(t *testing.T) {
		counter := meter.Counter("requests")
		histogram := meter.Histogram("latency")
		upDownCounter := meter.UpDownCounter("connections")

		// Use all metrics
		counter.Add(ctx, 1, hyperion.String("path", "/api"))
		histogram.Record(ctx, 123.45, hyperion.String("path", "/api"))
		upDownCounter.Add(ctx, 1, hyperion.String("pool", "main"))

		// Collect metrics
		var rm metricdata.ResourceMetrics
		if err := reader.Collect(ctx, &rm); err != nil {
			t.Fatalf("failed to collect metrics: %v", err)
		}

		// Verify all metrics were collected
		if len(rm.ScopeMetrics) == 0 {
			t.Fatal("expected scope metrics")
		}

		metrics := rm.ScopeMetrics[0].Metrics
		if len(metrics) != 3 {
			t.Errorf("expected 3 metrics, got %d", len(metrics))
		}

		// Shutdown
		if err := mp.Shutdown(ctx); err != nil {
			t.Fatalf("failed to shutdown meter provider: %v", err)
		}
	})
}
