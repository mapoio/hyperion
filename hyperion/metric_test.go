package hyperion

import (
	"context"
	"testing"
)

// TestNoOpMeter tests the no-op meter implementation.
func TestNoOpMeter(t *testing.T) {
	meter := NewNoOpMeter()
	ctx := context.Background()

	// Counter
	counter := meter.Counter("test.counter")
	counter.Add(ctx, 1)
	counter.Add(ctx, 5, String("label", "value"))

	// Histogram
	histogram := meter.Histogram("test.histogram")
	histogram.Record(ctx, 42.5)
	histogram.Record(ctx, 100.0, String("label", "value"))

	// Gauge
	gauge := meter.Gauge("test.gauge")
	gauge.Record(ctx, 1024.0)
	gauge.Record(ctx, 2048.0, String("label", "value"))

	// UpDownCounter
	upDownCounter := meter.UpDownCounter("test.updowncounter")
	upDownCounter.Add(ctx, 1)
	upDownCounter.Add(ctx, -1, String("label", "value"))

	// If we reach here without panicking, the test passes
	t.Log("No-op meter works without errors")
}
