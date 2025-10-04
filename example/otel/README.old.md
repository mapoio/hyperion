# Hyperion OTel Example with HyperDX

This example demonstrates a complete observability setup using:

- **Hyperion Framework** - Modular Go backend framework
- **OpenTelemetry Adapter** - Distributed tracing and metrics
- **Zap Adapter** - Structured logging with trace context injection
- **HyperDX** - Open-source observability platform (Traces + Logs + Metrics)

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Example Application                      │
│                                                             │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐              │
│  │   Gin    │──▶│ OTel     │──▶│  HyperDX │              │
│  │ HTTP     │   │ Tracer   │   │  OTLP    │              │
│  │ Server   │   │          │   │  Exporter│              │
│  └──────────┘   └──────────┘   └──────────┘              │
│       │              │                                      │
│       │         ┌────▼─────┐                               │
│       └────────▶│   Zap    │                               │
│                 │  Logger  │                               │
│                 └──────────┘                               │
│                      │                                      │
│                trace_id, span_id                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
                  ┌─────────────────┐
                  │    HyperDX      │
                  │  (localhost)    │
                  ├─────────────────┤
                  │ • Traces        │
                  │ • Logs          │
                  │ • Metrics       │
                  │ • Correlation   │
                  └─────────────────┘
```

## Features Demonstrated

### 1. Distributed Tracing

- ✅ Automatic span creation for HTTP requests
- ✅ Parent-child span relationships
- ✅ Span attributes (HTTP method, status, user-agent, etc.)
- ✅ Span events (milestones in request processing)
- ✅ Error recording with stack traces

### 2. Structured Logging with Trace Context

- ✅ JSON formatted logs
- ✅ Automatic trace_id and span_id injection
- ✅ Context-aware logging (InfoContext, ErrorContext, etc.)
- ✅ Correlation between logs and traces in HyperDX

### 3. Metrics Collection

- ✅ Request counters with labels
- ✅ Request duration histograms
- ✅ Prometheus-compatible metrics endpoint

## Prerequisites

- **Docker**: For running HyperDX locally
- **Go 1.24+**: For building the application
- **curl**: For testing endpoints (optional)

## Quick Start

### 1. Start HyperDX

```bash
make hyperdx-up
```

This will:
- Pull the HyperDX local Docker image
- Start HyperDX on ports 4317 (gRPC), 4318 (HTTP), 8080 (UI)
- Wait for HyperDX to be ready

**HyperDX UI**: http://localhost:8080

### 2. Run the Example Application

```bash
make run
```

This will:
- Download dependencies
- Start the HTTP server on http://localhost:8090
- Connect to HyperDX via OTLP gRPC (localhost:4317)

### 3. Generate Sample Traffic

In another terminal:

```bash
make load
```

This will:
- Send 10 requests to `/api/users/:id`
- Send 5 requests to `/api/slow` (simulates slow operations)
- Send 3 requests to `/api/error` (simulates errors)

### 4. View in HyperDX

Open http://localhost:8080 and explore:

- **Traces**: See distributed traces with parent-child spans
- **Logs**: View structured logs with trace correlation
- **Service Map**: Visualize service dependencies (coming soon)

## API Endpoints

### Health Check

```bash
curl http://localhost:8090/health
```

Response:
```json
{
  "status": "healthy",
  "time": "2025-10-03T23:30:00Z"
}
```

### Get User (with tracing)

```bash
curl http://localhost:8090/api/users/123
```

Response:
```json
{
  "id": "123",
  "name": "John Doe",
  "email": "john.doe@example.com"
}
```

**Trace Structure**:
```
GET /api/users/:id (root span)
├── fetchUser (child span)
└── processUser (child span)
```

### Slow Endpoint

```bash
curl http://localhost:8090/api/slow
```

Simulates a 2-second operation. Great for testing performance monitoring.

### Error Endpoint

```bash
curl http://localhost:8090/api/error
```

Simulates an error. The error will be:
- Recorded in the span
- Logged with trace context
- Visible in HyperDX with correlation

### Metrics Endpoint

```bash
curl http://localhost:8090/metrics
```

Returns metrics in Prometheus format (when Prometheus exporter is configured).

## Configuration

Edit `configs/config.yaml` to customize:

```yaml
tracing:
  enabled: true
  service_name: "hyperion-otel-example"
  exporter: "otlp"
  endpoint: "localhost:4317"  # HyperDX gRPC endpoint
  sample_rate: 1.0            # 100% sampling for demo

log:
  level: "debug"
  encoding: "json"

server:
  host: "localhost"
  port: 8090
```

## Observability Features

### Trace Context Injection

Every log includes `trace_id` and `span_id`:

```json
{
  "level": "info",
  "ts": "2025-10-03T23:30:00.123Z",
  "msg": "fetching user",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "user_id": "123"
}
```

### Span Attributes

Spans include rich metadata:

```go
span.SetAttributes(map[string]any{
    "http.method": "GET",
    "http.url": "/api/users/123",
    "http.status_code": 200,
    "user.id": "123",
})
```

### Error Recording

Errors are captured with full context:

```go
err := fmt.Errorf("database connection failed")
span.RecordError(err)
logger.ErrorContext(ctx, "operation failed", "error", err)
```

## HyperDX Features

Once data is flowing, you can:

1. **Search Traces**: Find traces by service, operation, or attributes
2. **View Trace Timeline**: See request flow across spans
3. **Correlate Logs**: Click trace_id to see related logs
4. **Analyze Performance**: Identify slow operations
5. **Debug Errors**: See errors with full context

## Testing the Integration

### Test 1: Trace Context Propagation

```bash
# Generate request
curl http://localhost:8090/api/users/123

# Check logs (should include trace_id and span_id)
# Check HyperDX UI - find the trace
# Click on trace - should see all logs with same trace_id
```

### Test 2: Error Tracking

```bash
# Generate error
curl http://localhost:8090/api/error

# Check HyperDX UI
# - Trace should be marked as error
# - Error details visible in span
# - Error log correlated with trace
```

### Test 3: Performance Monitoring

```bash
# Generate slow request
curl http://localhost:8090/api/slow

# Check HyperDX UI
# - Trace duration ~2 seconds
# - Span breakdown shows where time was spent
```

## Troubleshooting

### HyperDX Not Starting

```bash
# Check Docker logs
make hyperdx-logs

# Ensure ports are not in use
lsof -i :4317 -i :4318 -i :8080 -i :8123

# Restart HyperDX
make hyperdx-down
make hyperdx-up
```

### Application Not Sending Data

Check:
1. HyperDX is running: `docker ps | grep hyperdx`
2. OTLP endpoint is correct: `configs/config.yaml`
3. Application logs for connection errors

### No Traces in HyperDX UI

Wait a few seconds for data to be ingested. Try:

```bash
# Generate more traffic
make load

# Refresh HyperDX UI
# Check time range filter (should include recent data)
```

## Project Structure

```
example/
├── cmd/
│   └── app/
│       └── main.go           # Application entry point
├── configs/
│   └── config.yaml           # Configuration file
├── docker-compose.yml        # HyperDX local setup
├── Makefile                  # Convenient commands
├── go.mod                    # Go module definition
└── README.md                 # This file
```

## Key Code Patterns

### Tracer Middleware

```go
func TracingMiddleware(tracer hyperion.Tracer) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()
        newCtx, span := tracer.Start(ctx, c.Request.Method+" "+c.FullPath())
        defer span.End()

        span.SetAttributes(map[string]any{
            "http.method": c.Request.Method,
            "http.url": c.Request.URL.String(),
        })

        c.Request = c.Request.WithContext(newCtx)
        c.Next()
    }
}
```

### Context-Aware Logging

```go
logger.InfoContext(ctx, "processing request",
    "user_id", userID,
    "operation", "fetchUser",
)
// Automatically includes trace_id and span_id from ctx
```

### Metrics Recording

```go
requestCounter.Add(ctx, 1, map[string]any{
    "endpoint": "/api/users/:id",
    "method": "GET",
})

requestDuration.Record(ctx, duration, map[string]any{
    "endpoint": "/api/users/:id",
})
```

## Cleanup

Stop everything:

```bash
make clean
```

This will:
- Stop HyperDX
- Remove Docker volumes
- Clean up generated files

## Next Steps

1. **Add More Endpoints**: Extend the example with your use cases
2. **Custom Metrics**: Add business-specific metrics
3. **Sampling**: Experiment with different sample rates
4. **Multiple Services**: Run multiple instances to see service mesh
5. **Production Setup**: Deploy HyperDX to production

## Resources

- [Hyperion Documentation](https://github.com/mapoio/hyperion)
- [HyperDX Documentation](https://www.hyperdx.io/docs)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Zap Logger](https://github.com/uber-go/zap)

## License

This example is part of the Hyperion framework and follows the same license.
