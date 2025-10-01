# Hyperion Architecture Decision Records (ADR)

**Version**: 1.0
**Date**: January 2025

This document records key architectural decisions made during the design of the Hyperion framework and the rationale behind them.

---

## Table of Contents

1. [ADR-001: Choosing go.uber.org/fx as the Dependency Injection Framework](#adr-001-choosing-gouberorgfx-as-the-dependency-injection-framework)
2. [ADR-002: hyperctx.Context Design - Type-Safe Context](#adr-002-hyperctxcontext-design---type-safe-context)
3. [ADR-003: Full OpenTelemetry Integration into Context](#adr-003-full-opentelemetry-integration-into-context)
4. [ADR-004: Error Handling - Typed Error Code Design](#adr-004-error-handling---typed-error-code-design)
5. [ADR-005: UnitOfWork Pattern for Transaction Management](#adr-005-unitofwork-pattern-for-transaction-management)
6. [ADR-006: Configuration Hot Reload Support](#adr-006-configuration-hot-reload-support)
7. [ADR-007: Dynamic Log Level Adjustment](#adr-007-dynamic-log-level-adjustment)
8. [ADR-008: Modular Architecture - Microkernel Design](#adr-008-modular-architecture---microkernel-design)

---

## ADR-001: Choosing go.uber.org/fx as the Dependency Injection Framework

### Status
✅ **Accepted**

### Context
We need a mature dependency injection framework to manage component lifecycles and dependencies.

### Decision
Choose `go.uber.org/fx` as the core dependency injection framework.

### Rationale

**Advantages:**
1. **Production-Grade Maturity**: Extensively used at Uber, battle-tested at scale
2. **Explicit Dependency Declaration**: Dependencies declared through function signatures, clear and traceable
3. **Lifecycle Management**: Built-in `OnStart`/`OnStop` hooks for graceful startup and shutdown
4. **Modular Support**: Native support for component modularity via `fx.Module`
5. **Error Handling**: Dependency resolution failures reported immediately at startup
6. **Visualization**: Support for dependency graph visualization (`fx.Visualize`)

**Alternatives Considered:**
- **Wire (Google)**: Compile-time DI but lacks lifecycle management
- **Dig (Uber)**: fx is built on Dig, but fx provides higher-level abstractions
- **Manual DI**: Flexible but high maintenance cost, lacks standardization

### Consequences

**Positive:**
- Loose coupling between components, easy to test
- Standardized module definition approach
- Automatic lifecycle management

**Negative:**
- Learning curve (for developers unfamiliar with DI)
- Slightly longer startup time (dependency graph construction)

### References
- [fx Documentation](https://uber-go.github.io/fx/)

---

## ADR-002: hyperctx.Context Design - Type-Safe Context

### Status
✅ **Accepted**

### Context
Go's standard `context.Context` uses `Value()` for passing data, which is not type-safe, error-prone, and hard to maintain.

### Decision
Design `hyperctx.Context` interface that embeds `context.Context` and provides type-safe dependency access.

### Core Design

```go
type Context interface {
    context.Context

    Logger() hyperlog.Logger  // Type-safe
    DB() *gorm.DB             // Type-safe
    User() User               // Type-safe
    Span() trace.Span         // Type-safe

    // Tracing methods
    StartSpan(layer, component, operation string) (Context, func())
    RecordError(err error)
    SetAttributes(attrs ...attribute.KeyValue)

    // Context management
    WithTimeout(timeout time.Duration) (Context, context.CancelFunc)
    WithCancel() (Context, context.CancelFunc)
}
```

### Rationale

**Advantages:**
1. **Type Safety**: Compile-time checking, avoids type assertion errors
2. **IDE-Friendly**: Auto-completion, clear interface definitions
3. **Integrated Tracing**: OpenTelemetry operations integrated into Context
4. **Simplified API**: `ctx.Logger()` is cleaner than `ctx.Value("logger").(Logger)`
5. **Immutability**: `WithXxx()` returns new instances, avoiding concurrency issues

**Alternatives Considered:**
- **Standard context.Context + Value()**: Flexible but not type-safe
- **Global Variables**: Violates dependency injection principles, hard to test
- **Multiple Function Parameters**: Bloated function signatures

### Consequences

**Positive:**
- Improved code readability
- Errors caught at compile time
- Unified context management approach

**Negative:**
- Requires type conversion `ctx.(hyperctx.Context)` (one-time, at entry points)
- Not fully compatible with standard library (mitigated by embedding `context.Context`)

---

## ADR-003: Full OpenTelemetry Integration into Context

### Status
✅ **Accepted**

### Context
Distributed tracing requires creating spans in every function, and the traditional approach requires introducing additional tracer objects.

### Decision
Directly integrate OpenTelemetry tracing operations into the `hyperctx.Context` interface.

### Core Methods

```go
StartSpan(layer, component, operation string) (Context, func())
RecordError(err error)
SetAttributes(attrs ...attribute.KeyValue)
AddEvent(name string, attrs ...attribute.KeyValue)
```

### Rationale

**Advantages:**
1. **Zero Intrusion**: No need for separate `tracer` objects
2. **Automatic Propagation**: Spans automatically passed in Context
3. **Concise API**: One line of code to create a span
4. **Naming Convention**: Enforces `{layer}.{component}.{operation}` format
5. **Error Association**: `RecordError()` automatically associates with current span

**Usage Comparison:**

```go
// Traditional approach
tracer := otel.Tracer("app")
ctx, span := tracer.Start(ctx, "service.UserService.GetByID")
defer span.End()
span.SetAttributes(attribute.String("user_id", id))
if err != nil {
    span.RecordError(err)
}

// Hyperion approach
ctx, end := ctx.StartSpan("service", "UserService", "GetByID")
defer end()
ctx.SetAttributes(attribute.String("user_id", id))
if err != nil {
    ctx.RecordError(err)
}
```

### Consequences

**Positive:**
- More concise code
- Enforces standardized span naming
- Reduces cognitive load for developers

**Negative:**
- Need to learn Hyperion-specific API (but simpler)

---

## ADR-004: Error Handling - Typed Error Code Design

### Status
✅ **Accepted**

### Context
We need a unified error handling mechanism that supports HTTP/gRPC status code mapping and can carry business context.

### Decision
Design a typed `Code` struct and `Error` type with multi-layer wrapping support.

### Core Design

```go
type Code struct {
    Code       string     // "USER_NOT_FOUND"
    HTTPStatus int        // 404
    GRPCCode   codes.Code // codes.NotFound
}

var CodeNotFound = Code{"NOT_FOUND", 404, codes.NotFound}

type Error struct {
    code    Code
    message string
    cause   error          // Support wrapping
    fields  map[string]any // Business context
}

// Convenient constructors
func NotFound(message string) *Error
func Wrap(code Code, message string, err error) *Error
```

### Rationale

**Advantages:**
1. **Type Safety**: `Code` is a typed constant, not a string
2. **Automatic Mapping**: One Code contains both HTTP and gRPC status codes
3. **Multi-Layer Wrapping**: Supports standard `errors.Unwrap()` interface
4. **Context Fields**: `WithField()` adds business information
5. **Reduced Boilerplate**: Predefined common error constructors

**Comparison with Other Approaches:**

| Approach | Pros | Cons |
|----------|------|------|
| Standard `error` | Simple | No status code, no context |
| `pkg/errors` | Stack support | No status code mapping |
| String error codes | Flexible | Not type-safe |
| **Hyperion Code** | Type-safe + status + wrapping | Need to learn new API |

### Usage Examples

```go
// Simple error
return hypererror.NotFound("user not found")

// With fields
return hypererror.ResourceNotFound("user", userID)

// Wrap error
return hypererror.InternalWrap("query failed", err)

// Multi-layer wrapping
return hypererror.Wrap(
    hypererror.CodeInternal,
    "failed to create user",
    err,
).WithField("email", email)
```

### Consequences

**Positive:**
- Unified error handling approach
- Automatic HTTP/gRPC response conversion
- Rich error context

**Negative:**
- Need to learn Hyperion error API

---

## ADR-005: UnitOfWork Pattern for Transaction Management

### Status
✅ **Accepted**

### Context
Database transaction boundaries should be managed at the Service layer, but the Repository layer should not be aware of transactions.

### Decision
Implement `UnitOfWork` interface to declaratively manage transactions via `WithTransaction()`.

### Core Design

```go
type UnitOfWork interface {
    WithTransaction(ctx hyperctx.Context, fn func(txCtx hyperctx.Context) error) error
}

// Automatic transaction propagation
func (ctx *hyperContext) DB() *gorm.DB {
    return ctx.db // Automatically returns tx or pool
}
```

### Rationale

**Advantages:**
1. **Declarative**: Service layer clearly declares transaction boundaries
2. **Automatic Propagation**: `ctx.DB()` automatically recognizes transaction context
3. **Repository Transparent**: Repository doesn't need to care if it's in a transaction
4. **Automatic Rollback**: Function returning error automatically rolls back
5. **Nested Calls**: Services can call each other, transactions propagate naturally

**Usage Comparison:**

```go
// Traditional approach
tx := db.Begin()
if err := repo.Create(tx, user); err != nil {
    tx.Rollback()
    return err
}
if err := repo.CreateProfile(tx, profile); err != nil {
    tx.Rollback()
    return err
}
tx.Commit()

// Hyperion UnitOfWork
uow.WithTransaction(ctx, func(txCtx hyperctx.Context) error {
    if err := repo.Create(txCtx, user); err != nil {
        return err // Auto rollback
    }
    if err := repo.CreateProfile(txCtx, profile); err != nil {
        return err // Auto rollback
    }
    return nil // Auto commit
})
```

### Consequences

**Positive:**
- More concise code
- Reduced transaction management errors
- Unified transaction boundary management

**Negative:**
- No support for nested transactions (GORM limitation, can use savepoints)

---

## ADR-006: Configuration Hot Reload Support

### Status
✅ **Accepted**

### Context
Production environments need to dynamically adjust configurations (e.g., log levels) without restarting services.

### Decision
`hyperconfig.Provider` provides `Watch()` interface to support configuration change callbacks.

### Core Design

```go
type Watcher interface {
    Watch(callback func(event ChangeEvent)) (stop func(), err error)
}

// Usage example
cfgProvider.Watch(func(event ChangeEvent) {
    // Reload configuration
    var cfg LogConfig
    cfgProvider.Unmarshal("log", &cfg)
    logger.SetLevel(cfg.Level)
})
```

### Rationale

**Advantages:**
1. **Runtime Adjustment**: No need to restart services
2. **Multiple Sources**: File (fsnotify) / Consul / Etcd
3. **Callback Mechanism**: Components can independently listen for config changes
4. **Safety**: Config validation failures don't affect existing config

**Implementation:**
- File config: Use `fsnotify` to watch file changes
- Consul: Use Watch API with long polling
- Etcd: Use Watch API

### Consequences

**Positive:**
- Enhanced operational flexibility
- Support for dynamic config adjustments

**Negative:**
- Increased implementation complexity
- Need to handle concurrency safety for config changes

---

## ADR-007: Dynamic Log Level Adjustment

### Status
✅ **Accepted**

### Context
Troubleshooting production issues requires temporarily elevating log levels without restarting services.

### Decision
`hyperlog.Logger` provides `SetLevel()` method to support runtime adjustment.

### Core Design

```go
type Logger interface {
    SetLevel(level Level)
    GetLevel() Level
}

// Integration with config hot reload
cfgProvider.Watch(func(event ChangeEvent) {
    var cfg LogConfig
    cfgProvider.Unmarshal("log", &cfg)
    logger.SetLevel(parseLevel(cfg.Level))
})
```

### Rationale

**Advantages:**
1. **Runtime Adjustment**: No restart needed
2. **Controlled Performance Impact**: Log level checking on hot path, Zap uses atomic
3. **Config System Integration**: Automatically takes effect via config hot reload

**Implementation:**
- Use `zap.AtomicLevel` to ensure concurrency safety
- Call `SetLevel()` on config changes

### Consequences

**Positive:**
- Improved troubleshooting efficiency
- Reduced production environment risk

**Negative:**
- Need to ensure concurrency safety for log level changes

---

## ADR-008: Modular Architecture - Microkernel Design

### Status
✅ **Accepted**

### Context
The framework should be lightweight, and applications should only import needed functionality.

### Decision
Adopt microkernel architecture, providing all functionality as `fx.Module`.

### Core Design

```go
// Core entry
func Core() fx.Option {
    return fx.Options(
        hyperconfig.Module,
        hyperlog.Module,
        hyperdb.Module,
    )
}

// Web application
func Web() fx.Option {
    return fx.Options(Core(), hyperweb.Module)
}

// gRPC application
func GRPC() fx.Option {
    return fx.Options(Core(), hypergrpc.Module)
}

// Full-stack application
func FullStack() fx.Option {
    return fx.Options(Core(), hyperweb.Module, hypergrpc.Module)
}
```

### Rationale

**Advantages:**
1. **Import on Demand**: Applications only import needed modules
2. **Independent Evolution**: Modules can be upgraded independently
3. **Clear Boundaries**: Clear dependencies between modules
4. **Easy Extension**: Third parties can provide custom modules

**Module List:**

| Module | Required | Description |
|--------|----------|-------------|
| `hyperconfig` | ✅ | Configuration management |
| `hyperlog` | ✅ | Logging |
| `hyperdb` | ❌ | Database (optional) |
| `hypercache` | ❌ | Cache (optional) |
| `hyperweb` | ❌ | Web server (optional) |
| `hypergrpc` | ❌ | gRPC server (optional) |

### Consequences

**Positive:**
- Smaller application binaries
- Clearer dependencies
- Easier to understand and maintain

**Negative:**
- Need to explicitly import modules

---

## Summary

Hyperion framework's core design decisions revolve around these principles:

1. **Type Safety**: Achieved through `hyperctx.Context` and `hypererror.Code`
2. **Concise API**: Tracing integrated into Context, reduced boilerplate
3. **Declarative**: UnitOfWork transaction management, fx dependency injection
4. **Runtime Flexibility**: Config hot reload, dynamic log levels
5. **Modularity**: Microkernel architecture, import on demand

These decisions collectively form Hyperion's technical foundation, making it a production-grade, observable, and maintainable Go backend framework.

---

**End of ADR Document**
