# Epic 1: Core Foundation

**Priority**: ⭐⭐⭐⭐⭐ (Highest)
**Estimated Duration**: 2 weeks
**Status**: In Progress (1/4 stories completed)
**Dependencies**: None
**Progress**: Story 1.1 ✅ Complete

---

## Overview

Implement the minimum viable framework core, providing essential infrastructure for configuration management, structured logging, context abstraction, and the framework entry point.

---

## Goals

- Establish foundational components that all other modules depend on
- Provide type-safe context management with integrated tracing
- Enable configuration hot reload for operational flexibility
- Support structured logging with dynamic level adjustment

---

## User Stories

### Story 1.1: Configuration Management (hyperconfig) ✅

**Status**: ✅ **DONE** (Completed: 2025-10-01)

**As a** framework user
**I want** to manage application configuration from multiple sources
**So that** I can easily configure my application in different environments

**Acceptance Criteria**:
- [x] Can load configuration from YAML/JSON files
- [x] Can override configuration with environment variables
- [x] Configuration file changes trigger callbacks for hot reload
- [x] Support for nested configuration keys
- [x] Type-safe configuration unmarshalling

**Tasks**:
- [x] Define `Provider` interface
- [x] Implement `ViperProvider` with file support
- [x] Implement environment variable override
- [x] Implement `Watch()` interface using fsnotify
- [x] Write unit tests (>80% coverage)
- [x] Write integration tests
- [x] Write godoc documentation

**Actual**: 1 day (3 days estimated)

**Implementation**:
- Package: `pkg/hyperconfig/`
- Files: config.go, viper.go, module.go, doc.go
- Tests: 96.8% coverage
- QA Gate: PASS
- Story Doc: `docs/stories/1.1.hyperconfig.md`

---

### Story 1.2: Structured Logging (hyperlog)

**As a** framework user
**I want** structured logging with dynamic level control
**So that** I can debug issues in production without restarting

**Acceptance Criteria**:
- [ ] Can output structured JSON logs
- [ ] Can dynamically adjust log level at runtime
- [ ] Can output to both stdout and files
- [ ] File auto-rotation works with configurable size/age limits
- [ ] Automatic trace context injection in logs

**Tasks**:
- [ ] Define `Logger` interface
- [ ] Implement `ZapLogger` based on go.uber.org/zap
- [ ] Implement JSON and Console encoders
- [ ] Implement `SetLevel()` for dynamic level adjustment
- [ ] Implement file output using lumberjack
- [ ] Integrate with hyperconfig for configuration
- [ ] Implement config hot reload callback
- [ ] Write unit tests (>80% coverage)
- [ ] Write godoc documentation

**Estimated**: 2 days

---

### Story 1.3: Context Abstraction (hyperctx)

**As a** framework user
**I want** type-safe access to request-scoped dependencies
**So that** I can write cleaner code without manual context value casting

**Acceptance Criteria**:
- [ ] Can create context with integrated tracing
- [ ] Can extract trace context from HTTP headers
- [ ] Can create child spans with automatic propagation
- [ ] Can type-safely access Logger and DB
- [ ] All `WithXxx()` methods return new instances (immutable design)
- [ ] Baggage support for cross-service data propagation

**Tasks**:
- [ ] Define `Context` interface extending context.Context
- [ ] Implement `hyperContext` struct
- [ ] Implement basic methods: `Logger()`, `TraceID()`, `SpanID()`
- [ ] Implement `StartSpan()` with OpenTelemetry integration
- [ ] Implement `RecordError()`, `SetAttributes()`, `AddEvent()`
- [ ] Implement `Baggage` support (WithBaggage, GetBaggage)
- [ ] Implement `WithTimeout()`, `WithCancel()`, `WithDeadline()`
- [ ] Implement `User` interface and default implementation
- [ ] Implement type-safe `ContextKey[T]`
- [ ] Implement `NewFromIncoming()` for trace extraction
- [ ] Write unit tests (>80% coverage)
- [ ] Write integration tests with OpenTelemetry
- [ ] Write godoc documentation

**Estimated**: 4 days

---

### Story 1.4: Framework Entry Point (hyperion)

**As a** framework user
**I want** simple entry point functions to bootstrap my application
**So that** I can quickly set up Web, gRPC, or full-stack applications

**Acceptance Criteria**:
- [ ] Can create basic application with `hyperion.Core()`
- [ ] Can create web application with `hyperion.Web()`
- [ ] Can create gRPC application with `hyperion.GRPC()`
- [ ] Can create full-stack with `hyperion.FullStack()`
- [ ] fx modules are correctly composed

**Tasks**:
- [ ] Implement `Core()` function (config + log + ctx)
- [ ] Implement `Web()` function (Core + hyperweb)
- [ ] Implement `GRPC()` function (Core + hypergrpc)
- [ ] Implement `FullStack()` function (Core + hyperweb + hypergrpc)
- [ ] Write example applications for each mode
- [ ] Write unit tests
- [ ] Write godoc documentation

**Estimated**: 1 day

---

## Milestone

**Deliverable**: Functional core framework with configuration, logging, and context management

**Demo Scenario**:
```go
package main

import (
    "github.com/mapoio/hyperion/pkg/hyperion"
    "go.uber.org/fx"
)

func main() {
    app := fx.New(
        hyperion.Core(),
        fx.Invoke(func(ctx hyperctx.Context) {
            ctx.Logger().Info("Hello, Hyperion!")
        }),
    )
    app.Run()
}
```

---

## Technical Notes

### Architecture Decisions

- **Viper for Configuration**: Industry standard, supports multiple formats and remote sources
- **Zap for Logging**: Fastest structured logger in Go, production-proven at Uber
- **OpenTelemetry Integration**: Trace context directly in hyperctx.Context for zero-intrusion tracing

### Dependencies

- `go.uber.org/fx` - Dependency injection
- `go.uber.org/zap` - Structured logging
- `spf13/viper` - Configuration management
- `go.opentelemetry.io/otel` - Distributed tracing
- `gopkg.in/natefinch/lumberjack.v2` - Log rotation

### Testing Strategy

- Unit tests with mocks for interfaces
- Integration tests with real file watching
- Benchmark tests for performance validation

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| OpenTelemetry API complexity | Medium | Wrap in simple hyperctx methods |
| Config hot reload edge cases | Low | Extensive integration testing |
| Performance overhead of tracing | Medium | Benchmark and optimize hot paths |

---

## Related Documentation

- [Architecture Design](../architecture.md#5-component-details)
- [Coding Standards](../architecture/coding-standards.md)
- [Tech Stack Rationale](../architecture/tech-stack.md)

---

**Last Updated**: 2025-01-XX
