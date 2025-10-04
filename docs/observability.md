# Observability Design

This document describes Hyperion's observability architecture and how Logs, Traces, and Metrics are integrated through OpenTelemetry.

## Overview

Hyperion provides a unified observability solution that automatically correlates:
- **Logs** - Structured logging with automatic trace context
- **Traces** - Distributed tracing with OpenTelemetry
- **Metrics** - Performance metrics with trace exemplars

## Three Pillars of Observability

### 1. Logs (Logger)

**Interface**: `hyperion.Logger`

```go
type Logger interface {
    Debug(msg string, fields ...any)
    Info(msg string, fields ...any)
    Warn(msg string, fields ...any)
    Error(msg string, fields ...any)
    Fatal(msg string, fields ...any)

    With(fields ...any) Logger
    WithError(err error) Logger
}
```

**Automatic Trace Correlation**:
When using OpenTelemetry adapters (e.g., `adapter/zap` with OTel Logs Bridge), logs automatically include:
- `trace_id` - Links log to trace
- `span_id` - Links log to specific span

**Usage**:
```go
func (s *Service) GetUser(ctx hyperion.Context, id string) error {
    // Logs automatically include trace_id and span_id
    ctx.Logger().Info("fetching user", "user_id", id)
    return nil
}
```

### 2. Traces (Tracer)

**Interface**: `hyperion.Tracer`

```go
type Tracer interface {
    Start(ctx context.Context, spanName string, opts ...SpanOption) (context.Context, Span)
}

type Span interface {
    End(opts ...SpanEndOption)
    SetAttributes(attrs ...Attribute)
    RecordError(err error, opts ...EventOption)
    AddEvent(name string, opts ...EventOption)
    SpanContext() SpanContext
}
```

**Automatic Span Creation**:
When using `UseIntercept` with `TracingInterceptorModule`, spans are automatically created:

```go
func (s *Service) GetUser(ctx hyperion.Context, id string) (err error) {
    // Automatically creates span "Service.GetUser"
    ctx, end := ctx.UseIntercept("Service", "GetUser")
    defer end(&err)

    // Business logic
    return nil
}
```

**Manual Span Creation**:
```go
func (s *Service) complexOperation(ctx hyperion.Context) error {
    // Create child span for specific operation
    newCtx, span := ctx.Tracer().Start(ctx, "database.query")
    defer span.End()

    span.SetAttributes(
        hyperion.String("query", "SELECT * FROM users"),
        hyperion.Int("timeout", 30),
    )

    // Use newCtx for nested operations
    return s.repository.Query(newCtx, query)
}
```

### 3. Metrics (Meter)

**Interface**: `hyperion.Meter`

```go
type Meter interface {
    Counter(name string, opts ...MetricOption) Counter
    Histogram(name string, opts ...MetricOption) Histogram
    Gauge(name string, opts ...MetricOption) Gauge
    UpDownCounter(name string, opts ...MetricOption) UpDownCounter
}
```

**Automatic Trace Correlation via Exemplars**:
When using OpenTelemetry adapters, metrics automatically record exemplars linking to traces:

```go
func (s *Service) GetUser(ctx hyperion.Context, id string) (err error) {
    ctx, end := ctx.UseIntercept("Service", "GetUser")
    defer end(&err)

    // Counter with automatic trace exemplar
    requestCounter := ctx.Meter().Counter("user.requests",
        hyperion.WithMetricDescription("Total user requests"),
        hyperion.WithMetricUnit("1"),
    )
    requestCounter.Add(ctx, 1,
        hyperion.String("method", "GetUser"),
        hyperion.String("status", "success"),
    )

    // Histogram with automatic trace exemplar
    latency := ctx.Meter().Histogram("user.latency",
        hyperion.WithMetricDescription("User request latency"),
        hyperion.WithMetricUnit("ms"),
    )
    latency.Record(ctx, 42.5)

    return nil
}
```

### Exemplar Mechanism Deep Dive

**What are Exemplars?**

Exemplars are sample data points attached to metric observations that link metrics back to the specific trace that generated them. This creates a bidirectional correlation:
- **Metrics → Traces**: Click on a metric spike to see example traces
- **Traces → Metrics**: See which metrics were recorded during trace execution

**How Exemplars Work in Hyperion**:

1. **Automatic Extraction**: When you pass `hyperion.Context` to metric recording methods, the OTel adapter automatically extracts the current trace context:

```go
// Inside adapter/otel/meter.go
func (c *otelCounter) Add(ctx context.Context, value int64, attrs ...hyperion.Attribute) {
    // OTel SDK automatically:
    // 1. Checks if ctx contains a span context
    // 2. If yes, attaches trace_id + span_id as exemplar
    // 3. Records the exemplar with timestamp
    c.counter.Add(ctx, value, metric.WithAttributes(convertAttributes(attrs)...))
}
```

2. **Storage Format**: Exemplars are stored alongside metric data points in Prometheus/Mimir format:

```
# Metric without exemplar (no context)
http_requests_total{method="GET",path="/users"} 1523

# Metric with exemplar (with context) - shows example trace
http_requests_total{method="GET",path="/users"} 1523 # {trace_id="4bf92f3577b34da6a3ce929d0e0e4736",span_id="00f067aa0ba902b7"} 1.0 1672531200000
```

3. **Sampling Strategy**: Not every metric observation needs an exemplar (would explode storage). OTel SDK automatically samples exemplars:
   - **Default**: Last exemplar per time bucket (e.g., last trace per minute)
   - **Configurable**: Can configure to keep exemplars for high-latency requests, errors, etc.

**Use Cases**:

**Use Case 1: Debug Latency Spikes**
```
1. Notice p99 latency spike in Grafana dashboard
2. Click on the spike → See exemplar trace ID
3. Open trace in Jaeger/Tempo
4. See exactly what caused the slow request
```

**Use Case 2: Debug Error Rate Increase**
```
1. Notice error rate increase in metrics
2. Find exemplar trace for failed request
3. See error stack trace and context
4. Identify root cause
```

**Use Case 3: Capacity Planning**
```
1. Notice high request count on specific endpoint
2. Click exemplar to see typical trace
3. Analyze resource usage per request
4. Calculate infrastructure needs
```

**Configuration in Hyperion**:

```go
// Example: Configure exemplar sampling (future feature in Epic 6)
otel.Module,  // Uses OTel SDK default exemplar configuration

// Default behavior:
// - Automatically enabled for all metrics when context is passed
// - Samples 1 exemplar per time series per collection interval
// - Stores trace_id, span_id, and timestamp
```

**Visualization in Grafana**:

When viewing metrics in Grafana with Tempo/Jaeger as trace datasource:
1. Metrics show as usual (counter, histogram, etc.)
2. Data points with exemplars show a **special icon** (usually a diamond)
3. Click the icon → Opens trace in Tempo/Jaeger in a new panel
4. Navigate between metrics and traces seamlessly

**Backend Requirements**:

For exemplars to work, you need:
- ✅ **Metrics Backend**: Prometheus (>= 2.26), Mimir, or OTLP-compatible collector
- ✅ **Traces Backend**: Tempo, Jaeger, or any OTLP trace backend
- ✅ **Visualization**: Grafana (>= 7.5) with both datasources configured
- ✅ **Exporter**: OTLP exporter with exemplar support (included in `adapter/otel`)

**Performance Impact**:

Exemplars have minimal overhead:
- **Storage**: ~50 bytes per exemplar (trace_id + span_id + timestamp)
- **CPU**: Negligible (< 0.1% increase)
- **Network**: Only sampled exemplars exported (default: 1 per metric per interval)

**Example: End-to-End Flow**

```go
// 1. Service method with tracing
func (s *OrderService) CreateOrder(ctx hyperion.Context, req *CreateOrderRequest) (err error) {
    // UseIntercept creates span with trace_id
    ctx, end := ctx.UseIntercept("OrderService", "CreateOrder")
    defer end(&err)

    // 2. Record metrics with context (automatic exemplar)
    orderCounter := ctx.Meter().Counter("orders.created")
    orderCounter.Add(ctx, 1,
        hyperion.String("payment_method", req.PaymentMethod),
    )

    // 3. Record latency with exemplar
    start := time.Now()
    result, err := s.createOrder(ctx, req)
    duration := time.Since(start)

    latencyHistogram := ctx.Meter().Histogram("orders.creation_latency")
    latencyHistogram.Record(ctx, float64(duration.Milliseconds()))

    return err
}

// Result in Prometheus:
// orders_created_total{payment_method="credit_card"} 1523
//   # {trace_id="4bf92f3577b34da6a3ce929d0e0e4736"} 1.0 1672531200000

// In Grafana:
// 1. View orders_created_total metric
// 2. See spike at timestamp 1672531200000
// 3. Click exemplar diamond icon
// 4. Opens trace "4bf92f3577b34da6a3ce929d0e0e4736" in Tempo
// 5. See full CreateOrder execution with all spans
```

**See Also**:
- [OpenTelemetry Metrics Exemplars Spec](https://opentelemetry.io/docs/specs/otel/metrics/data-model/#exemplars)
- [Prometheus Exemplars Documentation](https://prometheus.io/docs/prometheus/latest/feature_flags/#exemplars-storage)
- [Grafana Exemplars Guide](https://grafana.com/docs/grafana/latest/fundamentals/exemplars/)

## Architecture: How They Work Together

### Correlation Flow

```
HTTP Request
    ↓
[Tracer] Creates Trace ID + Root Span
    ↓
[Context] Embeds trace context in context.Context
    ↓
┌─────────────────────────────────────────┐
│  hyperion.Context (with trace context)  │
│                                         │
│  ┌──────────┐  ┌──────────┐  ┌───────┐│
│  │ Logger() │  │ Tracer() │  │Meter()││
│  └────┬─────┘  └────┬─────┘  └───┬───┘│
│       │             │              │   │
│       └─────────────┴──────────────┘   │
│              Shared context.Context    │
│         (contains TraceID + SpanID)    │
└─────────────────────────────────────────┘
    ↓               ↓              ↓
[Logs]          [Spans]        [Metrics]
trace_id        parent_id      exemplars
span_id         span_id        → trace_id
```

### OpenTelemetry Backend Correlation

**In Jaeger/Tempo (Traces)**:
```
Trace: abc123...
└─ Span: Service.GetUser (def456...)
   ├─ Log: "fetching user" (auto-linked)
   ├─ Metric: user.requests exemplar → trace_id
   └─ Span: database.query (xyz789...)
```

**In Loki/Elasticsearch (Logs)**:
```json
{
  "timestamp": "2025-01-03T10:00:00Z",
  "level": "info",
  "message": "fetching user",
  "user_id": "123",
  "trace_id": "abc123...",   // Auto-embedded
  "span_id": "def456..."     // Auto-embedded
}
```

**In Prometheus/Mimir (Metrics)**:
```
user_requests_total{method="GetUser",status="success"} 1
# Exemplar: trace_id="abc123..." timestamp=...

user_latency_bucket{le="50"} 1
# Exemplar: trace_id="abc123..." timestamp=...
```

## Implementation Strategies

### Strategy 1: Framework-Only Integration (Simple)

Use Hyperion's core interfaces without OpenTelemetry:

```go
fx.New(
    hyperion.CoreModule,  // Uses no-op implementations
    myapp.Module,
).Run()
```

**Result**:
- ✅ Code works without changes
- ✅ No external dependencies
- ❌ No correlation between logs/traces/metrics

### Strategy 2: OpenTelemetry Integration (Recommended)

Use OpenTelemetry adapters for full correlation:

```go
fx.New(
    hyperion.CoreModule,
    hyperion.TracingInterceptorModule,  // Auto-create spans
    hyperion.LoggingInterceptorModule,  // Auto-log method calls

    zap.Module,   // Logger with OTel Logs Bridge
    otel.Module,  // Tracer + Meter with OTel SDK

    myapp.Module,
).Run()
```

**Result**:
- ✅ Automatic correlation
- ✅ Logs → Traces navigation
- ✅ Metrics → Traces navigation
- ✅ Unified observability backend

### Strategy 3: Selective Integration

Mix and match based on needs:

```go
fx.New(
    hyperion.CoreModule,
    hyperion.TracingInterceptorModule,  // Only tracing

    zap.Module,   // Structured logging (no OTel bridge)
    otel.Module,  // Only Tracer

    myapp.Module,
).Run()
```

## Adapter Implementation Guide

### Logger Adapter with OTel Logs Bridge

```go
// adapter/zap/logger.go
import (
    "go.opentelemetry.io/contrib/bridges/otelzap"
    "go.uber.org/zap"
)

type zapLogger struct {
    logger     *zap.Logger
    otelLogger *otelzap.Logger  // OTel Logs Bridge
}

func NewLogger() hyperion.Logger {
    zapLogger, _ := zap.NewProduction()
    otelLogger := otelzap.New(zapLogger)

    return &zapLogger{
        logger:     zapLogger,
        otelLogger: otelLogger,
    }
}

// Implement context-aware logging
func (l *zapLogger) InfoContext(ctx context.Context, msg string, fields ...any) {
    // OTel Logs Bridge automatically extracts span context
    l.otelLogger.InfoContext(ctx, msg, convertFields(fields)...)
}
```

### Tracer Adapter with OTel SDK

```go
// adapter/otel/tracer.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

type otelTracer struct {
    tracer trace.Tracer
}

func NewTracer() hyperion.Tracer {
    return &otelTracer{
        tracer: otel.Tracer("hyperion"),
    }
}

func (t *otelTracer) Start(ctx context.Context, spanName string, opts ...hyperion.SpanOption) (context.Context, hyperion.Span) {
    ctx, span := t.tracer.Start(ctx, spanName, convertOptions(opts)...)
    return ctx, &otelSpan{span: span}
}
```

### Meter Adapter with OTel SDK

```go
// adapter/otel/meter.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/metric"
)

type otelMeter struct {
    meter metric.Meter
}

func NewMeter() hyperion.Meter {
    return &otelMeter{
        meter: otel.Meter("hyperion"),
    }
}

func (m *otelMeter) Counter(name string, opts ...hyperion.MetricOption) hyperion.Counter {
    counter, _ := m.meter.Int64Counter(name, convertOptions(opts)...)
    return &otelCounter{counter: counter}
}

func (c *otelCounter) Add(ctx context.Context, value int64, attrs ...hyperion.Attribute) {
    // OTel SDK automatically extracts trace context and creates exemplars
    c.counter.Add(ctx, value, convertAttributes(attrs)...)
}
```

## Best Practices

### 1. Always Pass Context

```go
// ✅ Good: Pass context through the call chain
func (s *Service) GetUser(ctx hyperion.Context, id string) error {
    ctx.Logger().Info("fetching user")  // Auto-correlated
    ctx.Meter().Counter("requests").Add(ctx, 1)  // Auto-correlated
    return s.repository.GetUser(ctx, id)
}

// ❌ Bad: Don't create new context
func (s *Service) GetUser(ctx hyperion.Context, id string) error {
    newCtx := context.Background()  // Loses trace context!
    return s.repository.GetUser(newCtx, id)
}
```

### 2. Use Interceptors for Automatic Instrumentation

```go
// ✅ Good: UseIntercept creates spans automatically
func (s *Service) GetUser(ctx hyperion.Context, id string) (err error) {
    ctx, end := ctx.UseIntercept("Service", "GetUser")
    defer end(&err)

    // Business logic
    return nil
}

// ⚠️ Less optimal: Manual span management
func (s *Service) GetUser(ctx hyperion.Context, id string) error {
    newCtx, span := ctx.Tracer().Start(ctx, "Service.GetUser")
    defer span.End()

    // Business logic
    return nil
}
```

### 3. Add Meaningful Attributes

```go
// ✅ Good: Rich context
ctx.Logger().Info("user created",
    "user_id", user.ID,
    "email", user.Email,
    "role", user.Role,
)

ctx.Meter().Counter("user.created").Add(ctx, 1,
    hyperion.String("role", user.Role),
    hyperion.String("plan", user.Plan),
)

// ❌ Bad: No context
ctx.Logger().Info("user created")
ctx.Meter().Counter("user.created").Add(ctx, 1)
```

### 4. Measure What Matters

```go
// ✅ Good: Actionable metrics
requestCounter := ctx.Meter().Counter("http.requests",
    hyperion.WithMetricDescription("Total HTTP requests"),
)
requestCounter.Add(ctx, 1,
    hyperion.String("method", "GET"),
    hyperion.String("path", "/users"),
    hyperion.String("status", "200"),
)

latencyHistogram := ctx.Meter().Histogram("http.latency",
    hyperion.WithMetricDescription("HTTP request latency"),
    hyperion.WithMetricUnit("ms"),
)
latencyHistogram.Record(ctx, duration.Milliseconds())

// ❌ Bad: Too granular, cardinality explosion
ctx.Meter().Counter("requests").Add(ctx, 1,
    hyperion.String("user_id", userID),  // High cardinality!
    hyperion.String("timestamp", time.Now().String()),  // Infinite cardinality!
)
```

## Performance Considerations

### No-Op Performance

When using no-op implementations (default), there is **zero overhead**:
- No allocations
- No I/O operations
- Inlined function calls

### OTel Performance

When using OpenTelemetry adapters:
- **Logs**: Asynchronous, buffered writes
- **Traces**: Sampling reduces overhead (default: 0.1%)
- **Metrics**: Aggregated in-memory, exported periodically

**Typical overhead**: < 1% CPU, < 50MB memory for 1000 req/sec

## Troubleshooting

### Logs Not Correlated with Traces

**Symptom**: Logs have trace_id but not visible in trace view

**Solution**: Ensure logger adapter implements OTel Logs Bridge
```go
// Check: Is logger using OTel Logs Bridge?
otelLogger := otelzap.New(zapLogger)  // ✅ Correct
```

### Metrics Missing Exemplars

**Symptom**: Metrics recorded but no trace links

**Solution**: Ensure context is passed to metric methods
```go
// ❌ Wrong: No context
counter.Add(1)

// ✅ Correct: Pass context
counter.Add(ctx, 1)
```

### TraceID Not Propagated

**Symptom**: New trace created for each service call

**Solution**: Ensure context flows through the call chain
```go
// ✅ Correct: Pass ctx through
func (s *Service) GetUser(ctx hyperion.Context, id string) error {
    return s.repository.GetUser(ctx, id)  // Pass ctx!
}
```

## References

- [OpenTelemetry Specification](https://opentelemetry.io/docs/specs/)
- [OTel Logs Bridge](https://opentelemetry.io/docs/specs/otel/logs/bridge-api/)
- [OTel Metrics Exemplars](https://opentelemetry.io/docs/specs/otel/metrics/api/#exemplars)
- [Hyperion Interceptor Design](./INTERCEPTOR.md)
