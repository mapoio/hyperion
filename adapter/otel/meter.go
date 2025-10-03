package otel

import (
	"github.com/mapoio/hyperion"
	"go.opentelemetry.io/otel/metric"
)

// otelMeter wraps an OpenTelemetry meter to implement hyperion.Meter.
type otelMeter struct {
	meter metric.Meter
}

// Counter creates or retrieves a counter instrument.
func (m *otelMeter) Counter(name string, opts ...hyperion.MetricOption) hyperion.Counter {
	// For now, create without options - we'll add option support later
	counter, err := m.meter.Int64Counter(name)
	if err != nil {
		// In production, we should handle this error properly
		// For now, return a no-op counter
		return &otelCounter{counter: counter}
	}
	return &otelCounter{counter: counter}
}

// Histogram creates or retrieves a histogram instrument.
func (m *otelMeter) Histogram(name string, opts ...hyperion.MetricOption) hyperion.Histogram {
	histogram, err := m.meter.Float64Histogram(name)
	if err != nil {
		return &otelHistogram{histogram: histogram}
	}
	return &otelHistogram{histogram: histogram}
}

// Gauge creates or retrieves a gauge instrument.
func (m *otelMeter) Gauge(name string, opts ...hyperion.MetricOption) hyperion.Gauge {
	// Use histogram for synchronous gauge-like behavior
	histogram, err := m.meter.Float64Histogram(name)
	if err != nil {
		return &otelGauge{histogram: histogram}
	}
	return &otelGauge{histogram: histogram}
}

// UpDownCounter creates or retrieves an up-down counter.
func (m *otelMeter) UpDownCounter(name string, opts ...hyperion.MetricOption) hyperion.UpDownCounter {
	counter, err := m.meter.Int64UpDownCounter(name)
	if err != nil {
		return &otelUpDownCounter{counter: counter}
	}
	return &otelUpDownCounter{counter: counter}
}
