# Interceptor Design

This document describes Hyperion's interceptor pattern for implementing cross-cutting concerns in service layer methods.

## Overview

Interceptors provide a clean, explicit way to add cross-cutting concerns (tracing, logging, metrics, transactions) to service methods without code generation or runtime magic.

### The 3-Line Pattern

```go
func (s *Service) Method(ctx hyperion.Context, ...) (err error) {
    ctx, end := ctx.UseIntercept("Service", "Method")
    defer end(&err)

    // Business logic here
    return nil
}
```

**Why This Pattern?**
1. **Explicit**: Clear where interceptors are applied
2. **Simple**: Only 3 lines of boilerplate
3. **Flexible**: Can be customized per-method
4. **Type-safe**: No reflection or code generation
5. **Composable**: Multiple interceptors work together

## Core Concepts

### 1. Interceptor Interface

```go
type Interceptor interface {
    // Name returns unique identifier for filtering
    Name() string

    // Intercept wraps the method execution
    Intercept(ctx Context, fullPath string) (newCtx Context, end func(err *error), ierr error)

    // Order determines execution order (lower = outer)
    Order() int
}
```

### 2. Execution Order

Interceptors execute in **ascending Order()** value:

```
Request → [Tracing:100] → [Logging:200] → [Custom:300] → Business Logic
                                                           ↓
Response ← [Tracing] ← [Logging] ← [Custom] ← Return Value
```

**LIFO (Last-In-First-Out)** cleanup:
- Tracing starts first, ends last
- Logging starts second, ends second-to-last
- Custom starts last, ends first

### 3. Context Propagation

Each interceptor can modify the context:

```go
func (i *MyInterceptor) Intercept(ctx Context, fullPath string) (Context, func(err *error), error) {
    // Modify context (e.g., add fields to logger)
    newLogger := ctx.Logger().With("interceptor", "my")
    newCtx := hyperion.WithLogger(ctx, newLogger)

    // Return modified context
    end := func(errPtr *error) {
        // Cleanup logic
    }

    return newCtx, end, nil
}
```

## Built-In Interceptors

### TracingInterceptor (Order: 100)

Creates OpenTelemetry spans automatically.

**Module**: `hyperion.TracingInterceptorModule`

**Behavior**:
- Creates span with name `{fullPath}` (e.g., "UserService.GetUser")
- Records errors on span if method returns error
- Ends span in defer (LIFO)

**Implementation**:
```go
func (ti *TracingInterceptor) Intercept(ctx Context, fullPath string) (Context, func(err *error), error) {
    // Start OTel span
    newStdCtx, span := ti.tracer.Start(ctx, fullPath)

    // Create new hyperion.Context with updated context.Context
    newCtx := /* preserve all fields, update Context */

    end := func(errPtr *error) {
        if errPtr != nil && *errPtr != nil {
            span.RecordError(*errPtr)
        }
        span.End()
    }

    return newCtx, end, nil
}
```

**Usage**:
```go
fx.New(
    hyperion.CoreModule,
    hyperion.TracingInterceptorModule,  // Enable tracing
    otel.Module,                        // Provide OTel Tracer
)
```

### LoggingInterceptor (Order: 200)

Logs method entry/exit with duration.

**Module**: `hyperion.LoggingInterceptorModule`

**Behavior**:
- Logs "Method started" at DEBUG level
- Logs "Method completed" or "Method failed" with duration
- Includes error in log if method fails

**Implementation**:
```go
func (li *LoggingInterceptor) Intercept(ctx Context, fullPath string) (Context, func(err *error), error) {
    start := time.Now()
    li.logger.Debug("Method started", "path", fullPath)

    end := func(errPtr *error) {
        duration := time.Since(start)
        if errPtr != nil && *errPtr != nil {
            li.logger.Error("Method failed", "path", fullPath, "duration", duration, "error", *errPtr)
        } else {
            li.logger.Debug("Method completed", "path", fullPath, "duration", duration)
        }
    }

    return ctx, end, nil
}
```

**Usage**:
```go
fx.New(
    hyperion.CoreModule,
    hyperion.LoggingInterceptorModule,  // Enable logging
    zap.Module,                         // Provide Logger
)
```

### AllInterceptorsModule

Convenience module that enables all built-in interceptors.

```go
fx.New(
    hyperion.CoreModule,
    hyperion.AllInterceptorsModule,  // Tracing + Logging
)
```

## Custom Interceptors

### Example: Metrics Interceptor

```go
type MetricsInterceptor struct {
    meter hyperion.Meter
}

func NewMetricsInterceptor(meter hyperion.Meter) hyperion.Interceptor {
    return &MetricsInterceptor{meter: meter}
}

func (m *MetricsInterceptor) Name() string {
    return "metrics"
}

func (m *MetricsInterceptor) Order() int {
    return 300  // After tracing and logging
}

func (m *MetricsInterceptor) Intercept(
    ctx hyperion.Context,
    fullPath string,
) (hyperion.Context, func(err *error), error) {
    start := time.Now()

    // Create counter and histogram
    requestCounter := m.meter.Counter("method.calls",
        hyperion.WithMetricDescription("Method call count"),
    )
    latencyHistogram := m.meter.Histogram("method.latency",
        hyperion.WithMetricDescription("Method latency"),
        hyperion.WithMetricUnit("ms"),
    )

    end := func(errPtr *error) {
        duration := time.Since(start)
        status := "success"
        if errPtr != nil && *errPtr != nil {
            status = "error"
        }

        // Record metrics
        requestCounter.Add(ctx, 1,
            hyperion.String("method", fullPath),
            hyperion.String("status", status),
        )
        latencyHistogram.Record(ctx, float64(duration.Milliseconds()),
            hyperion.String("method", fullPath),
        )
    }

    return ctx, end, nil
}
```

**Registration**:
```go
fx.Provide(
    fx.Annotate(
        NewMetricsInterceptor,
        fx.ResultTags(`group:"hyperion.interceptors"`),
    ),
)
```

### Example: Transaction Interceptor

```go
type TransactionInterceptor struct {
    uow hyperion.UnitOfWork
}

func (t *TransactionInterceptor) Name() string {
    return "transaction"
}

func (t *TransactionInterceptor) Order() int {
    return 400  // After all observability
}

func (t *TransactionInterceptor) Intercept(
    ctx hyperion.Context,
    fullPath string,
) (hyperion.Context, func(err *error), error) {
    var txCtx hyperion.Context
    var committed bool

    // Start transaction
    err := t.uow.WithTransaction(ctx, func(tc hyperion.Context) error {
        txCtx = tc
        return nil  // Don't commit yet
    })
    if err != nil {
        return ctx, func(*error) {}, err
    }

    end := func(errPtr *error) {
        if errPtr != nil && *errPtr != nil {
            // Rollback on error
            // (handled by UnitOfWork)
        } else {
            // Commit on success
            committed = true
        }
    }

    return txCtx, end, nil
}
```

## Selective Application

### WithOnly - Apply Specific Interceptors

```go
func (s *Service) GetUser(ctx hyperion.Context, id string) (err error) {
    // Only apply tracing, skip logging
    ctx, end := ctx.UseIntercept("Service", "GetUser",
        hyperion.WithOnly("tracing"))
    defer end(&err)

    return nil
}
```

### WithExclude - Skip Specific Interceptors

```go
func (s *Service) HealthCheck(ctx hyperion.Context) (err error) {
    // Skip logging for high-frequency health checks
    ctx, end := ctx.UseIntercept("Service", "HealthCheck",
        hyperion.WithExclude("logging"))
    defer end(&err)

    return nil
}
```

### WithAdditional - Add Method-Specific Interceptors

```go
func (s *Service) CriticalOperation(ctx hyperion.Context) (err error) {
    // Add custom rate limiter for this method only
    ctx, end := ctx.UseIntercept("Service", "CriticalOperation",
        hyperion.WithAdditional(rateLimiterInterceptor))
    defer end(&err)

    return nil
}
```

## Path Naming

### Simple Path

```go
ctx, end := ctx.UseIntercept("UserService", "GetUser")
// Path: "UserService.GetUser"
```

### Namespaced Path

```go
ctx, end := ctx.UseIntercept("Service", "User", "GetUser")
// Path: "Service.User.GetUser"
```

### Dynamic Path

```go
serviceName := "UserService"
methodName := "GetUser"
ctx, end := ctx.UseIntercept(serviceName, methodName)
// Path: "UserService.GetUser"
```

## Configuration Patterns

### Pattern 1: Global Default

All methods use all registered interceptors:

```go
func (s *Service) Method1(ctx hyperion.Context) (err error) {
    ctx, end := ctx.UseIntercept("Service", "Method1")
    defer end(&err)
    return nil
}

func (s *Service) Method2(ctx hyperion.Context) (err error) {
    ctx, end := ctx.UseIntercept("Service", "Method2")
    defer end(&err)
    return nil
}
```

### Pattern 2: Per-Method Customization

```go
func (s *Service) PublicAPI(ctx hyperion.Context) (err error) {
    // Full observability
    ctx, end := ctx.UseIntercept("Service", "PublicAPI")
    defer end(&err)
    return nil
}

func (s *Service) InternalHelper(ctx hyperion.Context) (err error) {
    // No logging (too verbose)
    ctx, end := ctx.UseIntercept("Service", "InternalHelper",
        hyperion.WithExclude("logging"))
    defer end(&err)
    return nil
}
```

### Pattern 3: Service-Level Wrapper

```go
type Service struct {
    config InterceptConfig
}

func (s *Service) intercept(ctx hyperion.Context, method string) (hyperion.Context, func(err *error)) {
    opts := s.config.ToOptions()  // Convert config to options
    return ctx.UseIntercept("Service", method, opts...)
}

func (s *Service) GetUser(ctx hyperion.Context, id string) (err error) {
    ctx, end := s.intercept(ctx, "GetUser")
    defer end(&err)
    return nil
}
```

## Error Handling

### Interceptor Errors

If an interceptor returns an error, it is skipped and logged:

```go
func (i *MyInterceptor) Intercept(ctx Context, fullPath string) (Context, func(err *error), error) {
    if !i.initialized {
        return ctx, func(*error) {}, errors.New("not initialized")
    }
    // ...
}
```

**Behavior**:
- Error logged: `[ERROR] Interceptor error: interceptor=my path=Service.GetUser error=not initialized`
- Other interceptors still execute
- Method proceeds normally

### Method Errors

Errors returned by methods are passed to all `end` functions:

```go
func (s *Service) GetUser(ctx hyperion.Context, id string) (err error) {
    ctx, end := ctx.UseIntercept("Service", "GetUser")
    defer end(&err)  // err will be passed to all end functions

    if id == "" {
        return errors.New("empty ID")  // Passed to all end functions
    }

    return nil
}
```

**Interceptor can observe error**:
```go
end := func(errPtr *error) {
    if errPtr != nil && *errPtr != nil {
        // Error occurred: *errPtr contains the error
        span.RecordError(*errPtr)
        logger.Error("method failed", "error", *errPtr)
    }
}
```

## Performance

### No Interceptors (Default)

```go
// Zero overhead
ctx, end := ctx.UseIntercept("Service", "GetUser")
defer end(&err)
// end is a no-op function
```

**Cost**: ~2 ns (function call overhead only)

### With Interceptors

```go
// Overhead depends on number and type of interceptors
ctx, end := ctx.UseIntercept("Service", "GetUser")
defer end(&err)
```

**Typical costs**:
- Tracing: ~500 ns (span creation)
- Logging: ~200 ns (structured log entry)
- Metrics: ~100 ns (counter increment)

**Total**: ~800 ns for all built-in interceptors

### Optimization: Selective Application

For hot paths, exclude expensive interceptors:

```go
func (s *Service) HighFrequencyMethod(ctx hyperion.Context) (err error) {
    // Only tracing, skip logging
    ctx, end := ctx.UseIntercept("Service", "HighFrequencyMethod",
        hyperion.WithOnly("tracing"))
    defer end(&err)

    return nil
}
```

## Best Practices

### 1. Use Interceptors for Service Layer Only

```go
// ✅ Good: Service layer methods
type UserService struct{}

func (s *UserService) GetUser(ctx hyperion.Context, id string) (err error) {
    ctx, end := ctx.UseIntercept("UserService", "GetUser")
    defer end(&err)
    return nil
}

// ❌ Bad: Don't use in repository layer (too granular)
type UserRepository struct{}

func (r *UserRepository) query(ctx hyperion.Context, sql string) (err error) {
    ctx, end := ctx.UseIntercept("UserRepository", "query")  // Too verbose
    defer end(&err)
    return nil
}
```

### 2. Consistent Naming

```go
// ✅ Good: Consistent service.method pattern
ctx, end := ctx.UseIntercept("UserService", "GetUser")
ctx, end := ctx.UseIntercept("OrderService", "CreateOrder")
ctx, end := ctx.UseIntercept("PaymentService", "ProcessPayment")

// ❌ Bad: Inconsistent naming
ctx, end := ctx.UseIntercept("user-service", "get_user")
ctx, end := ctx.UseIntercept("order", "create-order")
ctx, end := ctx.UseIntercept("payments", "process")
```

### 3. Always Defer end()

```go
// ✅ Good: Defer guarantees cleanup
func (s *Service) GetUser(ctx hyperion.Context) (err error) {
    ctx, end := ctx.UseIntercept("Service", "GetUser")
    defer end(&err)
    return processUser()
}

// ❌ Bad: Manual cleanup can be skipped on early return
func (s *Service) GetUser(ctx hyperion.Context) (err error) {
    ctx, end := ctx.UseIntercept("Service", "GetUser")

    if !valid {
        return errors.New("invalid")  // end() never called!
    }

    end(&err)
    return nil
}
```

### 4. Use Named Return Value for Error

```go
// ✅ Good: Named return enables defer to capture error
func (s *Service) GetUser(ctx hyperion.Context, id string) (err error) {
    ctx, end := ctx.UseIntercept("Service", "GetUser")
    defer end(&err)  // err captured
    return errors.New("failed")
}

// ❌ Bad: Cannot capture error with unnamed return
func (s *Service) GetUser(ctx hyperion.Context, id string) error {
    ctx, end := ctx.UseIntercept("Service", "GetUser")
    defer end(???)  // What to pass?
    return errors.New("failed")
}
```

## Comparison with Other Patterns

### vs. Middleware (HTTP Layer)

**Middleware**: Request/Response level
```go
// HTTP middleware - wraps entire request
router.Use(LoggingMiddleware)
router.Use(TracingMiddleware)
```

**Interceptor**: Method level
```go
// Interceptor - wraps specific service methods
func (s *Service) GetUser(ctx hyperion.Context, id string) (err error) {
    ctx, end := ctx.UseIntercept("Service", "GetUser")
    defer end(&err)
    return nil
}
```

**When to use**:
- Middleware: Authentication, rate limiting, CORS
- Interceptor: Tracing, logging, metrics for business logic

### vs. Decorator Pattern (Removed)

**Decorator**: Compile-time wrapper
```go
// Old approach - decorator wraps logger
logger = LoggingDecorator(logger)
```

**Interceptor**: Runtime control
```go
// New approach - explicit interception
ctx, end := ctx.UseIntercept("Service", "GetUser")
defer end(&err)
```

**Why interceptor is better**:
- ✅ Explicit: Clear where applied
- ✅ Flexible: Can customize per-method
- ✅ Composable: Multiple interceptors work together
- ✅ No code generation

### vs. AOP/AspectJ

**AspectJ**: Magic annotations
```java
@Trace
@Log
public void getUser(String id) {
    // Business logic
}
```

**Interceptor**: Explicit pattern
```go
func (s *Service) GetUser(ctx hyperion.Context, id string) (err error) {
    ctx, end := ctx.UseIntercept("Service", "GetUser")
    defer end(&err)
    // Business logic
    return nil
}
```

**Why explicit is better**:
- ✅ No magic: Easy to understand
- ✅ Type-safe: Compile-time checks
- ✅ Debuggable: Step through in debugger
- ✅ Flexible: Can customize per-call

## Troubleshooting

### Interceptor Not Applied

**Symptom**: No logs/traces despite UseIntercept

**Solution**: Check if interceptor module is enabled
```go
fx.New(
    hyperion.CoreModule,
    hyperion.TracingInterceptorModule,  // Must enable!
    hyperion.LoggingInterceptorModule,  // Must enable!
)
```

### Wrong Execution Order

**Symptom**: Interceptors execute in wrong order

**Solution**: Check Order() values (lower = outer)
```go
func (i *MyInterceptor) Order() int {
    return 150  // Between tracing (100) and logging (200)
}
```

### Context Not Propagated

**Symptom**: Nested calls don't see parent span

**Solution**: Use returned context
```go
// ✅ Correct: Use returned ctx
ctx, end := ctx.UseIntercept("Service", "GetUser")
defer end(&err)
return s.repository.GetUser(ctx, id)  // Pass ctx!

// ❌ Wrong: Use original ctx
ctx, end := ctx.UseIntercept("Service", "GetUser")
defer end(&err)
return s.repository.GetUser(originalCtx, id)  // Wrong!
```

## References

- [Observability Design](./OBSERVABILITY.md)
- [Quick Start Guide](../QUICK_START.md)
- [OpenTelemetry Specification](https://opentelemetry.io/docs/specs/)
