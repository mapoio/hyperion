# Epic 2: Essential Adapters (v2.1)

**Version**: 2.1
**Status**: 🔜 **PLANNED** (December 2025)
**Duration**: 6 weeks
**Priority**: ⭐⭐⭐⭐⭐

---

## Overview

Implement production-ready adapters for **Logger** (Zap) and **Database** (GORM), providing enterprise-grade logging and data access capabilities while maintaining zero lock-in.

---

## Goals

### Primary Goals
1. Production-ready structured logging with Zap adapter
2. Full-featured database access with GORM adapter
3. Complete transaction management (UnitOfWork pattern)
4. Production Context implementation

### Success Criteria
- [ ] Zap adapter passes all Logger interface tests
- [ ] GORM adapter passes all Database interface tests
- [ ] Transaction propagation works correctly
- [ ] Performance overhead < 5% vs native libraries
- [ ] Example CRUD application demonstrates all features

---

## Deliverables

### 1. Zap Logger Adapter 🔜

**Package**: `adapter/zap/`

**Scope**:
- Implement `hyperion.Logger` interface using Zap
- Support JSON and Console encoders
- Dynamic log level adjustment
- File output with rotation (lumberjack integration)
- Configuration integration via `hyperion.Config`

**Interface Implementation**:
```go
type zapLogger struct {
    *zap.SugaredLogger
}

func (l *zapLogger) Debug(msg string, fields ...any) {
    l.Debugw(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...any) {
    l.Infow(msg, fields...)
}

// ... implement all Logger methods
```

**Module Export**:
```go
var Module = fx.Module("hyperion.adapter.zap",
    fx.Provide(
        fx.Annotate(
            NewZapLogger,
            fx.As(new(hyperion.Logger)),
        ),
    ),
)
```

**Configuration Example**:
```yaml
log:
  level: info              # debug, info, warn, error
  encoding: json          # json or console
  output: stdout          # stdout, stderr, or file path
  file:
    path: /var/log/app.log
    max_size: 100         # MB
    max_backups: 3
    max_age: 7            # days
```

**Tasks**:
- [ ] Implement zapLogger struct (2 days)
- [ ] Add encoder support (JSON/Console) (1 day)
- [ ] Integrate with Config (1 day)
- [ ] Add file rotation (lumberjack) (1 day)
- [ ] Write unit tests (>80% coverage) (1 day)
- [ ] Write integration tests (1 day)
- [ ] Documentation and examples (1 day)

**Timeline**: 5 working days

---

### 2. GORM Database Adapter 🔜

**Package**: `adapter/gorm/`

**Scope**:
- Implement `hyperion.Database` interface
- Implement `hyperion.Executor` interface (GORM wrapper)
- Implement `hyperion.UnitOfWork` interface
- Transaction propagation via Context
- GORM plugin for tracing integration
- Support PostgreSQL, MySQL, SQLite

**Interface Implementation**:
```go
type gormDatabase struct {
    db *gorm.DB
}

func (d *gormDatabase) DB() hyperion.Executor {
    return &gormExecutor{db: d.db}
}

type gormUnitOfWork struct {
    db *gorm.DB
}

func (u *gormUnitOfWork) WithTransaction(ctx hyperion.Context, fn func(txCtx hyperion.Context) error) error {
    return u.db.Transaction(func(tx *gorm.DB) error {
        txCtx := ctx.WithDB(tx)  // Internal method
        return fn(txCtx)
    })
}
```

**Module Export**:
```go
var Module = fx.Module("hyperion.adapter.gorm",
    fx.Provide(
        fx.Annotate(
            NewGormDatabase,
            fx.As(new(hyperion.Database)),
        ),
        fx.Annotate(
            NewGormUnitOfWork,
            fx.As(new(hyperion.UnitOfWork)),
        ),
    ),
)
```

**Configuration Example**:
```yaml
database:
  driver: postgres
  host: localhost
  port: 5432
  username: dbuser
  password: dbpass
  database: mydb
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
```

**Transaction Example**:
```go
func (s *UserService) RegisterUser(ctx hyperion.Context, req RegisterRequest) error {
    return s.uow.WithTransaction(ctx, func(txCtx hyperion.Context) error {
        // txCtx.DB() automatically uses transaction
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

**Tasks**:
- [ ] Implement gormDatabase struct (1 day)
- [ ] Implement gormExecutor (GORM wrapper) (2 days)
- [ ] Implement gormUnitOfWork (2 days)
- [ ] Add transaction propagation (1 day)
- [ ] Implement TracePlugin (GORM plugin) (1 day)
- [ ] Add support for PostgreSQL, MySQL, SQLite (1 day)
- [ ] Write unit tests (>80% coverage) (2 days)
- [ ] Write integration tests (real databases) (2 days)
- [ ] Documentation and examples (1 day)

**Timeline**: 10 working days

---

### 3. Production Context Implementation 🔜

**Package**: `hyperion/context_impl.go` (internal implementation)

**Scope**:
- Full `hyperion.Context` implementation
- Logger, Tracer, DB integration
- Transaction-aware DB accessor
- Immutable context pattern

**Implementation**:
```go
type contextImpl struct {
    context.Context
    logger   Logger
    tracer   Tracer
    db       Executor
}

func (c *contextImpl) Logger() Logger {
    return c.logger
}

func (c *contextImpl) DB() Executor {
    return c.db  // Returns transaction if in transaction context
}

func (c *contextImpl) Tracer() Tracer {
    return c.tracer
}
```

**Tasks**:
- [ ] Implement contextImpl struct (1 day)
- [ ] Add accessor methods (1 day)
- [ ] Add WithDB (internal method) (1 day)
- [ ] Write unit tests (1 day)
- [ ] Integration with web framework (planned for v2.3)

**Timeline**: 3 working days

---

### 4. Example CRUD Application 🔜

**Package**: `examples/crud-api/`

**Scope**:
- Simple user management API
- Demonstrates Zap logging
- Demonstrates GORM database access
- Demonstrates transaction management
- Dockerfile and deployment guide

**Features**:
- Create user (POST /users)
- Get user by ID (GET /users/:id)
- List users (GET /users)
- Update user (PUT /users/:id)
- Delete user (DELETE /users/:id)

**Technology Stack**:
```
- hyperion.CoreModule
- adapter/viper (Config)
- adapter/zap (Logger)
- adapter/gorm (Database)
- Standard http package (no web framework yet, coming in v2.3)
```

**Tasks**:
- [ ] Set up project structure (1 day)
- [ ] Implement user repository (1 day)
- [ ] Implement user service (1 day)
- [ ] Implement HTTP handlers (1 day)
- [ ] Add Dockerfile (0.5 day)
- [ ] Write deployment guide (0.5 day)
- [ ] Add README with usage instructions (1 day)

**Timeline**: 4 working days

---

### 5. Integration Testing 🔜

**Scope**:
- Integration tests for Zap adapter
- Integration tests for GORM adapter (real databases)
- End-to-end tests for example application

**Test Coverage**:
- [ ] Zap adapter with real configuration
- [ ] GORM with PostgreSQL (Docker)
- [ ] GORM with MySQL (Docker)
- [ ] GORM with SQLite (in-memory)
- [ ] Transaction commit scenarios
- [ ] Transaction rollback scenarios
- [ ] Example CRUD API endpoints

**Tasks**:
- [ ] Set up test databases (Docker Compose) (1 day)
- [ ] Write Zap integration tests (1 day)
- [ ] Write GORM integration tests (2 days)
- [ ] Write E2E tests for example app (1 day)

**Timeline**: 3 working days

---

## Implementation Timeline

### Week 1-2: Zap Adapter (2 weeks)
- Days 1-2: Core implementation
- Days 3-4: Encoder and file output
- Days 5-7: Testing and documentation
- Day 8: Code review and refinement

### Week 3-4: GORM Adapter (2.5 weeks)
- Days 1-3: Database and Executor implementation
- Days 4-6: UnitOfWork and transactions
- Days 7-9: Testing (unit + integration)
- Days 10-11: Documentation and examples

### Week 5: Context & Example App (1 week)
- Days 1-3: Context implementation
- Days 4-6: Example CRUD application
- Day 7: Documentation

### Week 6: Integration & Release (0.5 week)
- Days 1-3: Integration testing
- Day 4: Final testing and bug fixes
- Day 5: Release v2.1

**Total**: 6 weeks

---

## Technical Challenges

### Challenge 1: Transaction Propagation
**Problem**: How to pass transaction handle through Context
**Solution**: Internal `WithDB()` method on Context, not exposed in interface
**Status**: Planned

### Challenge 2: GORM Version Compatibility
**Problem**: Support GORM v1 and v2
**Solution**: Focus on GORM v2 only (v1 is legacy)
**Status**: Decided

### Challenge 3: Performance Overhead
**Problem**: Ensure minimal overhead vs native GORM
**Solution**: Thin wrapper, benchmark critical paths
**Status**: Planned

---

## Success Metrics

### Code Metrics
- Test coverage: >= 80% for both adapters
- Performance overhead: < 5% vs native libraries
- Lines of code: ~1,500 LOC (zap), ~2,000 LOC (gorm)

### Quality Metrics
- golangci-lint: Zero errors
- Integration tests: All passing
- Example app: Fully functional

### Community Metrics (Target)
- Production users: 10+ by end of v2.1
- GitHub stars: 200+
- Community feedback: Positive

---

## Next Epic

👉 **[Epic 3: Observability Stack](epic-3-observability-stack.md)** (v2.2 - Planned)

**Focus**: OpenTelemetry tracing, Ristretto/Redis caching, Prometheus metrics

**Timeline**: February 2026

---

## Related Documentation

- [Epic 1: Core Foundation](epic-1-core-foundation.md)
- [Architecture Overview](../architecture.md)
- [Implementation Plan](../implementation-plan.md)

---

**Epic Status**: 🔜 **PLANNED** (December 2025)

**Last Updated**: October 2025
**Version**: 2.1 Planning
