package hyperion

import (
	"context"
	"testing"
)

// TestMetricOption_WithMetricDescription tests WithMetricDescription option.
func TestMetricOption_WithMetricDescription(t *testing.T) {
	opt := WithMetricDescription("request count")

	if opt == nil {
		t.Fatal("WithMetricDescription() returned nil")
	}

	// Apply to a metric config
	config := &metricConfig{}
	opt.applyMetric(config)

	if config.Description != "request count" {
		t.Errorf("Description = %q, want %q", config.Description, "request count")
	}
}

// TestMetricOption_WithMetricUnit tests WithMetricUnit option.
func TestMetricOption_WithMetricUnit(t *testing.T) {
	opt := WithMetricUnit("milliseconds")

	if opt == nil {
		t.Fatal("WithMetricUnit() returned nil")
	}

	// Apply to a metric config
	config := &metricConfig{}
	opt.applyMetric(config)

	if config.Unit != "milliseconds" {
		t.Errorf("Unit = %q, want %q", config.Unit, "milliseconds")
	}
}

// TestNoOpMeter_WithOptions tests that NoOp meter accepts options without error.
func TestNoOpMeter_WithOptions(t *testing.T) {
	meter := NewNoOpMeter()
	ctx := context.Background()

	// Test Counter with options
	counter := meter.Counter("test.counter",
		WithMetricDescription("test counter"),
		WithMetricUnit("count"),
	)
	if counter == nil {
		t.Fatal("Counter() with options returned nil")
	}
	counter.Add(ctx, 1) // Should not panic

	// Test Histogram with options
	histogram := meter.Histogram("test.histogram",
		WithMetricDescription("test histogram"),
		WithMetricUnit("ms"),
	)
	if histogram == nil {
		t.Fatal("Histogram() with options returned nil")
	}
	histogram.Record(ctx, 100.0) // Should not panic

	// Test Gauge with options
	gauge := meter.Gauge("test.gauge",
		WithMetricDescription("test gauge"),
		WithMetricUnit("bytes"),
	)
	if gauge == nil {
		t.Fatal("Gauge() with options returned nil")
	}
	gauge.Record(ctx, 1024.0) // Should not panic

	// Test UpDownCounter with options
	upDownCounter := meter.UpDownCounter("test.updown",
		WithMetricDescription("test updown counter"),
		WithMetricUnit("items"),
	)
	if upDownCounter == nil {
		t.Fatal("UpDownCounter() with options returned nil")
	}
	upDownCounter.Add(ctx, -5) // Should not panic
}
