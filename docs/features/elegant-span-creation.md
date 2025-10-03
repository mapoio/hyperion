# Elegant Span Creation with `hyperion.StartSpan()`

## Overview

The `hyperion.StartSpan()` helper function provides an elegant solution for creating spans with automatic context and logger management, eliminating repetitive boilerplate code.

## The Problem

When creating spans with distributed tracing, the traditional approach requires manual context and logger updates:

```go
// ❌ Traditional approach - repetitive and error-prone
ctx, span := tracer.Start(ctx, "fetchUser",
    hyperion.WithAttributes(hyperion.String("user.id", userID)),
)
defer span.End()

// Must manually update logger with new context
var contextLogger hyperion.Logger
if ctxAware, ok := logger.(hyperion.ContextAwareLogger); ok {
    contextLogger = ctxAware.WithContext(ctx)
} else {
    contextLogger = logger
}

contextLogger.Info("fetching user", "user_id", userID)

// Creating child span requires repeating the same pattern
ctx, childSpan := tracer.Start(ctx, "processUser")
defer childSpan.End()

// Again, manually update logger
if ctxAware, ok := logger.(hyperion.ContextAwareLogger); ok {
    contextLogger = ctxAware.WithContext(ctx)
} else {
    contextLogger = logger
}

contextLogger.Info("processing user data", "user_id", userID)
```

This pattern has several issues:
- **Repetitive**: The type assertion and logger update code is duplicated everywhere
- **Error-prone**: Easy to forget updating the logger, causing logs to miss trace context
- **Verbose**: Simple span creation requires many lines of boilerplate code

## The Solution

`hyperion.StartSpan()` combines span creation and logger update in a single call:

```go
// ✅ Elegant approach - one line does it all
ctx, span, logger := hyperion.StartSpan(ctx, tracer, logger, "fetchUser",
    hyperion.WithAttributes(hyperion.String("user.id", userID)),
)
defer span.End()

logger.Info("fetching user", "user_id", userID)

// Creating child span is equally simple
ctx, childSpan, logger := hyperion.StartSpan(ctx, tracer, logger, "processUser")
defer childSpan.End()

logger.Info("processing user data", "user_id", userID)
```

## Function Signature

```go
func StartSpan(
    ctx context.Context,
    tracer Tracer,
    logger Logger,
    spanName string,
    opts ...SpanOption,
) (context.Context, Span, Logger)
```

**Parameters:**
- `ctx`: Current context (may contain parent span)
- `tracer`: Tracer instance for creating the span
- `logger`: Logger instance to be updated with new context
- `spanName`: Name of the span to create
- `opts`: Optional span configuration (attributes, span kind, etc.)

**Returns:**
- `context.Context`: Updated context containing the new span
- `Span`: The created span instance
- `Logger`: Updated logger with new trace context (or original logger if it doesn't implement `ContextAwareLogger`)

## How It Works

1. **Creates Span**: Calls `tracer.Start()` to create a new span with the updated context
2. **Updates Logger**: If the logger implements `ContextAwareLogger`, calls `WithContext()` to bind the new context
3. **Returns Everything**: Returns the updated context, span, and logger in one go

## Benefits

### 1. Less Boilerplate
Reduces span creation from ~7 lines to 1 line of code.

### 2. Type-Safe
Compiler ensures all return values are handled correctly.

### 3. Automatic Context Propagation
No risk of forgetting to use the updated context.

### 4. Automatic Trace Correlation
Logs automatically include `trace_id` and `span_id` from the new span.

### 5. Compatible
Works with any logger - if it doesn't implement `ContextAwareLogger`, the original logger is returned unchanged.

## Usage Examples

### Basic Usage

```go
ctx, span, logger := hyperion.StartSpan(ctx, tracer, logger, "operationName")
defer span.End()

logger.Info("operation started")
```

### With Attributes

```go
ctx, span, logger := hyperion.StartSpan(ctx, tracer, logger, "database.query",
    hyperion.WithAttributes(
        hyperion.String("db.table", "users"),
        hyperion.Int64("user.id", userID),
    ),
)
defer span.End()

logger.Info("querying database")
```

### Nested Spans (Parent-Child Relationship)

```go
// Parent span
ctx, parentSpan, logger := hyperion.StartSpan(ctx, tracer, logger, "handleRequest")
defer parentSpan.End()

logger.Info("request received")

// Child span - automatically uses parent context
ctx, childSpan, logger := hyperion.StartSpan(ctx, tracer, logger, "validateInput")
defer childSpan.End()

logger.Info("validating input")

// Another child span
ctx, childSpan2, logger := hyperion.StartSpan(ctx, tracer, logger, "processData")
defer childSpan2.End()

logger.Info("processing data")
```

### With Different Span Kinds

```go
// Client span for outgoing HTTP request
ctx, span, logger := hyperion.StartSpan(ctx, tracer, logger, "external.api.call",
    hyperion.WithSpanKind(hyperion.SpanKindClient),
    hyperion.WithAttributes(
        hyperion.String("http.url", apiURL),
        hyperion.String("http.method", "POST"),
    ),
)
defer span.End()

logger.Info("calling external API")
```

## Complete Example

Here's a real-world HTTP handler using `hyperion.StartSpan()`:

```go
func GetUserHandler(tracer hyperion.Tracer, logger hyperion.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()
        userID := c.Param("id")

        // Create span for the entire operation
        ctx, span, logger := hyperion.StartSpan(ctx, tracer, logger, "fetchUser",
            hyperion.WithAttributes(hyperion.String("user.id", userID)),
        )
        defer span.End()

        logger.Info("fetching user", "user_id", userID)

        // Database query sub-operation
        ctx, dbSpan, logger := hyperion.StartSpan(ctx, tracer, logger, "database.query")
        defer dbSpan.End()

        user, err := queryDatabase(ctx, userID)
        if err != nil {
            logger.Error("database query failed", "error", err)
            dbSpan.RecordError(err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
            return
        }

        dbSpan.AddEvent("user fetched")
        logger.Info("user found", "user_id", userID)

        // Data processing sub-operation
        ctx, processSpan, logger := hyperion.StartSpan(ctx, tracer, logger, "processUser")
        defer processSpan.End()

        processedUser := processUserData(user)
        logger.Info("user data processed")

        c.JSON(http.StatusOK, processedUser)
    }
}
```

## Trace Context in Logs

When using `hyperion.StartSpan()` with a `ContextAwareLogger` implementation (like `adapter/zap`), logs automatically include trace context:

```json
{
  "level": "info",
  "ts": "2025-10-04T01:25:25.302+0800",
  "msg": "fetching user",
  "trace_id": "0eb9495b8c3e9700b7960f15e4b4dfde",
  "span_id": "96d727ebe66a4441",
  "user_id": "700"
}
```

Notice how each operation has a **different `span_id`**, enabling precise tracing of individual operations while maintaining the same `trace_id` for correlation.

## Implementation

The implementation is simple and efficient:

```go
func StartSpan(ctx context.Context, tracer Tracer, logger Logger, spanName string, opts ...SpanOption) (context.Context, Span, Logger) {
    // Create new span
    newCtx, span := tracer.Start(ctx, spanName, opts...)

    // Update logger with new context if it supports it
    contextLogger := logger
    if ctxAware, ok := logger.(ContextAwareLogger); ok {
        contextLogger = ctxAware.WithContext(newCtx)
    }

    return newCtx, span, contextLogger
}
```

## Best Practices

1. **Always use returned context**: Don't reuse old context variable after calling `StartSpan()`
2. **Always defer span.End()**: Ensure spans are properly closed
3. **Use meaningful span names**: Follow OpenTelemetry semantic conventions
4. **Add relevant attributes**: Help with debugging and analysis
5. **Shadow variables**: Use `:=` to shadow `ctx`, `span`, and `logger` variables for cleaner code

## Migration Guide

### Before (without `StartSpan`)

```go
ctx, span := tracer.Start(ctx, "operation")
defer span.End()

var logger hyperion.Logger = baseLogger
if ctxAware, ok := baseLogger.(hyperion.ContextAwareLogger); ok {
    logger = ctxAware.WithContext(ctx)
}

logger.Info("operation started")
```

### After (with `StartSpan`)

```go
ctx, span, logger := hyperion.StartSpan(ctx, tracer, logger, "operation")
defer span.End()

logger.Info("operation started")
```

Simply replace the manual span creation and logger update with a single `hyperion.StartSpan()` call.

## See Also

- [Tracer Interface Documentation](../api/tracer.md)
- [Logger Interface Documentation](../api/logger.md)
- [ContextAwareLogger Interface](../api/context-aware-logger.md)
- [OpenTelemetry Integration Example](../../example/otel/README.md)
