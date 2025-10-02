# Hyperion Architecture Decision Records (ADR)

**Version**: 2.0
**Date**: October 2025
**Status**: Updated for v2.0 Core-Adapter Architecture

This document records key architectural decisions made during the design of the Hyperion framework and the rationale behind them.

---

## Table of Contents

1. [ADR-001: Choosing go.uber.org/fx as the Dependency Injection Framework](#adr-001-choosing-gouberorgfx-as-the-dependency-injection-framework)
2. [ADR-002: Context Accessor Pattern Design](#adr-002-context-accessor-pattern-design)
3. [ADR-003: OpenTelemetry-Compatible Tracing Without Direct Dependency](#adr-003-opentelemetry-compatible-tracing-without-direct-dependency)
4. [ADR-004: Core-Adapter Pattern for Zero Lock-in](#adr-004-core-adapter-pattern-for-zero-lock-in)
5. [ADR-005: NoOp Implementations in Same Package](#adr-005-noop-implementations-in-same-package)
6. [ADR-006: Monorepo Structure with Go Workspace](#adr-006-monorepo-structure-with-go-workspace)
7. [ADR-007: Two-Mode Module System (CoreModule vs CoreWithoutDefaultsModule)](#adr-007-two-mode-module-system-coremodule-vs-corewithoutdefaultsmodule)
8. [ADR-008: Simple fx.Provide for Default Implementations](#adr-008-simple-fxprovide-for-default-implementations)

---

## ADR-001: Choosing go.uber.org/fx as the Dependency Injection Framework

### Status
✅ **Accepted** (v1.0 and v2.0)

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

### v2.0 Impact
fx remains the ONLY dependency in the core library, reinforcing zero lock-in principle.

### References
- [fx Documentation](https://uber-go.github.io/fx/)

---

## ADR-002: Context Accessor Pattern Design

### Status
✅ **Accepted** (v2.0 - Revised from v1.0)

### Context
Go's standard `context.Context` uses `Value()` for passing data, which is not type-safe. We need a solution that provides type-safe access to dependencies WITHOUT polluting the Context interface with all methods from Logger, Tracer, DB, etc.

### Decision
Design `hyperion.Context` interface with **accessor methods** that return fully-functional interfaces.

### Core Design (v2.0)

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

### Rationale

**Why Accessor Pattern?**
1. **Clean Interface**: Context remains minimal and focused
2. **Separation of Concerns**: Each interface (Logger, Tracer, DB) has clear responsibility
3. **Easier Testing**: Mock individual components independently
4. **Interface Evolution**: Changes to Logger/Tracer don't affect Context
5. **Explicit Usage**: `ctx.Logger().Info()` is clearer than `ctx.Info()`

**Comparison with v1.0 Design:**

| Aspect | v1.0 (Rejected) | v2.0 (Accepted) |
|--------|-----------------|-----------------|
| **Context Methods** | All methods exposed (`ctx.Info()`, `ctx.RecordError()`) | Only accessors (`ctx.Logger()`, `ctx.Tracer()`) |
| **Interface Size** | Large (30+ methods) | Small (6 methods) |
| **Clarity** | Less explicit | More explicit |
| **Testability** | Harder to mock Context | Easy to mock individual components |

**Usage Pattern:**

```go
// v2.0 Accessor Pattern
func (s *UserService) GetByID(ctx hyperion.Context, id string) (*User, error) {
    // Access tracer through accessor
    _, span := ctx.Tracer().Start(ctx, "UserService.GetByID")
    defer span.End()

    // Use span methods directly
    span.SetAttributes(hyperion.String("user_id", id))

    // Access logger through accessor
    ctx.Logger().Info("fetching user", "user_id", id)

    // Access database through accessor
    return s.userRepo.FindByID(ctx, id)
}
```

### Consequences

**Positive:**
- Improved code readability
- Clear separation of concerns
- Easier to extend without breaking Context interface
- Compatible with standard `context.Context`

**Negative:**
- One extra method call: `ctx.Logger().Info()` vs `ctx.Info()`
- Requires type conversion at entry points: `ctx.(hyperion.Context)`

### User Feedback Integration
This decision directly implements user feedback:
> "tracer和其他的接口一样在context直接在Tracer中就行了，不要全部都暴露在context中"

---

## ADR-003: OpenTelemetry-Compatible Tracing Without Direct Dependency

### Status
✅ **Accepted** (v2.0)

### Context
We need distributed tracing support, but requiring OpenTelemetry as a core dependency would violate the zero lock-in principle.

### Decision
Define tracing interfaces in core that are OpenTelemetry-compatible, with actual OTel integration provided via adapter.

### Core Design

**Core Interfaces** (`hyperion/tracer.go`):
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

type Attribute struct {
    Key   string
    Value any
}
```

**Adapter Implementation** (`adapter/otel` - planned):
```go
type otelTracer struct {
    tracer trace.Tracer
}

func (t *otelTracer) Start(ctx context.Context, spanName string, opts ...hyperion.SpanOption) (context.Context, hyperion.Span) {
    ctx, span := t.tracer.Start(ctx, spanName)
    return ctx, &otelSpan{span: span}
}
```

### Rationale

**Advantages:**
1. **Zero Lock-in**: Core has no OTel dependency
2. **OTel Compatible**: Semantics match OpenTelemetry
3. **Flexible**: Can wrap other tracing systems (Jaeger, Zipkin)
4. **Type-Safe**: Attribute helpers prevent errors
5. **Simplified API**: No need to import `attribute` package

**Comparison with Alternatives:**

| Approach | Core Dependency | Flexibility | Complexity |
|----------|----------------|-------------|------------|
| **Direct OTel** | ❌ High | ❌ Low | Low |
| **Our Approach** | ✅ Zero | ✅ High | Medium |
| **Custom Tracing** | ✅ Zero | ❌ Low | High |

### Usage

**Service Layer:**
```go
_, span := ctx.Tracer().Start(ctx, "UserService.GetByID")
defer span.End()

span.SetAttributes(hyperion.String("user_id", id))
if err != nil {
    span.RecordError(err)
}
```

### Consequences

**Positive:**
- Framework users can choose tracing backend
- Easy to test with NoOp tracer
- Compatible with standard OpenTelemetry tools

**Negative:**
- Need to maintain compatibility with OTel semantics
- Attribute conversion overhead (minimal)

---

## ADR-004: Core-Adapter Pattern for Zero Lock-in

### Status
✅ **Accepted** (v2.0 - NEW)

### Context
v1.0 bundled all implementations (Zap, GORM, Viper) which created tight coupling and made it difficult to swap implementations.

### Decision
Adopt **Core-Adapter Pattern** with strict separation:
- **Core** (`hyperion/`): Defines interfaces, ZERO 3rd-party dependencies
- **Adapters** (`adapter/*/`): Provide concrete implementations

### Architecture

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

### Dependency Policy

**Core Library** (`hyperion/go.mod`):
```go
module github.com/mapoio/hyperion

require go.uber.org/fx v1.24.0  // ONLY dependency
```

**Viper Adapter** (`adapter/viper/go.mod`):
```go
module github.com/mapoio/hyperion/adapter/viper

require (
    github.com/mapoio/hyperion v0.0.0
    github.com/spf13/viper v1.21.0
)
```

### Rationale

**Advantages:**
1. **Zero Lock-in**: Users choose implementations
2. **Independent Versioning**: Adapters upgrade independently
3. **Easy Testing**: Test with NoOp, run with real adapters
4. **Smaller Binaries**: Import only needed adapters
5. **Clear Boundaries**: Interface contracts prevent leakage

**Migration Path:**
- v1.0 users: Change imports, add adapter modules to fx.New()
- New users: Start with CoreModule, add adapters as needed

### Consequences

**Positive:**
- Framework is truly pluggable
- Community can provide custom adapters
- Easier to maintain and test

**Negative:**
- Users must explicitly choose adapters
- More import statements (mitigated by clear documentation)

---

## ADR-005: NoOp Implementations in Same Package

### Status
✅ **Accepted** (v2.0 - NEW)

### Context
We need default implementations that allow applications to start without configuration, but where should they live?

### Decision
Place NoOp implementations in the same package as interfaces (`hyperion/logger_noop.go`), not in a separate `internal/` package.

### File Organization

```
hyperion/
├── logger.go       # Logger interface
├── logger_noop.go  # NoOp Logger
├── tracer.go       # Tracer interface
├── tracer_noop.go  # NoOp Tracer
├── ...
```

**Previous Approach (Rejected):**
```
hyperion/internal/noop.go → Complex adapters → hyperion/defaults.go
```

**Current Approach:**
```
hyperion/logger_noop.go → hyperion/defaults.go (simple fx.Provide)
```

### Rationale

**Advantages:**
1. **Simplicity**: No complex adapter pattern needed
2. **Clarity**: Interface definition + default impl co-located
3. **Zero Circular Deps**: Direct implementation of public interfaces
4. **User Friendly**: `hyperion.NewNoOpLogger()` is intuitive
5. **Less Code**: Fewer types, simpler logic

**Comparison:**

| Approach | Files | Complexity | Type Safety |
|----------|-------|------------|-------------|
| **internal/** | More | High (adapters needed) | Medium |
| **Same Package** | Fewer | Low (direct impl) | High |

### Implementation

```go
// hyperion/logger_noop.go
type noopLogger struct {
    level LogLevel
}

func NewNoOpLogger() Logger {
    return &noopLogger{level: InfoLevel}
}

func (l *noopLogger) Info(msg string, fields ...any) {}
// ... all methods are no-op
```

### Consequences

**Positive:**
- Simpler codebase
- Easier to understand for newcomers
- Direct interface implementation

**Negative:**
- NoOp types visible in package namespace (mitigated by lowercase naming)

---

## ADR-006: Monorepo Structure with Go Workspace

### Status
✅ **Accepted** (v2.0 - NEW)

### Context
v1.0 used a single module structure (`pkg/hyper*`). We need better modularity and independent versioning for adapters.

### Decision
Adopt Go workspace-based monorepo with independent modules.

### Structure

```
hyperion/                    # Monorepo root
├── go.work                  # Workspace definition
├── hyperion/                # Core library (zero 3rd-party deps)
│   └── go.mod
└── adapter/                 # Adapter implementations
    ├── viper/               # Config adapter
    │   └── go.mod
    ├── zap/                 # Logger adapter (planned)
    │   └── go.mod
    └── gorm/                # Database adapter (planned)
        └── go.mod
```

**Workspace Definition** (`go.work`):
```
go 1.24

use (
    ./hyperion
    ./adapter/viper
)
```

### Rationale

**Advantages:**
1. **Independent Versioning**: Each adapter can have own version
2. **Zero Circular Dependencies**: Clear module boundaries
3. **Selective Import**: Users import only needed adapters
4. **Easier Testing**: Test each module independently
5. **Clear Ownership**: Each module has own go.mod

**Alternatives Considered:**
- **Single Module**: Simpler but forces all dependencies on users
- **Separate Repos**: Maximum independence but harder to maintain
- **Multi-Module in Single Repo without Workspace**: Complex to manage

### Consequences

**Positive:**
- Clean separation of concerns
- Better dependency management
- Easier to contribute new adapters

**Negative:**
- Requires Go 1.18+ for workspace support
- Slightly more complex initial setup

---

## ADR-007: Two-Mode Module System (CoreModule vs CoreWithoutDefaultsModule)

### Status
✅ **Accepted** (v2.0 - NEW)

### Context
Different use cases need different levels of strictness:
- Development: Quick start with defaults
- Production: Explicit adapter choice

### Decision
Provide two core modules with different behaviors.

### Module Definitions

**CoreModule** (Developer-Friendly):
```go
var CoreModule = fx.Module("hyperion.core",
    fx.Options(
        DefaultLoggerModule,    // NoOp Logger
        DefaultTracerModule,    // NoOp Tracer
        DefaultDatabaseModule,  // NoOp Database
        DefaultConfigModule,    // NoOp Config
        DefaultCacheModule,     // NoOp Cache
    ),
)
```

**CoreWithoutDefaultsModule** (Production-Strict):
```go
var CoreWithoutDefaultsModule = fx.Module("hyperion.core.minimal",
    // No default implementations
)
```

### Usage

**Development:**
```go
fx.New(
    hyperion.CoreModule,  // Works immediately
    app.Module,
).Run()
```

**Production:**
```go
fx.New(
    hyperion.CoreWithoutDefaultsModule,
    viper.Module,  // MUST provide
    zap.Module,    // MUST provide
    gorm.Module,   // MUST provide
    app.Module,
).Run()
// Missing adapter = fx error at startup
```

### Rationale

**Advantages:**
1. **Flexibility**: Users choose strictness level
2. **Quick Start**: CoreModule works out-of-box
3. **Safety**: Production forces explicit choices
4. **Clear Intent**: Module name indicates behavior

**Comparison:**

| Aspect | CoreModule | CoreWithoutDefaultsModule |
|--------|------------|---------------------------|
| **Startup** | Always works | Requires all adapters |
| **Use Case** | Dev, Testing, Prototyping | Production |
| **Safety** | Permissive | Strict |
| **Dependencies** | None (NoOp) | Must provide all |

### Consequences

**Positive:**
- Balances ease of use with production safety
- Clear migration path (start with Core, move to Strict)
- No surprises in production

**Negative:**
- Two modules to maintain (low overhead)
- Need to document when to use each

---

## ADR-008: Simple fx.Provide for Default Implementations

### Status
✅ **Accepted** (v2.0 - NEW)

### Context
We need to provide default (NoOp) implementations. Should we use complex decoration or simple provision?

### Decision
Use simple `fx.Provide` without `fx.Decorate` complexity.

### Implementation

**Approach Chosen:**
```go
var DefaultLoggerModule = fx.Module("hyperion.default_logger",
    fx.Provide(func() Logger {
        fmt.Println("[Hyperion] Using no-op Logger")
        return NewNoOpLogger()
    }),
)
```

**Approach Rejected:**
```go
// ❌ Overly complex, Decorate doesn't work as expected
fx.Provide(func() Logger { return nil }),
fx.Decorate(func(logger Logger) Logger {
    if logger == nil { return NewNoOpLogger() }
    return logger
}),
```

### Rationale

**Advantages:**
1. **Simplicity**: Straightforward to understand
2. **Explicit**: Clear what's being provided
3. **fx Best Practice**: Aligns with fx documentation
4. **Override Mechanism**: Later modules override via fx precedence
5. **No Magic**: No nil checks or decoration logic

**How Override Works:**
```go
fx.New(
    hyperion.CoreModule,      // Provides NoOp Logger
    zap.Module,               // Overrides with Zap Logger
)
// User gets Zap Logger (last provider wins)
```

### Consequences

**Positive:**
- Easier to understand
- Less code
- Clear override semantics

**Negative:**
- None significant

---

## Summary

Hyperion v2.0's architectural decisions revolve around these principles:

1. **Zero Lock-in**: Core-adapter pattern with zero core dependencies
2. **Type Safety**: Accessor pattern for clean interfaces
3. **Flexibility**: Two-mode module system for different use cases
4. **Simplicity**: NoOp in same package, simple fx.Provide
5. **Modularity**: Monorepo with independent adapter modules

These decisions collectively make Hyperion a truly pluggable, production-ready Go framework without vendor lock-in.

---

## Migration from v1.0

| v1.0 Decision | v2.0 Change | Impact |
|---------------|-------------|--------|
| Direct OTel integration | OTel-compatible interfaces | Zero core dependency |
| pkg/hyper* structure | Monorepo with adapters | Independent versioning |
| Rich Context interface | Accessor pattern | Cleaner interfaces |
| Bundled implementations | Core-adapter pattern | True pluggability |

---

**End of ADR Document**
