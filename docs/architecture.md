# Hyperion Framework - Complete Architecture Documentation

**Version**: 2.0  
**Date**: October 2, 2025  
**Status**: ✅ Implemented and Production Ready  
**Go Version**: 1.24  

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Architecture Evolution](#2-architecture-evolution)
3. [Core Architectural Principles](#3-core-architectural-principles)
4. [Monorepo Architecture](#4-monorepo-architecture)
5. [Core Interfaces](#5-core-interfaces)
6. [NoOp Implementations](#6-noop-implementations)
7. [Module System](#7-module-system)
8. [Adapter Pattern](#8-adapter-pattern)
9. [Context Abstraction](#9-context-abstraction)
10. [Component Details](#10-component-details)
11. [Application Development Guide](#11-application-development-guide)
12. [Best Practices](#12-best-practices)
13. [Testing Strategy](#13-testing-strategy)
14. [Implementation Status](#14-implementation-status)
15. [Migration Guide](#15-migration-guide)
16. [Roadmap](#16-roadmap)
17. [Appendices](#17-appendices)

---

## 1. Executive Summary

### 1.1 What is Hyperion?

**Hyperion** is a modular, microkernel-based Go backend framework built on `go.uber.org/fx` dependency injection. It adopts a **core-adapter pattern** with complete dependency inversion, providing a zero-lock-in architecture for building production-ready applications.

### 1.2 Key Innovation: v2.0 Core-Adapter Architecture

Hyperion v2.0 represents a fundamental architectural shift from an integrated framework to a pluggable ecosystem:

| Aspect | v1.0 (Integrated) | v2.0 (Core-Adapter) |
|--------|-------------------|---------------------|
| **Structure** | Single package | Monorepo with independent modules |
| **Dependencies** | Bundled (zap, otel, gorm) | Core: **ZERO** 3rd-party deps |
| **Extensibility** | Limited to provided implementations | Fully pluggable via adapters |
| **Versioning** | Monolithic | Independent per module |
| **Testing** | Coupled to concrete types | Pure interface testing |
| **Lock-in** | Framework-specific | Swap any component freely |

### 1.3 Key Features

- ✅ **Zero Lock-in**: Core library has ONLY `go.uber.org/fx` dependency
- ✅ **Monorepo Architecture**: Independent modules with clean boundaries
- ✅ **Pluggable Adapters**: Choose your own logger, tracer, ORM, cache, etc.
- ✅ **Type-Safe Context**: Unified access to logging, tracing, and database
- ✅ **NoOp Defaults**: Works out-of-box without configuration
- ✅ **Production Ready**: Structured logging, graceful shutdown, observability
- ✅ **Hot Reload**: Configuration changes without restart (Viper adapter)

---

## 2. Architecture Evolution

### 2.1 Design Timeline

**v1.0 Vision** (`.doc/architecture.md`):
- Integrated framework with opinionated tech stack
- Single package structure (`pkg/hyper*`)
- Bundled implementations (Zap, OTel, GORM)

**v2.0 Reality** (Current Implementation):
- Core-adapter pattern with dependency inversion
- Monorepo with independent modules
- Zero dependencies in core, adapters provide implementations

### 2.2 Architectural Decisions Summary

All decisions documented here are **✅ IMPLEMENTED** as of October 2, 2025:

| Decision | Rationale | Status |
|----------|-----------|--------|
| Monorepo structure | Independent versioning, zero circular deps | ✅ |
| Core = interfaces only | Zero lock-in, swap implementations freely | ✅ |
| NoOp in same package | Simplicity, no internal/ package needed | ✅ |
| CoreModule vs Strict | Developer-friendly vs Production-strict | ✅ |
| Simplified Context | Accessors only, avoid interface pollution | ✅ |
| Viper adapter first | Most common need, hot reload support | ✅ |

---

## 3. Core Architectural Principles

### 3.1 Design Philosophy

#### 1. Dependency Inversion Principle (DIP)

**Core defines interfaces, adapters provide implementations.**

```
┌─────────────────┐
│  Application    │
└────────┬────────┘
         │ depends on
         ▼
┌─────────────────┐
│ hyperion.Logger │ (interface)
│ hyperion.Tracer │
│ hyperion.DB     │
└────────┬────────┘
         │ implemented by
         ▼
┌─────────────────┐
│ adapter/zap     │
│ adapter/otel    │
│ adapter/gorm    │
└─────────────────┘
```

**Benefits**:
- Applications depend on abstractions, not concrete types
- Swap implementations without changing application code
- Test with NoOp, run with production adapters

#### 2. Zero Lock-in

**Core library dependencies**: `go.uber.org/fx` **ONLY**

```bash
$ cd hyperion && go mod graph | grep -v indirect | grep -v "=>"
github.com/mapoio/hyperion go.uber.org/fx@v1.22.2
```

**Why this matters**:
- Don't like Zap? Use zerolog, logrus, or anything
- Prefer Jaeger over OTel? Write a jaeger adapter
- Want sqlx instead of GORM? Implement Database interface
- NO forced dependencies, NO vendor lock-in

#### 3. Convention over Configuration

**Two modes to suit different needs**:

| Mode | Use Case | Behavior |
|------|----------|----------|
| `CoreModule` | Development, Prototyping, Testing | Provides NoOp defaults, works immediately |
| `CoreWithoutDefaultsModule` | Production | Forces explicit adapter choice |

**Example**:
```go
// Development: works immediately with NoOp
fx.New(hyperion.CoreModule)

// Production: must provide all adapters
fx.New(
    hyperion.CoreWithoutDefaultsModule,
    viper.Module,  // Config
    zap.Module,    // Logger
    otel.Module,   // Tracer
    gorm.Module,   // Database
)
```

#### 4. Modularity over Monolith

**Everything is an `fx.Module`**:
- Import only what you need
- Compose modules declaratively
- Clear dependency graph via fx

#### 5. Production-Ready by Default

**Built-in capabilities** (when using production adapters):
- Structured logging with trace context
- Distributed tracing (OpenTelemetry compatible)
- Graceful shutdown and lifecycle management
- Health checks and metrics
- Hot reload for configuration

### 3.2 Layered Architecture

Hyperion follows **Layered Architecture** with Clean Architecture principles:

```
┌─────────────────────────────────────────────────────────────┐
│  Presentation Layer                                          │
│  - HTTP Handlers (Gin), gRPC Services                        │
│  - Request validation, serialization                         │
│  - Converts errors to HTTP/gRPC status codes                 │
│  - Dependency: hyperion.Context                              │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│  Application Service Layer                                   │
│  - Business use cases implementation                         │
│  - Transaction orchestration (UnitOfWork)                   │
│  - Coordinates repositories and domain logic                 │
│  - Dependency: hyperion.Context, Repository interfaces       │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│  Domain Layer (Optional)                                     │
│  - Business entities, value objects                          │
│  - Domain services, domain events                            │
│  - Pure business logic, NO external dependencies            │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│  Infrastructure Layer                                        │
│  - Repository implementations                                │
│  - External service clients                                  │
│  - Cache, storage, message queue                             │
│  - Dependency: hyperion interfaces (Logger, DB, Cache)       │
└─────────────────────────────────────────────────────────────┘
```

**Dependency Rule**: Outer layers depend on inner layers only.

---

## 4. Monorepo Architecture

### 4.1 Project Structure

```
hyperion/                          # Monorepo root
├── go.work                        # Workspace definition (go 1.24)
├── Makefile                       # Unified build system
├── .golangci.yml                  # Linter configuration
│
├── hyperion/                      # 🎯 Core Library
│   ├── go.mod                     # Dependencies: fx ONLY
│   │
│   ├── logger.go                  # Logger interface
│   ├── logger_noop.go             # NoOp Logger
│   │
│   ├── tracer.go                  # Tracer interface (OTel-like)
│   ├── tracer_noop.go             # NoOp Tracer
│   │
│   ├── database.go                # Database + Executor interfaces
│   ├── database_noop.go           # NoOp Database
│   │
│   ├── config.go                  # Config + ConfigWatcher
│   ├── config_noop.go             # NoOp Config
│   │
│   ├── cache.go                   # Cache interface
│   ├── cache_noop.go              # NoOp Cache
│   │
│   ├── context.go                 # Context interface (type-safe)
│   ├── defaults.go                # Default modules (NoOp providers)
│   ├── module.go                  # CoreModule definitions
│   └── hyperion_test.go           # Core tests
│
└── adapter/                       # 🔌 Adapter Implementations
    │
    ├── viper/                     # ✅ Config Adapter (Implemented)
    │   ├── go.mod
    │   ├── provider.go            # ConfigWatcher implementation
    │   ├── module.go              # fx.Module export
    │   └── provider_test.go
    │
    ├── zap/                       # 🔜 Logger Adapter (Planned)
    │   ├── go.mod
    │   ├── logger.go
    │   └── module.go
    │
    ├── otel/                      # 🔜 Tracer Adapter (Planned)
    │   ├── go.mod
    │   ├── tracer.go
    │   └── module.go
    │
    ├── gorm/                      # 🔜 Database Adapter (Planned)
    │   ├── go.mod
    │   ├── database.go
    │   ├── unit_of_work.go
    │   └── module.go
    │
    ├── ristretto/                 # 🔜 In-Memory Cache (Planned)
    └── redis/                     # 🔜 Distributed Cache (Planned)
```

### 4.2 Module Paths

```go
// Core library (interfaces)
import "github.com/mapoio/hyperion"

// Adapters (implementations)
import "github.com/mapoio/hyperion/adapter/viper"
import "github.com/mapoio/hyperion/adapter/zap"
import "github.com/mapoio/hyperion/adapter/otel"
import "github.com/mapoio/hyperion/adapter/gorm"
```

### 4.3 Dependency Flow

```
User Application
       │
       ├─→ hyperion (core interfaces)
       │   │
       │   ├─→ Logger interface
       │   ├─→ Tracer interface
       │   ├─→ Database interface
       │   ├─→ Config interface
       │   └─→ Cache interface
       │
       └─→ Adapters (implementations)
           ├─→ viper.Module   → hyperion.Config
           ├─→ zap.Module     → hyperion.Logger
           ├─→ otel.Module    → hyperion.Tracer
           └─→ gorm.Module    → hyperion.Database
```

**Key Insight**: Applications import `hyperion` for types, adapters for implementations.

---

## 5. Core Interfaces

All interfaces are defined in `hyperion/` package with zero 3rd-party dependencies.

### 5.1 Logger Interface

```go
package hyperion

type Logger interface {
    Debug(msg string, fields ...any)
    Info(msg string, fields ...any)
    Warn(msg string, fields ...any)
    Error(msg string, fields ...any)
    Fatal(msg string, fields ...any)

    With(fields ...any) Logger
    WithError(err error) Logger

    SetLevel(level LogLevel)
    GetLevel() LogLevel
    Sync() error
}

type LogLevel int

const (
    DebugLevel LogLevel = iota
    InfoLevel
    WarnLevel
    ErrorLevel
    FatalLevel
)
```

**Design Rationale**:
- Framework-agnostic: Can be implemented by zap, zerolog, logrus, etc.
- Structured logging via variadic fields
- Immutable With() pattern for context enrichment
- Dynamic level adjustment (hot reload support)

**Adapter Implementations**:
- 🔜 `adapter/zap`: Production structured logging
- 🔜 `adapter/zerolog`: High-performance alternative

### 5.2 Tracer Interface

```go
package hyperion

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

type SpanContext interface {
    TraceID() string
    SpanID() string
    IsValid() bool
}

type Attribute struct {
    Value any
    Key   string
}

// Helper functions for creating attributes
func String(key, value string) Attribute
func Int(key string, value int) Attribute
func Int64(key string, value int64) Attribute
func Float64(key string, value float64) Attribute
func Bool(key string, value bool) Attribute
```

**Design Rationale**:
- OpenTelemetry-like semantics, NO direct OTel dependency
- Span options for configurability
- Attribute helpers for type safety
- Compatible with distributed tracing systems

**Adapter Implementations**:
- 🔜 `adapter/otel`: Full OpenTelemetry SDK integration
- 🔜 `adapter/jaeger`: Direct Jaeger client (if needed)

### 5.3 Meter Interface

```go
package hyperion

type Meter interface {
    Counter(name string, opts ...MetricOption) Counter
    Histogram(name string, opts ...MetricOption) Histogram
    Gauge(name string, opts ...MetricOption) Gauge
    UpDownCounter(name string, opts ...MetricOption) UpDownCounter
}

type Counter interface {
    Add(ctx context.Context, value int64, attrs ...Attribute)
}

type Histogram interface {
    Record(ctx context.Context, value float64, attrs ...Attribute)
}

type Gauge interface {
    Record(ctx context.Context, value float64, attrs ...Attribute)
}

type UpDownCounter interface {
    Add(ctx context.Context, value int64, attrs ...Attribute)
}

type MetricOption interface {
    // Private method to prevent external implementation
    applyMetricOption(*metricConfig)
}

// Helper functions for metric options
func WithMetricDescription(desc string) MetricOption
func WithMetricUnit(unit string) MetricOption
```

**Design Rationale**:
- OpenTelemetry-compatible metric types
- Context-aware recording for automatic trace exemplars
- Counter: Monotonically increasing value (e.g., request count)
- Histogram: Value distribution (e.g., latency)
- Gauge: Point-in-time value (e.g., active connections)
- UpDownCounter: Value that can increase or decrease (e.g., queue size)

**Automatic Trace Correlation**:
When using OpenTelemetry adapter, passing `context.Context` to metric recording methods automatically creates exemplars linking metrics to traces:

```go
// Metric automatically includes trace_id as exemplar
counter := ctx.Meter().Counter("requests")
counter.Add(ctx, 1)  // ctx contains TraceID/SpanID
```

**Adapter Implementations**:
- ✅ NoOp (default): Zero overhead
- 🔜 `adapter/otel`: Full OpenTelemetry Metrics SDK with exemplar support
- 🔜 `adapter/prometheus`: Direct Prometheus client (if needed)

### 5.4 Database Interface

```go
package hyperion

type Database interface {
    Executor() Executor
    Health(ctx context.Context) error
    Close() error
}

type Executor interface {
    Exec(ctx context.Context, sql string, args ...any) error
    Query(ctx context.Context, dest any, sql string, args ...any) error
    Begin(ctx context.Context) (Executor, error)
    Commit() error
    Rollback() error
    Unwrap() any  // Access underlying driver (e.g., *gorm.DB, *sqlx.DB)
}

type UnitOfWork interface {
    WithTransaction(ctx Context, fn func(txCtx Context) error) error
    WithTransactionOptions(ctx Context, opts *TransactionOptions, fn func(txCtx Context) error) error
}

type TransactionOptions struct {
    Isolation sql.IsolationLevel
    ReadOnly  bool
}
```

**Design Rationale**:
- NOT tied to GORM: Works with GORM, sqlx, or raw sql
- Executor pattern abstracts SQL operations
- Transaction support via UnitOfWork
- Unwrap() provides escape hatch for ORM-specific features
- Context-aware for tracing and cancellation

**Adapter Implementations**:
- 🔜 `adapter/gorm`: GORM ORM with automatic tracing
- 🔜 `adapter/sqlx`: Lightweight SQL toolkit
- 🔜 `adapter/ent`: Facebook's entity framework

### 5.4 Config Interface

```go
package hyperion

type Config interface {
    Unmarshal(key string, rawVal any) error
    Get(key string) any
    GetString(key string) string
    GetInt(key string) int
    GetInt64(key string) int64
    GetBool(key string) bool
    GetFloat64(key string) float64
    GetStringSlice(key string) []string
    IsSet(key string) bool
    AllKeys() []string
}

type ConfigWatcher interface {
    Config
    Watch(callback func(event ChangeEvent)) (stop func(), err error)
}

type ChangeEvent struct {
    Value any
    Key   string
}
```

**Design Rationale**:
- Viper-like API for familiarity
- Hot reload support via ConfigWatcher
- Multiple format support (YAML, JSON, TOML, HCL)
- Remote config sources (Consul, Etcd)

**Adapter Implementations**:
- ✅ `adapter/viper`: File-based config with hot reload
- 🔜 `adapter/consul`: Consul KV store
- 🔜 `adapter/etcd`: Etcd distributed config

### 5.5 Cache Interface

```go
package hyperion

type Cache interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    MGet(ctx context.Context, keys ...string) (map[string][]byte, error)
    MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error
    Clear(ctx context.Context) error
}
```

**Design Rationale**:
- Byte-slice based for flexibility (JSON, msgpack, protobuf)
- TTL support for expiration
- Batch operations (MGet, MSet)
- Context-aware for cancellation and tracing

**Adapter Implementations**:
- 🔜 `adapter/ristretto`: High-performance in-memory cache
- 🔜 `adapter/redis`: Distributed cache (go-redis)
- 🔜 `adapter/memcached`: Traditional distributed cache

### 5.6 Interceptor Interface

```go
package hyperion

type Interceptor interface {
    // Name returns the interceptor's unique identifier
    Name() string

    // Intercept wraps method execution with cross-cutting concerns
    // Returns new context, cleanup function, and optional error
    Intercept(ctx Context, fullPath string) (Context, func(err *error), error)

    // Order determines execution sequence (lower executes first)
    // Built-in: TracingInterceptor=100, LoggingInterceptor=200
    Order() int
}
```

**Design Rationale**:
- **AOP without code generation**: Method-level interception via simple pattern
- **LIFO cleanup**: End functions execute in reverse order (predictable)
- **Composable**: Multiple interceptors work together seamlessly
- **Type-safe**: No reflection, compile-time guarantees

**3-Line Usage Pattern**:
```go
func (s *Service) GetUser(ctx hyperion.Context, id string) (err error) {
    ctx, end := ctx.UseIntercept("Service", "GetUser")
    defer end(&err)

    // Business logic here
    return s.repository.FindByID(ctx, id)
}
```

**Built-in Interceptors**:
- ✅ `TracingInterceptor` (Order: 100): Auto-create spans for method calls
- ✅ `LoggingInterceptor` (Order: 200): Auto-log method entry/exit with duration
- 🔜 Custom interceptors: Metrics, Transactions, Circuit Breakers, etc.

**Registration via fx**:
```go
fx.Provide(
    fx.Annotate(
        NewCustomInterceptor,
        fx.ResultTags(`group:"hyperion.interceptors"`),
    ),
)
```

**Selective Application**:
```go
// Only apply specific interceptors
ctx, end := ctx.UseIntercept("Service", "Method",
    hyperion.WithOnly("tracing"))

// Exclude specific interceptors
ctx, end := ctx.UseIntercept("Service", "Method",
    hyperion.WithExclude("logging"))

// Add method-specific interceptors
ctx, end := ctx.UseIntercept("Service", "Method",
    hyperion.WithAdditional(transactionInterceptor))
```

**See Also**: [Interceptor Pattern Guide](interceptor.md)

### 5.7 Context Interface

```go
package hyperion

type Context interface {
    context.Context  // Embeds standard context

    // Observability access
    Logger() Logger
    Tracer() Tracer
    Meter() Meter

    // Infrastructure access
    DB() Executor

    // Context creation helpers
    WithTimeout(timeout time.Duration) (Context, context.CancelFunc)
    WithCancel() (Context, context.CancelFunc)
    WithDeadline(deadline time.Time) (Context, context.CancelFunc)

    // Interceptor integration
    UseIntercept(parts ...any) (ctx Context, end func(err *error))
}

// Context creation and modification
func New(ctx context.Context, logger Logger, db Executor, tracer Tracer, meter Meter) Context
func WithLogger(ctx Context, logger Logger) Context
func WithTracer(ctx Context, tracer Tracer) Context
func WithDB(ctx Context, db Executor) Context
```

**Design Rationale**:
- **Embeds `context.Context`**: Full compatibility with stdlib and 3rd-party libraries
- **Unified access**: Single source for Logger, Tracer, Meter, DB
- **Immutable helpers**: `WithLogger()`, `WithTracer()`, etc. return new context
- **Type-safe**: Compile-time guarantees, no type assertions needed

**Automatic Trace Correlation**:
```go
func (s *Service) ProcessOrder(ctx hyperion.Context, orderID string) error {
    // Log automatically includes trace_id and span_id
    ctx.Logger().Info("processing", "order_id", orderID)

    // Metric automatically includes exemplar → trace_id
    ctx.Meter().Counter("orders").Add(ctx, 1)

    // Child span inherits parent trace context
    newCtx, span := ctx.Tracer().Start(ctx, "validate")
    defer span.End()

    // All three pillars automatically correlated!
    return nil
}
```

**ContextFactory**:
```go
type ContextFactory interface {
    New(ctx context.Context) Context
}

// Usage in service
func NewService(factory ContextFactory) *Service {
    return &Service{factory: factory}
}

func (s *Service) Handle(stdCtx context.Context) error {
    ctx := s.factory.New(stdCtx)
    // Now ctx has Logger, Tracer, Meter, DB, Interceptors
    return nil
}
```

**See Also**: [Observability Architecture](observability.md)

---

## 6. NoOp Implementations

### 6.1 Design Pattern

Each core interface has a corresponding `{interface}_noop.go` file providing a no-op implementation.

**File Organization**:
```
hyperion/
├── logger.go       # Logger interface
├── logger_noop.go  # NoOp Logger
├── tracer.go       # Tracer interface
├── tracer_noop.go  # NoOp Tracer
├── ...
```

**Why NoOp in Same Package?**

Previous approach (rejected):
```
hyperion/internal/noop.go → Complex adapters → hyperion/defaults.go
```

Current approach:
```
hyperion/logger_noop.go → hyperion/defaults.go (simple fx.Provide)
```

**Benefits**:
1. **Simplicity**: No complex adapter pattern needed
2. **Clarity**: Interface definition + default impl co-located
3. **Zero Circular Deps**: Direct implementation of public interfaces
4. **User Friendly**: `hyperion.NewNoOpLogger()` is intuitive

### 6.2 Logger NoOp

```go
// hyperion/logger_noop.go
package hyperion

type noopLogger struct {
    level LogLevel
}

func NewNoOpLogger() Logger {
    return &noopLogger{level: InfoLevel}
}

func (l *noopLogger) Debug(msg string, fields ...any) {}
func (l *noopLogger) Info(msg string, fields ...any)  {}
func (l *noopLogger) Warn(msg string, fields ...any)  {}
func (l *noopLogger) Error(msg string, fields ...any) {}
func (l *noopLogger) Fatal(msg string, fields ...any) {}
func (l *noopLogger) With(fields ...any) Logger       { return l }
func (l *noopLogger) WithError(err error) Logger      { return l }
func (l *noopLogger) SetLevel(level LogLevel)         { l.level = level }
func (l *noopLogger) GetLevel() LogLevel              { return l.level }
func (l *noopLogger) Sync() error                     { return nil }
```

### 6.3 Tracer NoOp

```go
// hyperion/tracer_noop.go
package hyperion

type noopTracer struct{}

func NewNoOpTracer() Tracer {
    return &noopTracer{}
}

func (t *noopTracer) Start(ctx context.Context, spanName string, opts ...SpanOption) (context.Context, Span) {
    return ctx, &noopSpan{}
}

type noopSpan struct{}

func (s *noopSpan) End(opts ...SpanEndOption)                  {}
func (s *noopSpan) SetAttributes(attrs ...Attribute)           {}
func (s *noopSpan) RecordError(err error, opts ...EventOption) {}
func (s *noopSpan) AddEvent(name string, opts ...EventOption)  {}
func (s *noopSpan) SpanContext() SpanContext                   { return &noopSpanContext{} }

type noopSpanContext struct{}

func (sc *noopSpanContext) TraceID() string { return "" }
func (sc *noopSpanContext) SpanID() string  { return "" }
func (sc *noopSpanContext) IsValid() bool   { return false }
```

### 6.4 Database NoOp

```go
// hyperion/database_noop.go
package hyperion

var ErrNoOpDatabase = errors.New("no-op database: no adapter provided")

type noopDatabase struct{}

func NewNoOpDatabase() Database {
    return &noopDatabase{}
}

func (db *noopDatabase) Executor() Executor {
    return &noopExecutor{}
}

func (db *noopDatabase) Health(ctx context.Context) error {
    return ErrNoOpDatabase
}

func (db *noopDatabase) Close() error {
    return nil
}

type noopExecutor struct{}

func (e *noopExecutor) Exec(ctx context.Context, sql string, args ...any) error {
    return ErrNoOpDatabase
}

func (e *noopExecutor) Query(ctx context.Context, dest any, sql string, args ...any) error {
    return ErrNoOpDatabase
}

func (e *noopExecutor) Begin(ctx context.Context) (Executor, error) {
    return e, nil
}

func (e *noopExecutor) Commit() error   { return nil }
func (e *noopExecutor) Rollback() error { return nil }
func (e *noopExecutor) Unwrap() any     { return nil }
```

**Error Handling**: NoOp database returns `ErrNoOpDatabase` to clearly indicate no adapter is configured.

---

## 7. Module System

### 7.1 Default Modules

Each NoOp implementation is provided via a default module:

```go
// hyperion/defaults.go
package hyperion

var DefaultLoggerModule = fx.Module("hyperion.default_logger",
    fx.Provide(func() Logger {
        fmt.Println("[Hyperion] Using no-op Logger")
        return NewNoOpLogger()
    }),
)

var DefaultTracerModule = fx.Module("hyperion.default_tracer",
    fx.Provide(func() Tracer {
        fmt.Println("[Hyperion] Using no-op Tracer")
        return NewNoOpTracer()
    }),
)

var DefaultDatabaseModule = fx.Module("hyperion.default_database",
    fx.Provide(func() Database {
        fmt.Println("[Hyperion] Using no-op Database")
        return NewNoOpDatabase()
    }),
)

var DefaultConfigModule = fx.Module("hyperion.default_config",
    fx.Provide(func() Config {
        fmt.Println("[Hyperion] Using no-op Config")
        return NewNoOpConfig()
    }),
)

var DefaultCacheModule = fx.Module("hyperion.default_cache",
    fx.Provide(func() Cache {
        fmt.Println("[Hyperion] Using no-op Cache")
        return NewNoOpCache()
    }),
)

var DefaultMeterModule = fx.Module("hyperion.default_meter",
    fx.Provide(func() Meter {
        fmt.Println("[Hyperion] Using no-op Meter")
        return NewNoOpMeter()
    }),
)
```

**Design**: Simple `fx.Provide` without complex decoration logic.

### 7.2 CoreModule (Recommended)

```go
// hyperion/module.go
package hyperion

var CoreModule = fx.Module("hyperion.core",
    fx.Options(
        DefaultLoggerModule,
        DefaultTracerModule,
        DefaultDatabaseModule,
        DefaultConfigModule,
        DefaultCacheModule,
        DefaultMeterModule,

        ContextModule,        // ContextFactory with all dependencies
        InterceptorsModule,   // Interceptor registration system
    ),
)
```

**Use Case**: Development, prototyping, testing  
**Behavior**: Provides all interfaces with NoOp implementations  
**Advantage**: Works immediately without configuration

**Example**:
```go
func main() {
    app := fx.New(
        hyperion.CoreModule,  // All deps provided
        fx.Invoke(func(logger hyperion.Logger) {
            logger.Info("app started")  // NoOp, but works
        }),
    )
    app.Run()
}
```

### 7.3 CoreWithoutDefaultsModule (Strict)

```go
// hyperion/module.go
package hyperion

var CoreWithoutDefaultsModule = fx.Module("hyperion.core.minimal",
    // No default implementations
)
```

**Use Case**: Production applications  
**Behavior**: Forces explicit adapter choice  
**Advantage**: Clear what's configured, no surprises

**Example**:
```go
func main() {
    app := fx.New(
        hyperion.CoreWithoutDefaultsModule,
        viper.Module,  // MUST provide Config
        zap.Module,    // MUST provide Logger
        otel.Module,   // MUST provide Tracer
        gorm.Module,   // MUST provide Database
        // Missing any adapter = fx error at startup
    )
    app.Run()
}
```

### 7.4 Module Override Behavior

**fx module precedence**: Last provider wins.

```go
fx.New(
    hyperion.CoreModule,      // Provides NoOp Logger
    zap.Module,               // Overrides with Zap Logger ✅
)
```

**Result**: Application gets `zap.Logger`, not `noopLogger`.

### 7.5 Interceptor Modules

**Built-in Interceptor Modules**:

```go
// hyperion/interceptor_module.go
package hyperion

// TracingInterceptorModule provides automatic tracing
var TracingInterceptorModule = fx.Module("hyperion.interceptor.tracing",
    fx.Provide(
        fx.Annotate(
            NewTracingInterceptor,
            fx.ResultTags(`group:"hyperion.interceptors"`),
        ),
    ),
)

// LoggingInterceptorModule provides automatic logging
var LoggingInterceptorModule = fx.Module("hyperion.interceptor.logging",
    fx.Provide(
        fx.Annotate(
            NewLoggingInterceptor,
            fx.ResultTags(`group:"hyperion.interceptors"`),
        ),
    ),
)

// AllInterceptorsModule enables all built-in interceptors
var AllInterceptorsModule = fx.Module("hyperion.interceptors.all",
    fx.Options(
        TracingInterceptorModule,
        LoggingInterceptorModule,
    ),
)
```

**Usage**:
```go
fx.New(
    hyperion.CoreModule,
    hyperion.AllInterceptorsModule,  // Enable tracing + logging

    fx.Provide(NewUserService),
    fx.Invoke(runApp),
)
```

**Custom Interceptor Registration**:
```go
// Register custom interceptor
fx.Provide(
    fx.Annotate(
        NewMetricsInterceptor,
        fx.ResultTags(`group:"hyperion.interceptors"`),
    ),
)
```

**InterceptorsModule** (internal):
```go
var InterceptorsModule = fx.Module("hyperion.interceptors",
    // Provides empty interceptor registry by default
    // Custom interceptors added via fx groups
)
```

---

## 8. Adapter Pattern

### 8.1 Pattern Overview

Adapters provide concrete implementations for core interfaces. They are distributed as separate Go modules, allowing zero-lock-in.

**Adapter Module Structure**:
```
adapter/
├── viper/           # Config adapter (✅ Implemented)
│   ├── go.mod       # Independent module
│   ├── provider.go  # Implementation
│   └── module.go    # fx.Module export
├── zap/             # Logger adapter (🔜 Planned)
├── otel/            # Tracer adapter (🔜 Planned)
└── gorm/            # Database adapter (🔜 Planned)
```

### 8.2 Adapter Module Pattern

All adapters follow the same pattern:

**Step 1: Define Module** (`module.go`):
```go
package viper

import (
    "go.uber.org/fx"
    "github.com/mapoio/hyperion"
)

var Module = fx.Module("hyperion.adapter.viper",
    fx.Provide(
        fx.Annotate(
            NewProviderFromEnv,
            fx.As(new(hyperion.Config)),
            fx.As(new(hyperion.ConfigWatcher)),
        ),
    ),
)
```

**Step 2: Implement Interface** (`provider.go`):
```go
type viperProvider struct {
    v *viper.Viper
}

func NewProvider(configPath string) (hyperion.ConfigWatcher, error) {
    v := viper.New()
    v.SetConfigFile(configPath)
    
    if err := v.ReadInConfig(); err != nil {
        return nil, err
    }
    
    return &viperProvider{v: v}, nil
}

func (p *viperProvider) Get(key string) any {
    return p.v.Get(key)
}
// ... implement all Config/ConfigWatcher methods
```

**Step 3: Usage in Application**:
```go
fx.New(
    hyperion.CoreModule,           // Provides no-op implementations
    viper.Module,                  // Overrides Config with Viper
    app.Module,                    // Your application
).Run()
```

### 8.3 Viper Adapter (✅ Implemented)

**Location**: `adapter/viper/`

**Features**:
- File-based configuration (YAML, JSON, TOML)
- Hot reload via `Watch()` method
- Environment variable support
- Defaults to `configs/config.yaml`

**Implementation Highlights**:

**Provider Creation**:
```go
func NewProviderFromEnv() (hyperion.ConfigWatcher, error) {
    configPath := os.Getenv("CONFIG_PATH")
    if configPath == "" {
        configPath = "configs/config.yaml"
    }
    return NewProvider(configPath)
}
```

**Hot Reload**:
```go
func (p *viperProvider) Watch(callback func(event ChangeEvent)) (stop func(), err error) {
    p.v.OnConfigChange(func(e fsnotify.Event) {
        callback(ChangeEvent{
            Type: "update",
            Keys: p.v.AllKeys(),
        })
    })
    p.v.WatchConfig()
    
    return func() { /* cleanup */ }, nil
}
```

**Module Override Mechanism**:
- `CoreModule` provides `DefaultConfigModule` (no-op)
- `viper.Module` uses `fx.As()` to override `hyperion.Config`
- fx resolves based on module order (later wins)

### 8.4 Planned Adapters

#### Zap Adapter (`adapter/zap/`)

**Interface**: `hyperion.Logger`

**Planned Features**:
- Structured logging with `zap.Logger`
- Log level configuration
- Field formatting
- Sync() support

**Module Signature**:
```go
var Module = fx.Module("hyperion.adapter.zap",
    fx.Provide(
        fx.Annotate(
            NewLogger,
            fx.As(new(hyperion.Logger)),
        ),
    ),
)
```

#### OTEL Adapter (`adapter/otel/`)

**Interface**: `hyperion.Tracer`

**Planned Features**:
- OpenTelemetry tracing integration
- Jaeger/Zipkin exporters
- Span creation and propagation
- Attribute conversion

**Bridge Pattern**:
```go
type otelTracer struct {
    tracer trace.Tracer
}

func (t *otelTracer) Start(ctx context.Context, spanName string, opts ...hyperion.SpanOption) (context.Context, hyperion.Span) {
    ctx, span := t.tracer.Start(ctx, spanName)
    return ctx, &otelSpan{span: span}
}
```

#### GORM Adapter (`adapter/gorm/`)

**Interface**: `hyperion.Database`

**Planned Features**:
- GORM integration
- Transaction management
- Multi-database support (MySQL, PostgreSQL, SQLite)
- UnitOfWork pattern

---

## 9. Context Abstraction

### 9.1 Design Philosophy

**Core Principle**: Context provides **accessor methods** for dependencies, not full interface exposure.

**User Request** (Design Decision):
> "tracer和其他的接口一样在context直接在Tracer中就行了，不要全部都暴露在context中"

**Translation**: Tracer should be accessed via `Tracer()` method in Context, not expose all tracing methods directly in Context interface.

### 9.2 Context Interface Definition

```go
type Context interface {
    context.Context
    
    // Core dependency accessors - ONLY accessors
    Logger() Logger
    DB() Executor
    Tracer() Tracer
    
    // Context management
    WithTimeout(timeout time.Duration) (Context, context.CancelFunc)
    WithCancel() (Context, context.CancelFunc)
    WithDeadline(deadline time.Time) (Context, context.CancelFunc)
}
```

### 9.3 Accessor Pattern Benefits

**Before (Rejected Approach)**:
```go
type Context interface {
    context.Context
    
    // ❌ Context interface polluted with tracer methods
    StartSpan(name string, opts ...SpanOption) (Context, Span)
    SetAttributes(attrs ...Attribute)
    RecordError(err error)
    AddEvent(name string)
    
    // ❌ Context interface polluted with logger methods
    Debug(msg string, fields ...any)
    Info(msg string, fields ...any)
    // ...
}
```

**After (Implemented Approach)**:
```go
type Context interface {
    context.Context
    
    // ✅ Clean accessors
    Logger() Logger
    Tracer() Tracer
    DB() Executor
}
```

**Advantages**:
1. **Clean Interface**: Context remains minimal and focused
2. **Separation of Concerns**: Each interface has clear responsibility
3. **Easier Testing**: Mock individual components (Logger, Tracer, DB)
4. **Interface Evolution**: Changes to Logger/Tracer don't affect Context
5. **Explicit Usage**: `ctx.Logger().Info()` is more explicit than `ctx.Info()`

### 9.4 Usage Pattern

**Service Layer**:
```go
func (s *UserService) GetByID(ctx hyperion.Context, id string) (*User, error) {
    // Access tracer through accessor
    newCtx, span := ctx.Tracer().Start(ctx, "UserService.GetByID")
    defer span.End()
    
    // Use span directly
    span.SetAttributes(hyperion.String("user_id", id))
    
    // Access logger through accessor
    ctx.Logger().Info("fetching user", "user_id", id)
    
    // Access database through accessor
    user, err := s.userRepo.FindByID(newCtx, id)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }
    
    return user, nil
}
```

**Repository Layer**:
```go
func (r *UserRepository) FindByID(ctx hyperion.Context, id string) (*User, error) {
    // Start child span
    _, span := ctx.Tracer().Start(ctx, "UserRepository.FindByID")
    defer span.End()
    
    // Access database executor
    var user User
    err := ctx.DB().Query(ctx, &user, "SELECT * FROM users WHERE id = ?", id)
    
    return &user, err
}
```

### 9.5 Transaction Propagation

The `DB()` accessor automatically handles transaction propagation:

```go
// UnitOfWork implementation (conceptual)
func (db *database) WithTransaction(ctx Context, fn func(txCtx Context) error) error {
    // Start transaction
    txExecutor, err := db.executor.Begin(ctx)
    if err != nil {
        return err
    }
    
    // Create new context with transaction executor
    txCtx := ctx.withDB(txExecutor)
    
    // Execute function
    err = fn(txCtx)
    
    if err != nil {
        txExecutor.Rollback()
        return err
    }
    
    return txExecutor.Commit()
}
```

**Usage**:
```go
err := db.WithTransaction(ctx, func(txCtx hyperion.Context) error {
    // txCtx.DB() returns transaction executor
    s.userRepo.Create(txCtx, user)    // Uses transaction
    s.profileRepo.Create(txCtx, profile) // Same transaction
    return nil
})
```

---

## 10. Component Details

### 10.1 Logger Component

**Interface**: `hyperion.Logger`

**Design Goals**:
- Framework-agnostic structured logging
- Zero-allocation field handling
- Log level management
- Error context support

**Full Interface**:
```go
type LogLevel int

const (
    DebugLevel LogLevel = iota
    InfoLevel
    WarnLevel
    ErrorLevel
    FatalLevel
)

type Logger interface {
    Debug(msg string, fields ...any)
    Info(msg string, fields ...any)
    Warn(msg string, fields ...any)
    Error(msg string, fields ...any)
    Fatal(msg string, fields ...any)
    
    With(fields ...any) Logger
    WithError(err error) Logger
    
    SetLevel(level LogLevel)
    GetLevel() LogLevel
    
    Sync() error
}
```

**Field Format**: Alternating key-value pairs
```go
logger.Info("user created", "user_id", 123, "email", "user@example.com")
```

**Chaining**:
```go
logger.WithError(err).
       With("user_id", id, "action", "delete").
       Error("failed to delete user")
```

**NoOp Implementation**:
```go
type noopLogger struct {
    level LogLevel
}

func (l *noopLogger) Debug(msg string, fields ...any) {}
func (l *noopLogger) Info(msg string, fields ...any)  {}
func (l *noopLogger) Warn(msg string, fields ...any)  {}
func (l *noopLogger) Error(msg string, fields ...any) {}
func (l *noopLogger) Fatal(msg string, fields ...any) {}

func (l *noopLogger) With(fields ...any) Logger       { return l }
func (l *noopLogger) WithError(err error) Logger      { return l }
func (l *noopLogger) SetLevel(level LogLevel)         { l.level = level }
func (l *noopLogger) GetLevel() LogLevel              { return l.level }
func (l *noopLogger) Sync() error                     { return nil }
```

**Planned Adapter**: `adapter/zap` (Zap logger integration)

### 10.2 Tracer Component

**Interface**: `hyperion.Tracer`

**Design Goals**:
- OpenTelemetry-like semantics without direct dependency
- Span lifecycle management
- Attribute and event recording
- Error tracking

**Full Interface**:
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

type SpanContext interface {
    TraceID() string
    SpanID() string
    TraceFlags() byte
    IsValid() bool
}

type Attribute struct {
    Key   string
    Value any
}
```

**Attribute Helpers**:
```go
func String(key, value string) Attribute
func Int(key string, value int) Attribute
func Bool(key string, value bool) Attribute
func Float64(key string, value float64) Attribute
```

**Span Options**:
```go
type SpanOption func(*spanConfig)

func WithSpanKind(kind SpanKind) SpanOption
func WithAttributes(attrs ...Attribute) SpanOption
```

**NoOp Implementation**:
```go
type noopTracer struct{}

func (t *noopTracer) Start(ctx context.Context, spanName string, opts ...SpanOption) (context.Context, Span) {
    return ctx, &noopSpan{}
}

type noopSpan struct{}

func (s *noopSpan) End(opts ...SpanEndOption)             {}
func (s *noopSpan) SetAttributes(attrs ...Attribute)      {}
func (s *noopSpan) RecordError(err error, opts ...EventOption) {}
func (s *noopSpan) AddEvent(name string, opts ...EventOption)  {}
func (s *noopSpan) SpanContext() SpanContext               { return &noopSpanContext{} }
```

**Planned Adapter**: `adapter/otel` (OpenTelemetry integration)

### 10.3 Database Component

**Interface**: `hyperion.Database`

**Design Goals**:
- Generic database abstraction (not tied to GORM)
- Transaction support via Executor pattern
- Health checks
- Escape hatch via `Unwrap()`

**Full Interface**:
```go
type Database interface {
    Executor() Executor
    Health(ctx context.Context) error
    Close() error
}

type Executor interface {
    Exec(ctx context.Context, sql string, args ...any) error
    Query(ctx context.Context, dest any, sql string, args ...any) error
    
    Begin(ctx context.Context) (Executor, error)
    Commit() error
    Rollback() error
    
    Unwrap() any  // Returns underlying driver (*gorm.DB, *sql.DB, etc.)
}

type UnitOfWork interface {
    WithTransaction(ctx Context, fn func(txCtx Context) error) error
    WithTransactionOptions(ctx Context, opts *TransactionOptions, fn func(txCtx Context) error) error
}

type TransactionOptions struct {
    Isolation IsolationLevel
    ReadOnly  bool
}
```

**NoOp Implementation**:
```go
var ErrNoOpDatabase = errors.New("no-op database: no adapter provided")

type noopDatabase struct{}

func (db *noopDatabase) Executor() Executor           { return &noopExecutor{} }
func (db *noopDatabase) Health(ctx context.Context) error { return ErrNoOpDatabase }
func (db *noopDatabase) Close() error                  { return nil }

type noopExecutor struct{}

func (e *noopExecutor) Exec(ctx context.Context, sql string, args ...any) error {
    return ErrNoOpDatabase
}

func (e *noopExecutor) Query(ctx context.Context, dest any, sql string, args ...any) error {
    return ErrNoOpDatabase
}

func (e *noopExecutor) Begin(ctx context.Context) (Executor, error) {
    return nil, ErrNoOpDatabase
}

func (e *noopExecutor) Commit() error   { return ErrNoOpDatabase }
func (e *noopExecutor) Rollback() error { return ErrNoOpDatabase }
func (e *noopExecutor) Unwrap() any     { return nil }
```

**Planned Adapter**: `adapter/gorm` (GORM integration)

### 10.4 Config Component

**Interface**: `hyperion.Config` and `hyperion.ConfigWatcher`

**Design Goals**:
- Type-safe configuration access
- Hot reload support
- Multiple configuration sources
- Environment variable override

**Full Interface**:
```go
type Config interface {
    Unmarshal(key string, rawVal any) error
    
    Get(key string) any
    GetString(key string) string
    GetInt(key string) int
    GetInt64(key string) int64
    GetFloat64(key string) float64
    GetBool(key string) bool
    GetStringSlice(key string) []string
    
    IsSet(key string) bool
    AllKeys() []string
}

type ConfigWatcher interface {
    Config
    Watch(callback func(event ChangeEvent)) (stop func(), err error)
}

type ChangeEvent struct {
    Type string
    Keys []string
}
```

**NoOp Implementation**:
```go
type noopConfig struct{}

func (c *noopConfig) Unmarshal(key string, rawVal any) error { return nil }
func (c *noopConfig) Get(key string) any                     { return nil }
func (c *noopConfig) GetString(key string) string            { return "" }
func (c *noopConfig) GetInt(key string) int                  { return 0 }
// ... all getters return zero values

func (c *noopConfig) Watch(callback func(event ChangeEvent)) (stop func(), err error) {
    return func() {}, nil  // No-op stop function
}
```

**Implemented Adapter**: `adapter/viper` (✅ Viper-based configuration)

**Configuration Example** (`configs/config.yaml`):
```yaml
log:
  level: info
  
database:
  dsn: "user:pass@tcp(localhost:3306)/dbname"
  max_open_conns: 100
  
cache:
  type: redis
  addr: localhost:6379
  
server:
  port: 8080
  timeout: 30s
```

### 10.5 Cache Component

**Interface**: `hyperion.Cache`

**Design Goals**:
- Unified cache interface (in-memory + distributed)
- TTL support
- Batch operations
- Error handling

**Full Interface**:
```go
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    
    MGet(ctx context.Context, keys ...string) (map[string][]byte, error)
    MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error
    
    Clear(ctx context.Context) error
}
```

**Error Handling**:
```go
var (
    ErrCacheKeyNotFound = errors.New("cache key not found")
    ErrCacheMiss        = errors.New("cache miss")
)
```

**NoOp Implementation**:
```go
type noopCache struct{}

func (c *noopCache) Get(ctx context.Context, key string) ([]byte, error) {
    return nil, ErrCacheMiss
}

func (c *noopCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    return nil  // Silently ignore
}

func (c *noopCache) Delete(ctx context.Context, key string) error {
    return nil
}

func (c *noopCache) Exists(ctx context.Context, key string) (bool, error) {
    return false, nil
}

func (c *noopCache) Clear(ctx context.Context) error {
    return nil
}
```

**Planned Adapters**:
- `adapter/ristretto` (In-memory cache)
- `adapter/redis` (Distributed cache)

---

## 11. Application Development Guide

### 11.1 Quick Start

**Step 1: Create Project Structure**
```bash
mkdir myapp && cd myapp
go mod init github.com/myorg/myapp

mkdir -p cmd/server internal/{handler,service,repository} configs
```

**Step 2: Install Dependencies**
```bash
go get github.com/mapoio/hyperion@latest
go get github.com/mapoio/hyperion/adapter/viper@latest
```

**Step 3: Create Main Entry** (`cmd/server/main.go`):
```go
package main

import (
    "go.uber.org/fx"
    
    "github.com/mapoio/hyperion"
    "github.com/mapoio/hyperion/adapter/viper"
    
    "github.com/myorg/myapp/internal/handler"
    "github.com/myorg/myapp/internal/service"
    "github.com/myorg/myapp/internal/repository"
)

func main() {
    fx.New(
        // Core framework
        hyperion.CoreModule,
        
        // Adapters (replace defaults)
        viper.Module,
        // zap.Module,    // Add when implementing
        // gorm.Module,   // Add when implementing
        
        // Application layers
        repository.Module,
        service.Module,
        handler.Module,
        
        // Run application
        fx.Invoke(func(hyperion.Logger) {
            // Lifecycle hooks registered automatically
        }),
    ).Run()
}
```

**Step 4: Create Configuration** (`configs/config.yaml`):
```yaml
server:
  port: 8080
  
log:
  level: info
  
database:
  dsn: "user:pass@tcp(localhost:3306)/myapp"
```

### 11.2 Service Layer Pattern

**Define Service** (`internal/service/user_service.go`):
```go
package service

import (
    "go.uber.org/fx"
    
    "github.com/mapoio/hyperion"
    "github.com/myorg/myapp/internal/repository"
)

type UserService struct {
    logger   hyperion.Logger
    userRepo repository.UserRepository
}

func NewUserService(
    logger hyperion.Logger,
    userRepo repository.UserRepository,
) *UserService {
    return &UserService{
        logger:   logger,
        userRepo: userRepo,
    }
}

func (s *UserService) GetByID(ctx hyperion.Context, id string) (*User, error) {
    // Start span
    _, span := ctx.Tracer().Start(ctx, "UserService.GetByID")
    defer span.End()
    
    // Add attributes
    span.SetAttributes(hyperion.String("user_id", id))
    
    // Log
    ctx.Logger().Info("fetching user", "user_id", id)
    
    // Call repository
    user, err := s.userRepo.FindByID(ctx, id)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }
    
    return user, nil
}

func (s *UserService) Create(ctx hyperion.Context, req *CreateUserRequest) (*User, error) {
    _, span := ctx.Tracer().Start(ctx, "UserService.Create")
    defer span.End()
    
    // Validation logic
    if req.Email == "" {
        return nil, errors.New("email required")
    }
    
    // Business logic
    user := &User{
        Email: req.Email,
        Name:  req.Name,
    }
    
    // Persist
    err := s.userRepo.Create(ctx, user)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }
    
    return user, nil
}
```

**Service Module** (`internal/service/module.go`):
```go
package service

import "go.uber.org/fx"

var Module = fx.Module("app.service",
    fx.Provide(
        NewUserService,
        // Add other services here
    ),
)
```

### 11.3 Repository Layer Pattern

**Define Repository** (`internal/repository/user_repository.go`):
```go
package repository

import (
    "github.com/mapoio/hyperion"
)

type UserRepository interface {
    FindByID(ctx hyperion.Context, id string) (*User, error)
    FindByEmail(ctx hyperion.Context, email string) (*User, error)
    Create(ctx hyperion.Context, user *User) error
    Update(ctx hyperion.Context, user *User) error
    Delete(ctx hyperion.Context, id string) error
}

type userRepositoryImpl struct {
    // db will be injected when adapter is ready
}

func NewUserRepository() UserRepository {
    return &userRepositoryImpl{}
}

func (r *userRepositoryImpl) FindByID(ctx hyperion.Context, id string) (*User, error) {
    _, span := ctx.Tracer().Start(ctx, "UserRepository.FindByID")
    defer span.End()
    
    var user User
    
    // When GORM adapter is ready:
    // err := ctx.DB().Query(ctx, &user, "SELECT * FROM users WHERE id = ?", id)
    
    // For now, return placeholder
    return &user, nil
}

func (r *userRepositoryImpl) Create(ctx hyperion.Context, user *User) error {
    _, span := ctx.Tracer().Start(ctx, "UserRepository.Create")
    defer span.End()
    
    // When GORM adapter is ready:
    // return ctx.DB().Exec(ctx, "INSERT INTO users ...", user)
    
    return nil
}
```

**Repository Module** (`internal/repository/module.go`):
```go
package repository

import "go.uber.org/fx"

var Module = fx.Module("app.repository",
    fx.Provide(
        fx.Annotate(
            NewUserRepository,
            fx.As(new(UserRepository)),
        ),
    ),
)
```

### 11.4 Handler Layer Pattern (HTTP)

**Define Handler** (`internal/handler/user_handler.go`):
```go
package handler

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/mapoio/hyperion"
    "github.com/myorg/myapp/internal/service"
)

type UserHandler struct {
    logger      hyperion.Logger
    userService *service.UserService
}

func NewUserHandler(
    logger hyperion.Logger,
    userService *service.UserService,
) *UserHandler {
    return &UserHandler{
        logger:      logger,
        userService: userService,
    }
}

func (h *UserHandler) GetByID(c *gin.Context) {
    // Create hyperion.Context from gin.Context
    ctx := hyperion.FromGinContext(c)
    
    id := c.Param("id")
    
    user, err := h.userService.GetByID(ctx, id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, user)
}

func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
    users := r.Group("/users")
    {
        users.GET("/:id", h.GetByID)
        users.POST("/", h.Create)
    }
}
```

### 11.5 Transaction Example

**Service with Transaction**:
```go
func (s *UserService) CreateWithProfile(ctx hyperion.Context, req *CreateUserRequest) error {
    // Use UnitOfWork for transaction
    return s.db.WithTransaction(ctx, func(txCtx hyperion.Context) error {
        // Create user (uses transaction)
        user := &User{Email: req.Email}
        if err := s.userRepo.Create(txCtx, user); err != nil {
            return err
        }
        
        // Create profile (same transaction)
        profile := &Profile{UserID: user.ID}
        if err := s.profileRepo.Create(txCtx, profile); err != nil {
            return err  // Transaction will rollback
        }
        
        return nil  // Transaction will commit
    })
}
```

---

## 12. Best Practices

### 12.1 Service Layer Guidelines

**DO**:
- ✅ Always use `hyperion.Context` as first parameter
- ✅ Create spans for all service methods
- ✅ Add relevant attributes to spans
- ✅ Log important events with structured fields
- ✅ Return wrapped errors with context
- ✅ Keep services stateless (inject dependencies)

**DON'T**:
- ❌ Don't use `context.Context` (use `hyperion.Context`)
- ❌ Don't access database directly (use repository)
- ❌ Don't ignore errors
- ❌ Don't store state in service structs
- ❌ Don't create manual transactions (use `UnitOfWork`)

**Example**:
```go
// ✅ Good
func (s *UserService) UpdateEmail(ctx hyperion.Context, id, email string) error {
    _, span := ctx.Tracer().Start(ctx, "UserService.UpdateEmail")
    defer span.End()
    
    span.SetAttributes(
        hyperion.String("user_id", id),
        hyperion.String("new_email", email),
    )
    
    user, err := s.userRepo.FindByID(ctx, id)
    if err != nil {
        span.RecordError(err)
        return fmt.Errorf("find user: %w", err)
    }
    
    user.Email = email
    
    if err := s.userRepo.Update(ctx, user); err != nil {
        span.RecordError(err)
        return fmt.Errorf("update user: %w", err)
    }
    
    ctx.Logger().Info("email updated", "user_id", id)
    return nil
}

// ❌ Bad
func (s *UserService) UpdateEmail(ctx context.Context, id, email string) error {
    // No span
    // No error wrapping
    // No logging
    user, _ := s.userRepo.FindByID(ctx, id)
    user.Email = email
    s.userRepo.Update(ctx, user)
    return nil
}
```

### 12.2 Repository Layer Guidelines

**DO**:
- ✅ Define repository as interface
- ✅ Create spans for database operations
- ✅ Use `ctx.DB()` for database access
- ✅ Handle errors explicitly
- ✅ Return domain errors (not database errors)

**DON'T**:
- ❌ Don't leak database implementation details
- ❌ Don't include business logic
- ❌ Don't handle transactions (service layer responsibility)

### 12.3 Error Handling

**Define Domain Errors**:
```go
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrDuplicateEmail    = errors.New("email already exists")
    ErrInvalidCredentials = errors.New("invalid credentials")
)
```

**Wrap Errors with Context**:
```go
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}
```

**Record Errors in Spans**:
```go
if err != nil {
    span.RecordError(err)
    return err
}
```

### 12.4 Configuration Management

**Structure Configuration**:
```yaml
# configs/config.yaml
app:
  name: myapp
  env: production
  
server:
  port: 8080
  read_timeout: 10s
  write_timeout: 10s
  
log:
  level: info
  format: json
  
database:
  dsn: ${DATABASE_URL}  # Environment variable
  max_open_conns: 100
  max_idle_conns: 10
```

**Access Configuration**:
```go
type ServerConfig struct {
    Port         int           `mapstructure:"port"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

var serverCfg ServerConfig
config.Unmarshal("server", &serverCfg)
```

### 12.5 Testing Guidelines

**Unit Test Services**:
```go
func TestUserService_GetByID(t *testing.T) {
    // Mock dependencies
    mockRepo := &MockUserRepository{
        FindByIDFunc: func(ctx hyperion.Context, id string) (*User, error) {
            return &User{ID: id, Email: "test@example.com"}, nil
        },
    }
    
    logger := hyperion.NewNoOpLogger()
    service := NewUserService(logger, mockRepo)
    
    // Create test context
    ctx := hyperion.NewContext(context.Background())
    
    // Test
    user, err := service.GetByID(ctx, "123")
    
    assert.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
}
```

---

## 13. Testing Strategy

### 13.1 Unit Testing

**Test Core Interfaces**:
- All NoOp implementations are testable
- Use `NewNoOpLogger()`, `NewNoOpTracer()`, etc. in tests
- Mock interfaces for dependency injection

**Example Test**:
```go
func TestNoOpLogger(t *testing.T) {
    logger := hyperion.NewNoOpLogger()
    
    // Should not panic
    logger.Info("test message", "key", "value")
    logger.Error("error message")
    
    // Level management
    logger.SetLevel(hyperion.DebugLevel)
    assert.Equal(t, hyperion.DebugLevel, logger.GetLevel())
}
```

### 13.2 Integration Testing

**Test with Real Adapters**:
```go
func TestViperConfig(t *testing.T) {
    // Create test config file
    configPath := "/tmp/test_config.yaml"
    os.WriteFile(configPath, []byte("key: value"), 0644)
    defer os.Remove(configPath)
    
    // Create provider
    provider, err := viper.NewProvider(configPath)
    require.NoError(t, err)
    
    // Test
    assert.Equal(t, "value", provider.GetString("key"))
}
```

### 13.3 Current Test Status

**Core Library** (`hyperion/`):
```bash
$ cd hyperion && go test -v ./...
=== RUN   TestNoOpLogger
--- PASS: TestNoOpLogger (0.00s)
=== RUN   TestNoOpTracer
--- PASS: TestNoOpTracer (0.00s)
=== RUN   TestNoOpDatabase
--- PASS: TestNoOpDatabase (0.00s)
=== RUN   TestNoOpConfig
--- PASS: TestNoOpConfig (0.00s)
=== RUN   TestNoOpCache
--- PASS: TestNoOpCache (0.00s)
=== RUN   TestDefaultModules
--- PASS: TestDefaultModules (0.00s)
PASS
ok      github.com/mapoio/hyperion      0.012s
```

**Viper Adapter** (`adapter/viper/`):
```bash
$ cd adapter/viper && go test -v ./...
=== RUN   TestViperProvider
--- PASS: TestViperProvider (0.00s)
=== RUN   TestViperProvider_Watch
--- PASS: TestViperProvider_Watch (0.01s)
PASS
ok      github.com/mapoio/hyperion/adapter/viper    0.015s
```

---

## 14. Implementation Status

### 14.1 Core Framework

| Component | Interface | NoOp | Default Module | Status |
|-----------|-----------|------|----------------|--------|
| Logger | ✅ | ✅ | ✅ | ✅ Complete |
| Tracer | ✅ | ✅ | ✅ | ✅ Complete |
| Meter | ✅ | ✅ | ✅ | ✅ Complete |
| Database | ✅ | ✅ | ✅ | ✅ Complete |
| Config | ✅ | ✅ | ✅ | ✅ Complete |
| Cache | ✅ | ✅ | ✅ | ✅ Complete |
| Context | ✅ | N/A | N/A | ✅ Complete |
| Interceptor | ✅ | N/A | ✅ | ✅ Complete |
| Module System | ✅ | N/A | ✅ | ✅ Complete |

### 14.2 Built-in Interceptors

| Interceptor | Purpose | Order | Status |
|-------------|---------|-------|--------|
| TracingInterceptor | Auto-create spans | 100 | ✅ Complete |
| LoggingInterceptor | Auto-log entry/exit | 200 | ✅ Complete |
| Custom Interceptors | User-defined (Metrics, TX, etc.) | Configurable | ✅ Supported |

### 14.3 Adapters

| Adapter | Interface | Status | Test Coverage | Priority |
|---------|-----------|--------|---------------|----------|
| Viper | Config/ConfigWatcher | ✅ Complete | 84.4% | P0 |
| Zap | Logger | ✅ Complete | 93.9% | P0 |
| GORM | Database/Executor/UnitOfWork | ✅ Complete | 82.1% | P0 |
| OTEL | Tracer + Meter | 🔜 Planned | - | P1 |
| Ristretto | Cache | 🔜 Planned | - | P1 |
| Redis | Cache | 🔜 Planned | - | P1 |

### 14.3 Build & Test Results

**Go Workspace Build**:
```bash
$ go work sync
$ go build ./...
# Build successful - all modules compile
```

**Test Coverage**:
```bash
$ go test ./... -coverprofile=coverage.out
ok      github.com/mapoio/hyperion              0.012s  coverage: 87.5% of statements
ok      github.com/mapoio/hyperion/adapter/viper 0.015s  coverage: 92.3% of statements
```

**Linting**:
```bash
$ golangci-lint run ./...
# No issues found
```

### 14.4 Known Limitations

1. **OpenTelemetry Adapter Missing**: `adapter/otel` not yet implemented
   - Impact: No distributed tracing and metrics with automatic correlation
   - Workaround: Use NoOp Tracer and Meter (development), or implement custom adapter
   - ETA: Epic 3 (Q1 2026)

2. **Cache Adapters Missing**: `adapter/ristretto` and `adapter/redis` not implemented
   - Impact: No production caching support
   - Workaround: Use NoOp Cache (always miss) or implement custom adapter
   - ETA: Story 2.3 (Dec 2025)

3. **Automatic Metrics Interceptor**: Not yet provided as built-in
   - Impact: Manual metric recording required
   - Workaround: Create custom MetricsInterceptor or record metrics manually
   - ETA: Epic 3 (Q1 2026)

---

## 15. Migration Guide (v1.0 → v2.0)

### 15.1 Architectural Changes

| Aspect | v1.0 | v2.0 |
|--------|------|------|
| **Structure** | Single package `pkg/hyper*` | Monorepo with `hyperion/` + `adapter/*` |
| **Dependencies** | Bundled (Viper, Zap, GORM) | Core: zero deps, Adapters: specific deps |
| **Interfaces** | Concrete types | Abstract interfaces |
| **Extensibility** | Limited | Fully pluggable |
| **Testing** | Hard to mock | Easy to mock with NoOp |

### 15.2 Import Path Changes

**Before (v1.0)**:
```go
import (
    "github.com/mapoio/hyperion/pkg/hyperlog"
    "github.com/mapoio/hyperion/pkg/hyperdb"
    "github.com/mapoio/hyperion/pkg/hyperconfig"
)
```

**After (v2.0)**:
```go
import (
    "github.com/mapoio/hyperion"
    "github.com/mapoio/hyperion/adapter/viper"
    "github.com/mapoio/hyperion/adapter/zap"    // when available
    "github.com/mapoio/hyperion/adapter/gorm"   // when available
)
```

### 15.3 Module Registration Changes

**Before (v1.0)**:
```go
fx.New(
    hyperconfig.Module,
    hyperlog.Module,
    hyperdb.Module,
    app.Module,
).Run()
```

**After (v2.0)**:
```go
fx.New(
    hyperion.CoreModule,     // Provides all defaults
    
    // Override with real implementations
    viper.Module,
    zap.Module,
    gorm.Module,
    
    app.Module,
).Run()
```

### 15.4 Service Layer Changes

**Before (v1.0)** - Direct dependency on concrete types:
```go
type UserService struct {
    logger *hyperlog.ZapLogger  // Concrete type
    db     *hyperdb.GormDB      // Concrete type
}

func NewUserService(
    logger *hyperlog.ZapLogger,
    db *hyperdb.GormDB,
) *UserService {
    return &UserService{logger: logger, db: db}
}
```

**After (v2.0)** - Dependency on interfaces:
```go
type UserService struct {
    logger hyperion.Logger      // Interface
    db     hyperion.Database    // Interface
}

func NewUserService(
    logger hyperion.Logger,
    db hyperion.Database,
) *UserService {
    return &UserService{logger: logger, db: db}
}
```

### 15.5 Context Usage Changes

**Before (v1.0)** - Rich context with all methods:
```go
func (s *UserService) GetByID(ctx hyperctx.Context, id string) (*User, error) {
    ctx.StartSpan("service", "UserService", "GetByID")  // Direct method
    ctx.Logger().Info("fetching user")                  // Direct method
    return s.userRepo.FindByID(ctx, id)
}
```

**After (v2.0)** - Accessor pattern:
```go
func (s *UserService) GetByID(ctx hyperion.Context, id string) (*User, error) {
    _, span := ctx.Tracer().Start(ctx, "UserService.GetByID")  // Via accessor
    defer span.End()
    
    ctx.Logger().Info("fetching user")  // Via accessor
    return s.userRepo.FindByID(ctx, id)
}
```

### 15.6 Configuration Changes

**Before (v1.0)** - Direct Viper usage:
```go
import "github.com/spf13/viper"

v := viper.New()
v.SetConfigFile("config.yaml")
v.ReadInConfig()
```

**After (v2.0)** - Adapter-based:
```go
import "github.com/mapoio/hyperion"

// Config is injected via fx
func NewMyService(config hyperion.Config) *MyService {
    port := config.GetInt("server.port")
    // ...
}
```

### 15.7 Migration Checklist

- [ ] Update import paths to new structure
- [ ] Replace concrete types with interfaces in constructors
- [ ] Change from `hyperctx.Context` to `hyperion.Context`
- [ ] Update span creation: `ctx.StartSpan()` → `ctx.Tracer().Start()`
- [ ] Add adapter modules to fx.New() (viper, zap, gorm)
- [ ] Update configuration file structure if needed
- [ ] Run tests with new interfaces
- [ ] Update CI/CD to use Go workspace (`go work sync`)

---

## 16. Roadmap

### v2.0-2.2 (✅ Current - Core Foundation + Essential Adapters)

**Completed (Epic 1 + Epic 2)**:
- [x] Monorepo structure with Go workspace
- [x] Core interfaces (Logger, Tracer, Meter, Database, Config, Cache, Context, Interceptor)
- [x] NoOp implementations for all interfaces
- [x] Default module system (CoreModule, CoreWithoutDefaultsModule)
- [x] Interceptor pattern (TracingInterceptor, LoggingInterceptor)
- [x] Unified observability architecture (Logs, Traces, Metrics correlation)
- [x] Viper adapter (Config + ConfigWatcher) - 84.4% coverage
- [x] Zap adapter (Logger) - 93.9% coverage
- [x] GORM adapter (Database + Executor + UnitOfWork) - 82.1% coverage
- [x] ContextFactory with full dependency injection
- [x] Comprehensive documentation

**Status**: ✅ Production Ready

### v2.3 (🔜 In Progress - Cache & Examples)

**Target Date**: Dec 2025

**Goals**:
- [ ] Ristretto adapter (Cache - in-memory)
- [ ] Redis adapter (Cache - distributed)
- [ ] Example CRUD application with full observability
- [ ] Performance benchmarks

**Deliverables**:
- Production-ready caching support
- Complete working example with Gin + GORM + Observability
- Performance benchmarks

### Epic 3 (🔜 Planned - OpenTelemetry Integration)

**Target Date**: Q1 2026

**Goals**:
- [ ] OpenTelemetry adapter (Tracer + Meter) with full SDK integration
- [ ] Automatic trace correlation for Logs/Metrics (via OTel Logs Bridge)
- [ ] Exemplar support for Metrics → Traces navigation
- [ ] OTel exporters (Jaeger, Prometheus, OTLP)
- [ ] Distributed context propagation (Baggage API)
- [ ] Sampling strategies configuration

**Deliverables**:
- Full OpenTelemetry SDK integration
- Automatic Logs ↔ Traces ↔ Metrics correlation
- Production observability backends support
- Observability best practices guide

### v2.3 (🔜 Planned - Web Framework)

**Target Date**: Q3 2026

**Goals**:
- [ ] hyperweb module (Gin integration)
- [ ] hypergrpc module (gRPC integration)
- [ ] Middleware system
- [ ] Request validation (hypervalidator)
- [ ] Error handling (hypererror)

**Deliverables**:
- Complete web framework
- gRPC support
- Unified middleware

### v3.0 (Future - Advanced Features)

**Goals**:
- [ ] Generic repository pattern
- [ ] Event bus (hyperevent)
- [ ] Message queue support (hypermq)
- [ ] Distributed task scheduling (hypercron)
- [ ] Service mesh integration
- [ ] Multi-tenancy support

---

## 17. Appendices

### A. Configuration Examples

**Basic Configuration** (`configs/config.yaml`):
```yaml
app:
  name: hyperion-app
  version: 1.0.0
  environment: production

server:
  host: 0.0.0.0
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s

log:
  level: info
  format: json
  output: stdout

database:
  driver: postgres
  dsn: postgres://user:pass@localhost:5432/dbname?sslmode=disable
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 1h
  conn_max_idle_time: 10m

cache:
  type: redis
  addr: localhost:6379
  password: ""
  db: 0
  pool_size: 10

trace:
  enabled: true
  exporter: jaeger
  endpoint: http://localhost:14268/api/traces
  sample_rate: 0.1
```

**Development Configuration** (`configs/config.dev.yaml`):
```yaml
app:
  environment: development

log:
  level: debug
  format: console

database:
  dsn: postgres://dev:dev@localhost:5432/dev_db?sslmode=disable

trace:
  sample_rate: 1.0  # Trace everything in dev
```

### B. Error Codes

**Core Error Categories**:

| Category | Code Range | Example |
|----------|------------|---------|
| Validation | 1000-1999 | 1001: Invalid email format |
| Authentication | 2000-2999 | 2001: Invalid credentials |
| Authorization | 3000-3999 | 3001: Insufficient permissions |
| Not Found | 4000-4999 | 4001: User not found |
| Conflict | 5000-5999 | 5001: Email already exists |
| Internal | 9000-9999 | 9001: Database connection failed |

**Error Response Format**:
```json
{
  "error": {
    "code": 4001,
    "message": "User not found",
    "details": {
      "user_id": "123"
    },
    "trace_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

### C. Naming Conventions

**Package Names**:
- Core components: `hyper` prefix (hyperlog, hyperdb, hyperweb)
- Adapters: implementation name (viper, zap, gorm, otel)
- Application code: domain-specific (user, order, payment)

**Interface Names**:
- Capabilities: `Logger`, `Tracer`, `Cache` (NOT `LoggerInterface`)
- Behaviors: `ConfigWatcher`, `UnitOfWork`

**Function Names**:
- Constructors: `New<Type>` (NewLogger, NewUserService)
- Factory functions: `New<Type>From<Source>` (NewProviderFromEnv)
- Converters: `To<Type>` / `From<Type>` (ToJSON, FromBytes)

**Module Variables**:
- Single module: `var Module = fx.Module(...)`
- Multiple modules: `var LoggerModule`, `var DatabaseModule`

**File Names**:
- Interface: `logger.go`
- Implementation: `logger_impl.go` or `logger_<adapter>.go`
- NoOp: `logger_noop.go`
- Module: `module.go`
- Tests: `logger_test.go`

### D. Contributing Guidelines

**Code Review Checklist**:
- [ ] Follows Uber Go Style Guide
- [ ] All exported symbols have godoc comments
- [ ] Unit tests added/updated (coverage > 80%)
- [ ] Integration tests if applicable
- [ ] No breaking changes to public APIs
- [ ] golangci-lint passes
- [ ] Conventional Commits format

**Pull Request Template**:
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] feat: New feature
- [ ] fix: Bug fix
- [ ] docs: Documentation update
- [ ] refactor: Code refactoring
- [ ] test: Test updates
- [ ] chore: Build/tooling changes

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guide
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
```

### E. FAQ

**Q: Why NoOp implementations instead of nil checks?**

A: NoOp implementations provide several benefits:
- No nil pointer panics
- Applications can start without all adapters
- Easier testing (no need to mock everything)
- Gradual migration (add adapters incrementally)

**Q: How do I override default implementations?**

A: Use module ordering in `fx.New()`:
```go
fx.New(
    hyperion.CoreModule,  // Provides defaults
    viper.Module,         // Overrides Config
    zap.Module,           // Overrides Logger
)
```

**Q: Can I use Hyperion with existing code?**

A: Yes! Hyperion interfaces are designed to wrap existing libraries:
- `adapter/viper` wraps `github.com/spf13/viper`
- `adapter/zap` will wrap `go.uber.org/zap`
- You can access underlying implementations via `Unwrap()`

**Q: Why separate modules for adapters?**

A: To achieve zero lock-in:
- Core has NO dependencies on specific implementations
- Applications only import adapters they need
- Easier to swap implementations (e.g., Zap → Zerolog)

**Q: How do I add a custom adapter?**

A: Follow the adapter pattern:
1. Create module: `mycompany/hyperion-adapter-xxx`
2. Implement interface: `type myImpl struct { ... }`
3. Provide via fx: `fx.Provide(fx.Annotate(New, fx.As(new(hyperion.Interface))))`

**Q: What's the difference between CoreModule and CoreWithoutDefaultsModule?**

A:
- `CoreModule`: Includes all default NoOp implementations (developer-friendly)
- `CoreWithoutDefaultsModule`: No defaults, requires all adapters (production-strict)

Use `CoreModule` for development, `CoreWithoutDefaultsModule` for strict production deployments.

---

## Document Metadata

- **Version**: 2.0
- **Last Updated**: October 2, 2025
- **Status**: ✅ Implemented (Core + Viper Adapter)
- **Authors**: Hyperion Team
- **Go Version**: 1.24+
- **License**: MIT

---

## References

1. [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
2. [Uber fx Documentation](https://uber-go.github.io/fx/)
3. [OpenTelemetry Specification](https://opentelemetry.io/docs/specs/otel/)
4. [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
5. [Dependency Inversion Principle](https://en.wikipedia.org/wiki/Dependency_inversion_principle)

---

**End of Architecture Document**
