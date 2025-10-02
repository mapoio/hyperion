# Hyperion v2.0 Implementation Plan

**Version**: 2.0
**Date**: October 2025
**Status**: Monorepo with Core-Adapter Architecture

This document outlines the implementation roadmap for Hyperion v2.0 based on the Core-Adapter pattern.

---

## Architecture Overview

Hyperion v2.0 follows a **Core-Adapter** architecture:

```
┌─────────────────────────────────────┐
│   hyperion/ (Core Library)           │
│   • Interfaces ONLY                  │
│   • Zero 3rd-party deps (except fx)  │
│   • NoOp implementations             │
└──────────────┬──────────────────────┘
               │
    ┌──────────┴──────────────────────┐
    │                                  │
┌───▼────────────┐          ┌─────────▼─────────┐
│ adapter/viper   │          │ adapter/zap        │
│ (Config impl)   │          │ (Logger impl)      │
└─────────────────┘          └────────────────────┘
    │                                  │
┌───▼────────────┐          ┌─────────▼─────────┐
│ adapter/otel    │          │ adapter/gorm       │
│ (Tracer impl)   │          │ (Database impl)    │
└─────────────────┘          └────────────────────┘
```

**Key Principles**:
- Core library defines interfaces with zero lock-in
- Adapters provide concrete implementations
- Independent versioning per adapter
- Users choose which adapters to use

---

## Implementation Status

### ✅ Completed (v2.0-alpha)

#### Core Library (`hyperion/`)
- ✅ **Config interface** - `Config` and `ConfigWatcher` (hyperion/config.go)
- ✅ **Logger interface** - Structured logging interface (hyperion/logger.go)
- ✅ **Tracer interface** - OTel-compatible tracing (hyperion/tracer.go)
- ✅ **Database interface** - Database + Executor + UnitOfWork (hyperion/database.go)
- ✅ **Cache interface** - Key-value cache abstraction (hyperion/cache.go)
- ✅ **Context interface** - Accessor pattern implementation (hyperion/context.go)
- ✅ **NoOp implementations** - All interfaces have NoOp defaults
- ✅ **Module system** - CoreModule and CoreWithoutDefaultsModule (hyperion/module.go)

#### Viper Adapter (`adapter/viper/`)
- ✅ **Config implementation** - Viper-based config provider
- ✅ **ConfigWatcher implementation** - File watching support
- ✅ **Hot reload** - Configuration hot-reloading
- ✅ **Module export** - fx.Module integration
- ✅ **Tests** - Unit and integration tests

#### Infrastructure
- ✅ **Monorepo structure** - Go workspace setup (go.work)
- ✅ **Build system** - Makefile for unified builds
- ✅ **Linting** - golangci-lint configuration
- ✅ **Git hooks** - Pre-commit and commit-msg hooks
- ✅ **Documentation** - Complete v2.0 architecture docs (5,770+ lines)

---

## Implementation Phases

### Phase 1: Core Foundation ✅ (Completed)

**Status**: ✅ **COMPLETED**

**Goal**: Establish core library with zero lock-in

#### 1.1 Core Interfaces ✅
- ✅ Logger interface (hyperion/logger.go)
- ✅ Tracer interface (hyperion/tracer.go)
- ✅ Database interface (hyperion/database.go)
- ✅ Config interface (hyperion/config.go)
- ✅ Cache interface (hyperion/cache.go)
- ✅ Context interface (hyperion/context.go)

**Design Decisions**:
- Accessor pattern for Context (`ctx.Logger()`, `ctx.Tracer()`)
- OTel-compatible Tracer WITHOUT depending on OTel
- Generic Executor interface for database operations
- UnitOfWork pattern for declarative transactions

#### 1.2 NoOp Implementations ✅
- ✅ NoOp Logger (logger_noop.go)
- ✅ NoOp Tracer (tracer_noop.go)
- ✅ NoOp Database (database_noop.go)
- ✅ NoOp Config (config_noop.go)
- ✅ NoOp Cache (cache_noop.go)

**Key Feature**: All NoOp implementations co-located with interfaces in same package.

#### 1.3 Module System ✅
- ✅ CoreModule (provides NoOp defaults)
- ✅ CoreWithoutDefaultsModule (strict mode)
- ✅ DefaultLoggerModule, DefaultTracerModule, etc.

**Module Composition**:
```go
// Developer-friendly (NoOp defaults)
fx.New(hyperion.CoreModule, ...)

// Production-strict (no defaults, must provide all)
fx.New(hyperion.CoreWithoutDefaultsModule, ...)
```

#### 1.4 Viper Adapter ✅
- ✅ Viper-based Config implementation
- ✅ Hot-reload support (file watching)
- ✅ Multi-source configuration (files, env vars)
- ✅ fx.Module integration

**Acceptance Criteria**: ✅ All met
- ✅ Can load configuration from YAML/JSON files
- ✅ Can override with environment variables
- ✅ File changes trigger callbacks
- ✅ Works with fx dependency injection

---

### Phase 2: Logger Adapter 🔜 (v2.1 - Planned)

**Status**: 🔜 **PLANNED**

**Goal**: Implement production-ready logger adapter

#### 2.1 Zap Adapter (`adapter/zap/`) - 5 Days

**Priority**: ⭐⭐⭐⭐⭐

**Dependencies**: Core library (hyperion/)

**Task List**:
- [ ] Implement `zapLogger` struct
- [ ] Implement Logger interface methods (Debug, Info, Warn, Error)
- [ ] Implement `With()` for field chaining
- [ ] Add JSON and Console encoder support
- [ ] Integrate with Config for initialization
- [ ] Implement dynamic level adjustment
- [ ] Add file output support (lumberjack integration)
- [ ] Export fx.Module
- [ ] Write unit tests
- [ ] Write integration tests

**Acceptance Criteria**:
- Can output structured JSON logs
- Can output human-readable console logs
- Can dynamically adjust log level
- Can log to both stdout and file with rotation
- Works seamlessly when replacing NoOp logger

**Module Usage**:
```go
import "github.com/mapoio/hyperion/adapter/zap"

fx.New(
    hyperion.CoreModule,
    viper.Module,
    zap.Module,  // Replaces NoOp Logger
    // ...
)
```

---

### Phase 3: Database Adapter 🔜 (v2.1 - Planned)

**Status**: 🔜 **PLANNED**

**Goal**: Implement GORM-based database adapter with UnitOfWork

#### 3.1 GORM Adapter (`adapter/gorm/`) - 7 Days

**Priority**: ⭐⭐⭐⭐⭐

**Dependencies**: Core library, Zap adapter (for logging)

**Task List**:
- [ ] Implement `gormDatabase` struct
- [ ] Implement Database interface
- [ ] Implement Executor interface (GORM wrapper)
- [ ] Implement connection pool configuration
- [ ] Implement health check (`Ping()`)
- [ ] Implement `gormUnitOfWork` struct
- [ ] Implement `WithTransaction()` method
- [ ] Add transaction propagation via Context
- [ ] Implement TracePlugin (GORM plugin for tracing)
- [ ] Add support for PostgreSQL, MySQL, SQLite
- [ ] Export fx.Module
- [ ] Write unit tests
- [ ] Write integration tests (real databases)

**Acceptance Criteria**:
- Can connect to PostgreSQL, MySQL, SQLite
- Transactions auto-commit on success
- Transactions auto-rollback on error
- `ctx.DB()` returns correct executor (tx or db)
- All database operations create spans (when tracer available)
- Connection pooling works correctly
- Health checks work

**Module Usage**:
```go
import "github.com/mapoio/hyperion/adapter/gorm"

fx.New(
    hyperion.CoreModule,
    viper.Module,
    zap.Module,
    gorm.Module,  // Replaces NoOp Database
    // ...
)
```

**Transaction Example**:
```go
func (s *UserService) RegisterUser(ctx hyperion.Context, req RegisterRequest) error {
    return s.uow.WithTransaction(ctx, func(txCtx hyperion.Context) error {
        // txCtx.DB() returns transaction handle
        if err := s.userRepo.Create(txCtx, user); err != nil {
            return err  // Auto-rollback
        }
        if err := s.profileRepo.Create(txCtx, profile); err != nil {
            return err  // Auto-rollback
        }
        return nil  // Auto-commit
    })
}
```

---

### Phase 4: Tracing Adapter 🔜 (v2.2 - Planned)

**Status**: 🔜 **PLANNED**

**Goal**: Implement OpenTelemetry tracer adapter

#### 4.1 OpenTelemetry Adapter (`adapter/otel/`) - 5 Days

**Priority**: ⭐⭐⭐⭐⭐

**Dependencies**: Core library

**Task List**:
- [ ] Implement `otelTracer` struct
- [ ] Implement Tracer interface (Start method)
- [ ] Implement Span interface (OpenTelemetry span wrapper)
- [ ] Add support for Jaeger exporter
- [ ] Add support for OTLP exporter
- [ ] Add support for stdout exporter (development)
- [ ] Implement trace context propagation
- [ ] Add sampling configuration
- [ ] Export fx.Module
- [ ] Write unit tests
- [ ] Write integration tests (Jaeger)

**Acceptance Criteria**:
- Can create and end spans
- Can add events and attributes to spans
- Can record errors on spans
- Spans correctly nest (parent-child relationships)
- Trace context propagates across service boundaries
- Supports multiple exporters (Jaeger, OTLP, stdout)
- Works with Context interface

**Module Usage**:
```go
import "github.com/mapoio/hyperion/adapter/otel"

fx.New(
    hyperion.CoreModule,
    viper.Module,
    zap.Module,
    gorm.Module,
    otel.Module,  // Replaces NoOp Tracer
    // ...
)
```

**Configuration Example**:
```yaml
tracing:
  enabled: true
  service_name: my-service
  exporter: jaeger
  jaeger:
    endpoint: http://localhost:14268/api/traces
  sampling:
    rate: 1.0  # 100% sampling
```

---

### Phase 5: Cache Adapters 🔜 (v2.2 - Planned)

**Status**: 🔜 **PLANNED**

**Goal**: Implement in-memory and distributed cache adapters

#### 5.1 Ristretto Adapter (`adapter/ristretto/`) - 3 Days

**Priority**: ⭐⭐⭐⭐

**Dependencies**: Core library

**Task List**:
- [ ] Implement `ristrettoCache` struct
- [ ] Implement Cache interface (Get, Set, Delete, etc.)
- [ ] Configure optimal cost-based eviction
- [ ] Add metrics integration (hit/miss rates)
- [ ] Export fx.Module
- [ ] Write unit tests
- [ ] Write benchmark tests

**Acceptance Criteria**:
- Can store and retrieve values
- Eviction works correctly under memory pressure
- High performance (20M+ ops/sec)
- Thread-safe operations

#### 5.2 Redis Adapter (`adapter/redis/`) - 4 Days

**Priority**: ⭐⭐⭐⭐

**Dependencies**: Core library

**Task List**:
- [ ] Implement `redisCache` struct
- [ ] Implement Cache interface
- [ ] Add Redis cluster support
- [ ] Add connection pooling
- [ ] Add retry logic
- [ ] Export fx.Module
- [ ] Write unit tests
- [ ] Write integration tests (real Redis)

**Acceptance Criteria**:
- Can connect to Redis standalone
- Can connect to Redis cluster
- Supports TTL for keys
- Handles connection failures gracefully
- Automatic reconnection

---

### Phase 6: Web Framework 🔜 (v2.3 - Planned)

**Status**: 🔜 **PLANNED**

**Goal**: Implement web server with Gin

#### 6.1 Gin Web Adapter (`adapter/gin/` or `hyperweb/`) - 7 Days

**Priority**: ⭐⭐⭐⭐⭐

**Dependencies**: Core library, all adapters

**Task List**:
- [ ] Define Server struct
- [ ] Implement Gin engine initialization
- [ ] Implement Context injection middleware
- [ ] Implement TraceMiddleware
- [ ] Implement LoggerMiddleware
- [ ] Implement RecoveryMiddleware
- [ ] Implement CORSMiddleware
- [ ] Add metrics middleware (Prometheus)
- [ ] Implement graceful shutdown
- [ ] Integrate with fx lifecycle
- [ ] Export fx.Module
- [ ] Write unit tests
- [ ] Write integration tests

**Acceptance Criteria**:
- Can start HTTP server
- Each request gets `hyperion.Context` with all dependencies
- Trace context automatically extracted from headers
- Request/response logged automatically
- Panics recovered with proper logging
- Graceful shutdown on SIGTERM/SIGINT
- Metrics exposed for monitoring

**Module Usage**:
```go
import "github.com/mapoio/hyperion/adapter/gin"

fx.New(
    hyperion.CoreModule,
    viper.Module,
    zap.Module,
    gorm.Module,
    otel.Module,
    gin.Module,  // Web server
    // Application modules
    handler.Module,
    service.Module,
    repository.Module,
)
```

**Handler Example**:
```go
func (h *UserHandler) CreateUser(c *gin.Context) {
    // Get hyperion.Context from Gin context
    ctx := c.MustGet("hyperion.Context").(hyperion.Context)

    // Use logger, tracer, db from context
    _, span := ctx.Tracer().Start(ctx, "UserHandler.CreateUser")
    defer span.End()

    ctx.Logger().Info("creating user", "endpoint", "/users")

    // Business logic...
}
```

---

### Phase 7: Error Handling & Validation 🔄 (v2.3 - Optional)

**Status**: 🔄 **UNDER CONSIDERATION**

**Note**: In v2.0, error handling is NOT prescribed. Applications can:
- Use standard `fmt.Errorf` (recommended for simplicity)
- Use any error library they prefer (no lock-in)

However, for convenience, we may provide optional utilities.

#### 7.1 Error Utilities (Optional) - 3 Days

**Priority**: ⭐⭐⭐

**Optional Package**: `hyperion/errors` (utility package, not interface)

**Task List**:
- [ ] Implement error code constants
- [ ] Implement HTTP status code mapping
- [ ] Implement gRPC status code mapping
- [ ] Provide helper functions (Is, As, etc.)
- [ ] Write unit tests

**Note**: This is NOT an adapter. Just utility functions. Applications can ignore it.

#### 7.2 Validation Utilities (Optional) - 2 Days

**Priority**: ⭐⭐⭐

**Optional Package**: `hyperion/validation` (utility package)

**Task List**:
- [ ] Wrapper for go-playground/validator
- [ ] Helper functions for validation
- [ ] Write unit tests

---

### Phase 8: Additional Adapters 🔄 (v2.4+ - Future)

**Status**: 🔄 **FUTURE WORK**

#### 8.1 gRPC Server Adapter

**Package**: `adapter/grpc` or `hypergrpc/`

**Task List**:
- [ ] Implement gRPC server with interceptors
- [ ] Context injection
- [ ] Tracing integration
- [ ] Health check service

#### 8.2 HTTP Client Adapter

**Package**: `adapter/resty` or `hyperhttp/`

**Task List**:
- [ ] Implement HTTP client with tracing
- [ ] Automatic trace propagation
- [ ] Retry logic
- [ ] Circuit breaker (optional)

#### 8.3 Message Queue Adapters (v3.0)

**Packages**: `adapter/kafka`, `adapter/rabbitmq`, `adapter/nats`

**Task List**:
- [ ] Define MessageQueue interface in core
- [ ] Implement adapters for popular message brokers
- [ ] Tracing integration
- [ ] Consumer group support

---

## Release Schedule

### v2.0-alpha (Current)
- ✅ Core library with interfaces
- ✅ NoOp implementations
- ✅ Viper adapter
- ✅ Complete documentation

**Release Date**: October 2025

---

### v2.1 (Planned: December 2025)
- 🔜 Zap adapter (Logger)
- 🔜 GORM adapter (Database + UnitOfWork)
- 🔜 Complete Context implementation
- 🔜 Transaction management
- 🔜 Example applications

**Focus**: Production-ready logging and database access

---

### v2.2 (Planned: February 2026)
- 🔜 OpenTelemetry adapter (Tracer)
- 🔜 Ristretto adapter (Cache)
- 🔜 Redis adapter (Cache)
- 🔜 Metrics support (Prometheus)

**Focus**: Observability and caching

---

### v2.3 (Planned: April 2026)
- 🔜 Gin web adapter (HTTP server)
- 🔜 Optional error utilities
- 🔜 Optional validation utilities
- 🔜 Complete web application example

**Focus**: Web framework integration

---

### v2.4 (Planned: June 2026)
- 🔜 gRPC server adapter
- 🔜 HTTP client adapter
- 🔜 Additional utility packages

**Focus**: Microservices support

---

### v3.0 (Future: 2027+)
- 🔄 Message queue interfaces and adapters
- 🔄 Task scheduling (Asynq integration)
- 🔄 Service mesh integration
- 🔄 Advanced observability features

**Focus**: Distributed systems

---

## Adapter Development Guidelines

### Creating a New Adapter

When implementing a new adapter:

1. **Create Independent Module**:
   ```bash
   mkdir -p adapter/mylib
   cd adapter/mylib
   go mod init github.com/mapoio/hyperion/adapter/mylib
   ```

2. **Implement Interface**:
   ```go
   package mylib

   import "github.com/mapoio/hyperion"

   type myImpl struct {
       // implementation fields
   }

   func NewMyImpl(cfg hyperion.Config) hyperion.MyInterface {
       return &myImpl{ /* ... */ }
   }

   // Implement all interface methods...
   ```

3. **Export fx.Module**:
   ```go
   // module.go
   package mylib

   import (
       "go.uber.org/fx"
       "github.com/mapoio/hyperion"
   )

   var Module = fx.Module("hyperion.adapter.mylib",
       fx.Provide(
           fx.Annotate(
               NewMyImpl,
               fx.As(new(hyperion.MyInterface)),
           ),
       ),
   )
   ```

4. **Add to Workspace**:
   ```bash
   # From repo root
   go work use ./adapter/mylib
   ```

5. **Write Tests**:
   - Unit tests for all methods
   - Integration tests if needed
   - Ensure >= 80% coverage

6. **Document**:
   - Add README.md with usage examples
   - Add godoc comments
   - Update main documentation

---

## Testing Strategy

### Core Library Tests
- **Unit tests**: 100% coverage required for core interfaces
- **NoOp tests**: Verify NoOp implementations do nothing safely
- **Module tests**: Test fx module composition

### Adapter Tests
- **Unit tests**: >= 80% coverage for each adapter
- **Integration tests**: Test with real services (databases, Redis, etc.)
- **Benchmark tests**: Performance benchmarks for critical paths
- **Compatibility tests**: Test with multiple versions of underlying libraries

### Example Applications
- **End-to-end tests**: Full application tests
- **Load tests**: Performance under load
- **Chaos tests**: Behavior under failures

---

## Quality Checklist

Each adapter must pass before release:

- [ ] Interface fully implemented
- [ ] Unit tests pass with >= 80% coverage
- [ ] Integration tests pass (if applicable)
- [ ] golangci-lint passes with no errors
- [ ] godoc documentation complete
- [ ] README with usage examples
- [ ] Example code in repo runs successfully
- [ ] Performance benchmarks added
- [ ] CHANGELOG updated

---

## Development Workflow

### For Core Library Changes

1. Update interface in `hyperion/*.go`
2. Update NoOp implementation
3. Update all existing adapters
4. Update tests
5. Update documentation
6. Create PR with `core:` prefix

### For New Adapter

1. Create adapter directory: `adapter/{name}/`
2. Initialize independent module
3. Implement interface
4. Write tests (unit + integration)
5. Write documentation
6. Add to workspace
7. Update main docs
8. Create PR with `adapter/{name}:` prefix

### For Documentation

1. Update relevant docs in `docs/`
2. Ensure consistency across all docs
3. Update examples if needed
4. Create PR with `docs:` prefix

---

## Community Contributions

We welcome adapter contributions! To contribute:

1. **Propose**: Open an issue describing the adapter
2. **Discuss**: Get feedback from maintainers
3. **Implement**: Follow adapter development guidelines
4. **Test**: Ensure quality checklist passes
5. **Document**: Write clear documentation
6. **Submit**: Create PR for review

**Priority Adapters** (Community help wanted):
- `adapter/sqlx` - sqlx database adapter (alternative to GORM)
- `adapter/zerolog` - Zerolog logger adapter (alternative to Zap)
- `adapter/chi` - Chi router adapter (alternative to Gin)
- `adapter/nats` - NATS message queue adapter
- `adapter/temporal` - Temporal workflow adapter

---

## Milestones

### ✅ Milestone 1: Core Available (October 2025)
- ✅ Core library with all interfaces
- ✅ NoOp implementations
- ✅ Viper adapter
- ✅ Complete v2.0 documentation
- ✅ Quick start guide

**Status**: **COMPLETED**

---

### 🔜 Milestone 2: Database Ready (December 2025)
- 🔜 Zap adapter (Logger)
- 🔜 GORM adapter (Database)
- 🔜 Full transaction support
- 🔜 Example: Simple CRUD API

**Status**: **PLANNED**

---

### 🔜 Milestone 3: Observability Ready (February 2026)
- 🔜 OpenTelemetry adapter (Tracer)
- 🔜 Ristretto + Redis adapters (Cache)
- 🔜 Prometheus metrics integration
- 🔜 Example: Traced and monitored service

**Status**: **PLANNED**

---

### 🔜 Milestone 4: Web Framework Ready (April 2026)
- 🔜 Gin web adapter
- 🔜 Complete middleware suite
- 🔜 Error handling utilities
- 🔜 Example: Full-stack web application

**Status**: **PLANNED**

---

### 🔜 Milestone 5: Microservices Ready (June 2026)
- 🔜 gRPC server adapter
- 🔜 HTTP client adapter
- 🔜 Service discovery integration
- 🔜 Example: Multi-service architecture

**Status**: **PLANNED**

---

## Next Actions (v2.1 Development)

### Immediate Priorities

1. **Zap Adapter Implementation** (Week 1-2)
   - Implement zapLogger struct
   - Add JSON and Console encoders
   - Integration with Config
   - Tests and documentation

2. **GORM Adapter Implementation** (Week 3-4)
   - Implement gormDatabase struct
   - Implement UnitOfWork pattern
   - Transaction propagation
   - Tests with real databases

3. **Context Implementation** (Week 5)
   - Create production Context implementation
   - Integrate with Logger, Tracer, Database
   - Tests and documentation

4. **Example Application** (Week 6)
   - Simple CRUD API using all v2.1 features
   - Dockerfile and deployment guide
   - Performance benchmarks

### Long-term Goals

- **Ecosystem Growth**: Encourage community adapter contributions
- **Performance**: Continuous performance optimization
- **Stability**: Maintain backward compatibility
- **Documentation**: Keep docs comprehensive and up-to-date

---

## Total Time Estimates

| Version | Components | Estimated Time | Target Date |
|---------|-----------|----------------|-------------|
| v2.0-alpha | Core + Viper | ✅ COMPLETED | October 2025 |
| v2.1 | Zap + GORM | 6 weeks | December 2025 |
| v2.2 | OTel + Cache | 4 weeks | February 2026 |
| v2.3 | Gin Web | 3 weeks | April 2026 |
| v2.4 | gRPC + HTTP Client | 3 weeks | June 2026 |

**Total Development Time**: ~4 months (October 2025 - June 2026)

---

## Success Metrics

### Technical Metrics
- Core library: 0 dependencies (except fx)
- Adapter coverage: >= 80%
- Performance overhead: < 5% vs native libraries
- Documentation completeness: 100%

### Adoption Metrics (Goals)
- GitHub Stars: 1,000+ by v2.4
- Production users: 50+ by v2.4
- Community adapters: 10+ by end of 2026
- Tutorial completions: 500+ by v2.4

---

**Let's build the future of Go frameworks - with zero lock-in! 🚀**

**Last Updated**: October 2025
**Version**: 2.0 (Core-Adapter Architecture)
