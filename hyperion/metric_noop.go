package hyperion

import (
	"context"
	"fmt"
)

// noOpMeter is a no-op implementation of Meter.
// It prints diagnostic messages when used but doesn't record any metrics.
type noOpMeter struct{}

// NewNoOpMeter creates a no-op meter for testing or when metrics are disabled.
func NewNoOpMeter() Meter {
	fmt.Println("[Hyperion] Using no-op Meter")
	return &noOpMeter{}
}

func (m *noOpMeter) Counter(name string, opts ...MetricOption) Counter {
	return &noOpCounter{name: name}
}

func (m *noOpMeter) Histogram(name string, opts ...MetricOption) Histogram {
	return &noOpHistogram{name: name}
}

func (m *noOpMeter) Gauge(name string, opts ...MetricOption) Gauge {
	return &noOpGauge{name: name}
}

func (m *noOpMeter) UpDownCounter(name string, opts ...MetricOption) UpDownCounter {
	return &noOpUpDownCounter{name: name}
}

// noOpCounter is a no-op counter.
type noOpCounter struct {
	name string
}

func (c *noOpCounter) Add(ctx context.Context, value int64, attrs ...Attribute) {
	// No-op
}

// noOpHistogram is a no-op histogram.
type noOpHistogram struct {
	name string
}

func (h *noOpHistogram) Record(ctx context.Context, value float64, attrs ...Attribute) {
	// No-op
}

// noOpGauge is a no-op gauge.
type noOpGauge struct {
	name string
}

func (g *noOpGauge) Record(ctx context.Context, value float64, attrs ...Attribute) {
	// No-op
}

// noOpUpDownCounter is a no-op up-down counter.
type noOpUpDownCounter struct {
	name string
}

func (u *noOpUpDownCounter) Add(ctx context.Context, value int64, attrs ...Attribute) {
	// No-op
}
