# OpenTelemetry Adapter for Hyperion

This adapter provides OpenTelemetry integration for the Hyperion framework, offering distributed tracing and metrics collection capabilities.

## Features

- **Unified Provider Architecture**: Single `otelProvider` manages both TracerProvider and MeterProvider with shared resource configuration
- **Trace Context Propagation**: Automatic trace context injection into logs (trace_id, span_id)
- **Multiple Exporters**:
  - **Tracing**: OTLP, Jaeger (via OTLP)
  - **Metrics**: Prometheus, OTLP (planned)
- **Graceful Shutdown**: Unified lifecycle management via fx hooks
- **Type-Safe API**: Clean abstractions over OpenTelemetry SDK
- **High Test Coverage**: 84.5% test coverage with comprehensive integration tests

## Architecture

### Unified Provider Pattern

The adapter uses a singleton provider pattern to ensure consistent resource attributes across all telemetry signals:

```
┌─────────────────────────────────────┐
│        otelProvider                 │
│  (Singleton, sync.Once)             │
├─────────────────────────────────────┤
│  - serviceName: string              │
│  - resource: *resource.Resource ←───┼─── Shared across signals
│  - tracerProvider                   │
│  - meterProvider                    │
└─────────────────────────────────────┘
         │                    │
         ↓                    ↓
    TracerProvider      MeterProvider
```

### Integration with Zap Logger

When using the Zap adapter alongside OTel:

1. **Tracer** creates spans with trace context
2. **Zap Logger** extracts trace_id and span_id from context
3. **Logs** include structured fields for correlation

```go
ctx := context.Background()
newCtx, span := tracer.Start(ctx, "operation")
defer span.End()

// Logger automatically includes trace_id and span_id
logger.InfoContext(newCtx, "processing request")
// Output: {"level":"info","msg":"processing request","trace_id":"...","span_id":"..."}
```

## Installation

```bash
go get github.com/mapoio/hyperion/adapter/otel
```

## Configuration

### Tracing Configuration

```yaml
tracing:
  enabled: true
  service_name: "my-service"
  exporter: "otlp"           # "otlp" or "jaeger"
  endpoint: "localhost:4317"
  sample_rate: 1.0           # 0.0 - 1.0 (1.0 = 100%)
  attributes:
    environment: "production"
    version: "1.0.0"
```

### Metrics Configuration

```yaml
metrics:
  enabled: true
  service_name: "my-service"
  exporter: "prometheus"     # "prometheus" or "otlp" (planned)
  interval: 10s
  attributes:
    environment: "production"
```

## Usage

### Basic Setup with fx

```go
package main

import (
    "github.com/mapoio/hyperion"
    "github.com/mapoio/hyperion/adapter/otel"
    "go.uber.org/fx"
)

func main() {
    fx.New(
        // Provide configuration
        fx.Provide(NewConfig),

        // Use OTel module (provides both Tracer and Meter)
        otel.Module,

        // Your application code
        fx.Invoke(Run),
    ).Run()
}

func Run(tracer hyperion.Tracer, meter hyperion.Meter) {
    // Use tracer and meter
}
```

### Tracer Only

```go
fx.New(
    fx.Provide(NewConfig),
    otel.TracerModule,  // Only tracer
    fx.Invoke(Run),
).Run()
```

### Meter Only

```go
fx.New(
    fx.Provide(NewConfig),
    otel.MeterModule,   // Only meter
    fx.Invoke(Run),
).Run()
```

### Creating Spans

```go
func ProcessRequest(ctx hyperion.Context, tracer hyperion.Tracer) error {
    // Start a new span
    newCtx, span := tracer.Start(ctx, "process-request")
    defer span.End()

    // Add attributes
    span.SetAttributes(map[string]any{
        "user.id": 123,
        "request.method": "POST",
    })

    // Record events
    span.AddEvent("validation completed")

    // Record errors
    if err := validate(); err != nil {
        span.RecordError(err)
        return err
    }

    return nil
}
```

### Creating Metrics

```go
func SetupMetrics(meter hyperion.Meter) {
    // Counter: monotonically increasing value
    requestCounter := meter.Counter("http.requests.total")
    requestCounter.Add(ctx, 1, map[string]any{
        "method": "GET",
        "status": 200,
    })

    // Histogram: statistical distribution
    latencyHistogram := meter.Histogram("http.request.duration")
    latencyHistogram.Record(ctx, 0.123, map[string]any{
        "endpoint": "/api/users",
    })

    // Gauge: current value
    activeConnections := meter.Gauge("system.active_connections")
    activeConnections.Record(ctx, 42)

    // UpDownCounter: can increase or decrease
    queueSize := meter.UpDownCounter("queue.size")
    queueSize.Add(ctx, 5)  // Enqueue
    queueSize.Add(ctx, -3) // Dequeue
}
```

## Implementation Details

### Shared Resource Management

Both TracerProvider and MeterProvider share the same `resource.Resource` which includes:

- `service.name`: Service identifier
- Custom attributes from configuration

This ensures consistent correlation across traces and metrics.

### Singleton Provider

The `otelProvider` is a singleton initialized using `sync.Once`:

```go
var (
    globalProvider *otelProvider
    providerOnce   sync.Once
)

func getOrCreateProvider(serviceName string, attrs map[string]string) (*otelProvider, error) {
    var err error
    providerOnce.Do(func() {
        globalProvider, err = initProvider(serviceName, attrs)
    })
    return globalProvider, err
}
```

### Graceful Shutdown

The adapter registers a unified shutdown hook that gracefully shuts down both providers:

```go
func RegisterShutdownHook(lc fx.Lifecycle) {
    lc.Append(fx.Hook{
        OnStop: func(ctx context.Context) error {
            if globalProvider != nil {
                return globalProvider.shutdown(ctx)
            }
            return nil
        },
    })
}
```

## Testing

### Running Tests

```bash
# Run all tests
go test -v -cover ./...

# Run integration tests only
go test -v -run "Integration|Extraction|Multiple" ./...

# Generate coverage report
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Coverage

Current test coverage: **84.5%**

Test categories:
- Configuration loading and validation
- Exporter creation
- Tracer and Meter functionality
- Integration tests for trace context propagation
- fx module lifecycle

## Supported Exporters

### Tracing

| Exporter | Status | Notes |
|----------|--------|-------|
| OTLP     | ✅ Supported | Default, works with any OTLP-compatible backend |
| Jaeger   | ✅ Supported | Uses OTLP protocol (Jaeger native exporter deprecated) |
| Zipkin   | ❌ Not supported | Use OTLP with Zipkin backend |

### Metrics

| Exporter | Status | Notes |
|----------|--------|-------|
| Prometheus | ✅ Supported | Pull-based metrics via `/metrics` endpoint |
| OTLP | 🚧 Planned | Push-based metrics to OTLP collector |

## OpenTelemetry Logs Bridge

**Status**: Deferred

The OTel Logs API is still experimental in v1.38.0. We currently provide trace context injection into Zap logs (trace_id, span_id) which is sufficient for correlation. Full OTel Logs Bridge will be implemented when the API stabilizes.

Current approach:
- Zap's `otel_bridge.go` extracts trace context from `context.Context`
- Structured log fields include `trace_id` and `span_id`
- Logs can be correlated with traces in observability platforms

## Roadmap

- [ ] OTLP metrics exporter
- [ ] OTel Logs Bridge (when API is stable)
- [ ] TLS configuration for OTLP exporters
- [ ] Custom resource attributes handling
- [ ] Sampling strategies configuration
- [ ] Exemplars support for metrics

## References

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [OpenTelemetry Go SDK](https://github.com/open-telemetry/opentelemetry-go)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
- [Hyperion Framework](https://github.com/mapoio/hyperion)

## License

This adapter is part of the Hyperion framework and follows the same license.
