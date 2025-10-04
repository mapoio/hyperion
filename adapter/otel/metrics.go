package otel

import (
	"context"

	"github.com/mapoio/hyperion"

	"go.opentelemetry.io/otel/metric"
)

// otelCounter wraps an OpenTelemetry Int64Counter to implement hyperion.Counter.
type otelCounter struct {
	counter metric.Int64Counter
}

// Add increments the counter by the given value with optional attributes.
func (c *otelCounter) Add(ctx context.Context, value int64, attrs ...hyperion.Attribute) {
	otelAttrs := metric.WithAttributes(convertAttributes(attrs...)...)
	c.counter.Add(ctx, value, otelAttrs)
}

// otelHistogram wraps an OpenTelemetry Float64Histogram to implement hyperion.Histogram.
type otelHistogram struct {
	histogram metric.Float64Histogram
}

// Record records a measurement with optional attributes.
func (h *otelHistogram) Record(ctx context.Context, value float64, attrs ...hyperion.Attribute) {
	otelAttrs := metric.WithAttributes(convertAttributes(attrs...)...)
	h.histogram.Record(ctx, value, otelAttrs)
}

// otelGauge wraps an OpenTelemetry Float64Histogram to implement hyperion.Gauge.
// Note: OpenTelemetry doesn't have a synchronous Gauge, so we use Float64Histogram instead.
type otelGauge struct {
	histogram metric.Float64Histogram
}

// Record records a gauge measurement with optional attributes.
func (g *otelGauge) Record(ctx context.Context, value float64, attrs ...hyperion.Attribute) {
	otelAttrs := metric.WithAttributes(convertAttributes(attrs...)...)
	g.histogram.Record(ctx, value, otelAttrs)
}

// otelUpDownCounter wraps an OpenTelemetry Int64UpDownCounter to implement hyperion.UpDownCounter.
type otelUpDownCounter struct {
	counter metric.Int64UpDownCounter
}

// Add adds the value to the up-down counter with optional attributes.
func (u *otelUpDownCounter) Add(ctx context.Context, value int64, attrs ...hyperion.Attribute) {
	otelAttrs := metric.WithAttributes(convertAttributes(attrs...)...)
	u.counter.Add(ctx, value, otelAttrs)
}
