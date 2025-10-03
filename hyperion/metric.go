package hyperion

import "context"

// Meter provides access to metric instruments for observability.
// It follows OpenTelemetry Metrics API semantics but doesn't depend on it.
//
// Metrics are used to record quantitative data about your application:
//   - Counters: monotonically increasing values (requests, errors)
//   - Histograms: distribution of values (latency, size)
//   - Gauges: current values that go up/down (memory, connections)
//   - UpDownCounters: values that can increase/decrease (queue depth)
//
// All metric recording methods accept a context.Context as the first parameter.
// When using hyperion.Context with OpenTelemetry, this enables:
//   - Automatic trace correlation via exemplars
//   - Metrics can be linked back to specific traces in your observability backend
type Meter interface {
	// Counter creates or retrieves a counter instrument.
	// Counters are monotonically increasing values (e.g., request count, error count).
	//
	// Example:
	//
	//	requestCounter := meter.Counter("http.requests",
	//	    hyperion.WithDescription("Total HTTP requests"),
	//	    hyperion.WithUnit("1"),
	//	)
	Counter(name string, opts ...MetricOption) Counter

	// Histogram creates or retrieves a histogram instrument.
	// Histograms record distribution of values (e.g., latency, request size).
	//
	// Example:
	//
	//	latencyHistogram := meter.Histogram("http.latency",
	//	    hyperion.WithDescription("HTTP request latency"),
	//	    hyperion.WithUnit("ms"),
	//	)
	Histogram(name string, opts ...MetricOption) Histogram

	// Gauge creates or retrieves a gauge instrument.
	// Gauges represent a current value that can go up or down (e.g., memory usage, queue depth).
	//
	// Example:
	//
	//	memoryGauge := meter.Gauge("process.memory",
	//	    hyperion.WithDescription("Process memory usage"),
	//	    hyperion.WithUnit("bytes"),
	//	)
	Gauge(name string, opts ...MetricOption) Gauge

	// UpDownCounter creates or retrieves an up-down counter.
	// Like counters but can decrease (e.g., active connections, items in queue).
	//
	// Example:
	//
	//	activeConns := meter.UpDownCounter("db.connections.active",
	//	    hyperion.WithDescription("Active database connections"),
	//	    hyperion.WithUnit("1"),
	//	)
	UpDownCounter(name string, opts ...MetricOption) UpDownCounter
}

// Counter is a monotonically increasing metric.
// Values must be non-negative.
type Counter interface {
	// Add increments the counter by the given value.
	// The context enables trace correlation in OpenTelemetry implementations.
	//
	// Example:
	//
	//	counter.Add(ctx, 1,
	//	    hyperion.String("method", "GetUser"),
	//	    hyperion.String("status", "success"),
	//	)
	Add(ctx context.Context, value int64, attrs ...Attribute)
}

// Histogram records a distribution of values.
// Useful for measuring latencies, sizes, or other distributions.
type Histogram interface {
	// Record adds a value to the histogram.
	// The context enables trace correlation in OpenTelemetry implementations.
	//
	// Example:
	//
	//	histogram.Record(ctx, 42.5,
	//	    hyperion.String("method", "GetUser"),
	//	)
	Record(ctx context.Context, value float64, attrs ...Attribute)
}

// Gauge represents a current value that can go up or down.
// Useful for measuring things like memory usage, queue depth, etc.
type Gauge interface {
	// Record sets the current gauge value.
	// The context enables trace correlation in OpenTelemetry implementations.
	//
	// Example:
	//
	//	gauge.Record(ctx, 1024.0,
	//	    hyperion.String("type", "heap"),
	//	)
	Record(ctx context.Context, value float64, attrs ...Attribute)
}

// UpDownCounter can increment and decrement.
// Useful for tracking values that can go up and down, like active connections.
type UpDownCounter interface {
	// Add changes the counter by the given value.
	// The value can be negative to decrement.
	// The context enables trace correlation in OpenTelemetry implementations.
	//
	// Example:
	//
	//	// Increment
	//	upDownCounter.Add(ctx, 1, hyperion.String("pool", "main"))
	//	// Decrement
	//	upDownCounter.Add(ctx, -1, hyperion.String("pool", "main"))
	Add(ctx context.Context, value int64, attrs ...Attribute)
}

// MetricOption configures a metric instrument.
type MetricOption interface {
	applyMetric(*metricConfig)
}

type metricConfig struct {
	Description string
	Unit        string
}

// WithMetricDescription sets the metric description.
// This helps document what the metric measures.
//
// Example:
//
//	meter.Counter("requests",
//	    hyperion.WithMetricDescription("Total number of requests"),
//	)
func WithMetricDescription(desc string) MetricOption {
	return metricOptionFunc(func(cfg *metricConfig) {
		cfg.Description = desc
	})
}

// WithMetricUnit sets the metric unit.
// Common units: "1" (dimensionless), "ms" (milliseconds), "bytes", "s" (seconds).
//
// Example:
//
//	meter.Histogram("latency",
//	    hyperion.WithMetricUnit("ms"),
//	)
func WithMetricUnit(unit string) MetricOption {
	return metricOptionFunc(func(cfg *metricConfig) {
		cfg.Unit = unit
	})
}

type metricOptionFunc func(*metricConfig)

func (f metricOptionFunc) applyMetric(cfg *metricConfig) {
	f(cfg)
}
