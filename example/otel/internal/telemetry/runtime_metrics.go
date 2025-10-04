package telemetry

import (
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/fx"
)

// EnableRuntimeMetrics configures automatic collection of Go runtime metrics.
// This includes CPU usage, memory statistics, goroutine count, and GC metrics.
//
// Metrics collected:
// - process.runtime.go.mem.heap_alloc
// - process.runtime.go.mem.heap_idle
// - process.runtime.go.mem.heap_inuse
// - process.runtime.go.mem.heap_objects
// - process.runtime.go.gc.count
// - process.runtime.go.gc.pause_ns
// - process.runtime.go.goroutines
// - process.runtime.go.cgo.calls
func EnableRuntimeMetrics(mp *metric.MeterProvider) error {
	// Start runtime metrics with custom configuration
	return runtime.Start(
		runtime.WithMeterProvider(mp),
		runtime.WithMinimumReadMemStatsInterval(time.Second),
	)
}

// RuntimeMetricsModule provides automatic Go runtime metrics collection.
var RuntimeMetricsModule = fx.Module("runtime_metrics",
	fx.Invoke(EnableRuntimeMetrics),
)
