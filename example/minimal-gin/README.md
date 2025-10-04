# Minimal Gin Example

The simplest possible Hyperion + Gin integration example.

## What's Included

- ✅ **Gin HTTP Server** with graceful shutdown
- ✅ **Structured Logging** via Zap adapter
- ✅ **Configuration** via Viper adapter
- ✅ **HyperionMiddleware** for context propagation
- ✅ **UseIntercept Pattern** demonstration
- ✅ **Clean Architecture** (just 1 file for simplicity)

## Prerequisites

```bash
go 1.24+
```

## Quick Start

### 1. Install Dependencies

```bash
cd example/minimal-gin
go mod tidy
```

### 2. Run the Server

```bash
go run cmd/app/main.go
```

Expected output:
```
{"level":"info","ts":"2025-10-04T12:00:00.000+0800","caller":"app/main.go:154","msg":"starting HTTP server","address":"localhost:8080"}
```

### 3. Test the Endpoints

**Health Check:**
```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2025-10-04T12:00:00Z"
}
```

**Hello Endpoint:**
```bash
curl "http://localhost:8080/hello?name=Alice"
```

Response:
```json
{
  "message": "Hello, Alice!"
}
```

Server logs:
```
{"level":"info","ts":"...","msg":"handling hello request","name":"Alice"}
```

**Demo Endpoint (UseIntercept pattern):**
```bash
curl "http://localhost:8080/demo?input=test"
```

Response:
```json
{
  "result": "Processed: test"
}
```

Server logs:
```
{"level":"debug","ts":"...","msg":"Method started","path":"DemoService.Process"}
{"level":"info","ts":"...","msg":"processing input","input":"test"}
{"level":"debug","ts":"...","msg":"Method completed","path":"DemoService.Process","duration":"50.2ms"}
```

**Info Endpoint (Config access):**
```bash
curl http://localhost:8080/info
```

Response:
```json
{
  "app": "minimal-gin-example",
  "environment": "development"
}
```

## Code Structure

```
example/minimal-gin/
├── cmd/app/
│   └── main.go           # Single file with everything
├── configs/
│   └── config.yaml       # Configuration file
├── go.mod
└── README.md
```

## Key Patterns Demonstrated

### 1. HyperionMiddleware Pattern

```go
func HyperionMiddleware(factory hyperion.ContextFactory) gin.HandlerFunc {
    return func(c *gin.Context) {
        hctx := factory.New(c.Request.Context())
        c.Set("hctx", hctx)
        c.Next()
    }
}
```

**Why:** Creates `hyperion.Context` once per request and stores it in Gin context for all handlers.

### 2. Context Extraction Helper

```go
func GetHyperionContext(c *gin.Context) hyperion.Context {
    hctx, exists := c.Get("hctx")
    if !exists {
        panic("hyperion context not found")
    }
    return hctx.(hyperion.Context)
}
```

**Why:** Convenient helper to extract hyperion.Context in handlers.

### 3. UseIntercept Pattern in Handlers

```go
func demoServiceCall(hctx hyperion.Context, input string) (result string, err error) {
    hctx, end := hctx.UseIntercept("DemoService", "Process")
    defer end(&err)

    // Business logic...
    return result, nil
}
```

**Why:** 3-line pattern automatically adds tracing, logging, and metrics.

## Configuration

Edit `configs/config.yaml` to customize:

```yaml
app:
  name: my-app
  env: production

server:
  host: 0.0.0.0
  port: 8080

log:
  level: info      # debug, info, warn, error
  encoding: json   # json or console
  output: stdout   # stdout, stderr, or file path
```

## Next Steps

1. **Add Database** - See [GORM adapter example](../otel/) for database integration
2. **Add Tracing** - See [OTel example](../otel/) for distributed tracing
3. **Add Validation** - Use `binding` tags in request structs
4. **Add Layers** - Split into `handler/`, `service/`, `repository/` packages

## Differences from Full Example

| Feature | Minimal Example | Full Example (example/otel) |
|---------|----------------|----------------------------|
| Files | 1 main.go | Multiple packages |
| Adapters | Viper + Zap | Viper + Zap + OTel + GORM |
| Tracing | No-Op (disabled) | OpenTelemetry |
| Metrics | No-Op (disabled) | OpenTelemetry |
| Database | None | PostgreSQL with GORM |
| Architecture | Single file | Clean layers (handler/service/repository) |

**Use minimal example for:**
- Learning Hyperion basics
- Simple HTTP APIs
- Prototyping

**Use full example for:**
- Production applications
- Microservices with full observability
- Complex business logic

## Learn More

- [Hyperion Core README](../../hyperion/README.md)
- [Quick Start Guide](../../QUICK_START.md)
- [Full OTel Example](../otel/README.md)
- [Architecture Guide](../../docs/architecture.md)
