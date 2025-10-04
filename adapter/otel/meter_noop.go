package otel

import (
	"context"

	"github.com/mapoio/hyperion"
)

// No-op implementations for metric instruments when creation fails.
// These prevent panics when OTel SDK returns errors during instrument creation.

type noOpCounter struct{}

func (c *noOpCounter) Add(ctx context.Context, value int64, attrs ...hyperion.Attribute) {
	// No-op: silently drop the metric
}

type noOpHistogram struct{}

func (h *noOpHistogram) Record(ctx context.Context, value float64, attrs ...hyperion.Attribute) {
	// No-op: silently drop the metric
}

type noOpGauge struct{}

func (g *noOpGauge) Record(ctx context.Context, value float64, attrs ...hyperion.Attribute) {
	// No-op: silently drop the metric
}

type noOpUpDownCounter struct{}

func (u *noOpUpDownCounter) Add(ctx context.Context, value int64, attrs ...hyperion.Attribute) {
	// No-op: silently drop the metric
}
