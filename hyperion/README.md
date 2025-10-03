# Hyperion Core

The ultra-lightweight core library of Hyperion framework, defining pure Go interfaces with **zero third-party dependencies** (except `go.uber.org/fx`).

## Overview

Hyperion Core provides the foundational interfaces for building production-ready Go applications with clean architecture principles. All core components ship with NoOp implementations, ensuring zero overhead when specific features aren't needed.

## Core Interfaces

### Observability

- **[Logger](logger.go)**: Structured logging interface
  - Methods: `Debug`, `Info`, `Warn`, `Error`, `Fatal`
  - Context enrichment: `With()`, `WithError()`
  - Default: NoOp logger (zero overhead)

- **[Tracer](tracer.go)**: Distributed tracing interface (OpenTelemetry compatible)
  - Span creation and management
  - Automatic error recording
  - Default: NoOp tracer (zero overhead)

- **[Meter](metric.go)**: Metrics collection interface (OpenTelemetry compatible)
  - Counter, Histogram, Gauge, UpDownCounter
  - Automatic trace exemplars (when using OTel adapter)
  - Default: NoOp meter (zero overhead)

### Infrastructure

- **[Config](config.go)**: Configuration management interface
  - Type-safe getters: `GetString()`, `GetInt()`, `GetBool()`, etc.
  - Hot reload support via `ConfigWatcher`
  - Default: NoOp config (returns zero values)

- **[Database](database.go)**: Database access interface
  - Query execution via `Executor`
  - Transaction management via `UnitOfWork`
  - Default: NoOp database (returns errors)

- **[Cache](cache.go)**: Caching interface
  - Get, Set, Delete operations
  - Batch operations: `MGet`, `MSet`
  - Default: NoOp cache (returns errors)

### Context & Composition

- **[Context](context.go)**: Type-safe request context
  - Embeds `context.Context`
  - Provides: `Logger()`, `Tracer()`, `Meter()`, `DB()`
  - Supports timeout, cancellation, deadline
  - Interceptor integration via `UseIntercept()`

- **[Interceptor](interceptor.go)**: Cross-cutting concerns pattern
  - Method-level AOP without code generation
  - Built-in: `TracingInterceptor`, `LoggingInterceptor`
  - Custom interceptors via fx groups
  - LIFO cleanup order (predictable)

## Key Design Patterns

### 1. The 3-Line Interceptor Pattern

Apply cross-cutting concerns with minimal boilerplate:

```go
func (s *Service) GetUser(ctx hyperion.Context, userID string) (err error) {
    ctx, end := ctx.UseIntercept("Service", "GetUser")
    defer end(&err)

    // Business logic here
    return s.repository.FindByID(ctx, userID)
}
```

**Benefits**:
- Automatic tracing spans
- Automatic logging (entry/exit)
- Automatic metrics recording
- Automatic error handling

### 2. Unified Observability via Context

Logs, Traces, and Metrics automatically correlate through shared context:

```go
func (s *Service) ProcessOrder(ctx hyperion.Context, orderID string) error {
    // Log automatically includes trace_id and span_id
    ctx.Logger().Info("processing order", "order_id", orderID)

    // Metric automatically includes exemplar linking to trace
    counter := ctx.Meter().Counter("orders.processed")
    counter.Add(ctx, 1, hyperion.String("status", "started"))

    // Child span automatically inherits parent trace
    newCtx, span := ctx.Tracer().Start(ctx, "validate-order")
    defer span.End()

    // All three pillars correlated automatically!
    return nil
}
```

### 3. NoOp Pattern for Zero Overhead

When features aren't needed, NoOp implementations provide zero overhead:

```go
// No observability adapters = zero overhead
fx.New(
    hyperion.CoreModule,  // Uses NoOp Logger, Tracer, Meter
    myapp.Module,
).Run()

// With observability = full correlation
fx.New(
    hyperion.CoreModule,
    hyperion.TracingInterceptorModule,
    hyperion.LoggingInterceptorModule,

    zap.Module,   // Real logger
    otel.Module,  // Real tracer + meter

    myapp.Module,
).Run()
```

## Module System

Hyperion uses `go.uber.org/fx` for dependency injection and lifecycle management.

### Core Modules

```go
import "github.com/mapoio/hyperion/hyperion"

// CoreModule provides all default NoOp implementations
hyperion.CoreModule

// Interceptor modules (optional)
hyperion.TracingInterceptorModule   // Auto-create spans
hyperion.LoggingInterceptorModule   // Auto-log method calls
hyperion.AllInterceptorsModule      // Both tracing and logging
```

### ContextFactory

The `ContextFactory` creates `hyperion.Context` instances with all dependencies injected:

```go
func NewService(factory hyperion.ContextFactory) *Service {
    return &Service{factory: factory}
}

func (s *Service) HandleRequest(stdCtx context.Context) error {
    // Create hyperion.Context with all dependencies
    ctx := s.factory.New(stdCtx)

    // Now ctx has Logger, Tracer, Meter, DB, Interceptors
    ctx.Logger().Info("handling request")
    return nil
}
```

## Architecture Principles

1. **Zero Dependencies**: Core only depends on `go.uber.org/fx`
2. **Interface-Driven**: Every component is an interface
3. **NoOp by Default**: Zero overhead when features not used
4. **Adapter Pattern**: Swap implementations without code changes
5. **Type Safety**: Compile-time guarantees, no reflection
6. **Production Ready**: Battle-tested patterns and best practices

## Performance Characteristics

### NoOp Implementations

```
Benchmark_NoOpLogger    1000000000    0.3 ns/op    0 B/op
Benchmark_NoOpTracer    1000000000    0.5 ns/op    0 B/op
Benchmark_NoOpMeter     1000000000    0.4 ns/op    0 B/op
```

**Impact**: Essentially zero (compiler inlines empty functions)

### Interceptor Overhead

```
Benchmark_NoInterceptors        100000000    0.5 ns/op
Benchmark_TracingInterceptor     50000000   30.0 ns/op
Benchmark_LoggingInterceptor     30000000   40.0 ns/op
Benchmark_AllInterceptors        20000000   80.0 ns/op
```

**Impact**: Minimal for typical use cases

## Usage Examples

### Example 1: Minimal Setup (No Observability)

```go
package main

import (
    "context"
    "go.uber.org/fx"
    "github.com/mapoio/hyperion/hyperion"
)

func main() {
    fx.New(
        hyperion.CoreModule,  // NoOp implementations
        fx.Invoke(run),
    ).Run()
}

func run(factory hyperion.ContextFactory) {
    ctx := factory.New(context.Background())
    ctx.Logger().Info("hello")  // No-op, zero overhead
}
```

### Example 2: Full Observability

```go
package main

import (
    "context"
    "go.uber.org/fx"
    "github.com/mapoio/hyperion/hyperion"
    "github.com/mapoio/hyperion/adapter/zap"
    "github.com/mapoio/hyperion/adapter/otel"
)

func main() {
    fx.New(
        hyperion.CoreModule,
        hyperion.AllInterceptorsModule,

        zap.Module,   // Real logger
        otel.Module,  // Real tracer + meter

        fx.Provide(NewUserService),
        fx.Invoke(run),
    ).Run()
}

func run(factory hyperion.ContextFactory, service *UserService) {
    ctx := factory.New(context.Background())

    // Logs include trace_id, metrics include exemplars
    service.GetUser(ctx, "user123")
}

type UserService struct{}

func NewUserService() *UserService {
    return &UserService{}
}

func (s *UserService) GetUser(ctx hyperion.Context, id string) (err error) {
    // Automatic tracing, logging, metrics
    ctx, end := ctx.UseIntercept("UserService", "GetUser")
    defer end(&err)

    ctx.Logger().Info("fetching user", "user_id", id)

    counter := ctx.Meter().Counter("user.requests")
    counter.Add(ctx, 1, hyperion.String("method", "GetUser"))

    return nil
}
```

### Example 3: Custom Interceptor

```go
package main

import (
    "time"
    "go.uber.org/fx"
    "github.com/mapoio/hyperion/hyperion"
)

// MetricsInterceptor records method duration
type MetricsInterceptor struct {
    meter hyperion.Meter
}

func NewMetricsInterceptor(meter hyperion.Meter) hyperion.Interceptor {
    return &MetricsInterceptor{meter: meter}
}

func (m *MetricsInterceptor) Name() string {
    return "metrics"
}

func (m *MetricsInterceptor) Intercept(
    ctx hyperion.Context,
    fullPath string,
) (hyperion.Context, func(err *error), error) {
    start := time.Now()

    counter := ctx.Meter().Counter("method.calls")
    counter.Add(ctx, 1, hyperion.String("method", fullPath))

    end := func(errPtr *error) {
        duration := time.Since(start).Milliseconds()
        histogram := ctx.Meter().Histogram("method.duration")
        histogram.Record(ctx, float64(duration), hyperion.String("method", fullPath))

        if errPtr != nil && *errPtr != nil {
            errorCounter := ctx.Meter().Counter("method.errors")
            errorCounter.Add(ctx, 1, hyperion.String("method", fullPath))
        }
    }

    return ctx, end, nil
}

func (m *MetricsInterceptor) Order() int {
    return 300 // After tracing (100) and logging (200)
}

func main() {
    fx.New(
        hyperion.CoreModule,
        hyperion.AllInterceptorsModule,

        // Register custom interceptor
        fx.Provide(
            fx.Annotate(
                NewMetricsInterceptor,
                fx.ResultTags(`group:"hyperion.interceptors"`),
            ),
        ),

        fx.Invoke(run),
    ).Run()
}
```

## Documentation

### Core Guides
- **[Interceptor Pattern](../docs/interceptor.md)**: Complete interceptor usage guide
- **[Observability Architecture](../docs/observability.md)**: Unified Logs, Traces, and Metrics
- **[Quick Start Guide](../QUICK_START.md)**: 5-minute tutorial with complete example

### Design Documents
- **[Interceptor Architecture](../.design/interceptor-architecture.md)**: Deep dive into interceptor design
- **[Observability Architecture](../.design/observability-architecture.md)**: Deep dive into observability correlation

### Adapter Documentation
- **[Viper Adapter](../adapter/viper/README.md)**: Configuration management
- **[Zap Adapter](../adapter/zap/README.md)**: Structured logging
- **[GORM Adapter](../adapter/gorm/README.md)**: Database access and transactions

## Testing

All core components include comprehensive tests:

```bash
# Run all tests
cd hyperion
go test -v ./...

# Run with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Current test coverage**: 90%+ for all core interfaces

## Best Practices

### DO

- ✅ Always pass `hyperion.Context` through the call chain
- ✅ Use `UseIntercept` for service-layer methods
- ✅ Return named error variable for cleanup: `(err error)`
- ✅ Use `defer end(&err)` pattern
- ✅ Pass context to all observability methods

### DON'T

- ❌ Create new `context.Background()` (loses trace context)
- ❌ Store `hyperion.Context` in struct fields
- ❌ Apply interceptors at repository layer (too granular)
- ❌ Ignore interceptor initialization errors
- ❌ Use observability methods without context

## Contributing

See the main [Contributing Guide](../CONTRIBUTING.md) for details on:
- Code standards
- Testing requirements
- Commit message format
- PR process

## License

MIT License - see [LICENSE](../LICENSE) for details

---

**Part of the Hyperion Framework** | [Documentation](../docs/) | [GitHub](https://github.com/mapoio/hyperion)
