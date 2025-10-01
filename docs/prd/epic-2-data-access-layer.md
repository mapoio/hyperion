# Epic 2: Data Access Layer

**Priority**: ⭐⭐⭐⭐⭐ (Highest)
**Estimated Duration**: 1.5 weeks
**Status**: Not Started
**Dependencies**: Epic 1 (Core Foundation)

---

## Overview

Implement database access with UnitOfWork pattern for declarative transaction management and cache abstraction for both in-memory and distributed caching.

---

## Goals

- Provide GORM-based database access with automatic tracing
- Enable declarative transaction management with automatic propagation
- Support both in-memory and distributed caching
- Ensure all data operations are automatically traced

---

## User Stories

### Story 2.1: Database + UnitOfWork (hyperdb)

**As a** framework user
**I want** declarative transaction management
**So that** I can write clean service code without manual transaction handling

**Acceptance Criteria**:
- [ ] Can connect to PostgreSQL, MySQL, and SQLite
- [ ] Transactions auto-commit on success and rollback on error
- [ ] `ctx.DB()` returns transaction handle when inside `WithTransaction()`
- [ ] All database operations automatically create spans
- [ ] Connection pool properly configured with health checks

**Tasks**:
- [ ] Define `DB` interface
- [ ] Implement `GormDB` with connection pool
- [ ] Implement database health check
- [ ] Define `UnitOfWork` interface
- [ ] Implement `GormUnitOfWork`
- [ ] Implement `WithTransaction()` method
- [ ] Implement `WithTransactionOptions()` for advanced control
- [ ] Implement `TracePlugin` for GORM (automatic span creation)
- [ ] Implement `ctx.WithDB()` internal method
- [ ] Write unit tests with mocked database (>80% coverage)
- [ ] Write integration tests with real PostgreSQL/MySQL
- [ ] Write godoc documentation

**Technical Details**:
```go
// Automatic transaction propagation
err := uow.WithTransaction(ctx, func(txCtx hyperctx.Context) error {
    // txCtx.DB() returns *gorm.DB with transaction
    if err := repo.Create(txCtx, user); err != nil {
        return err // Auto-rollback
    }
    return nil // Auto-commit
})
```

**Estimated**: 5 days

---

### Story 2.2: Cache Abstraction (hypercache)

**As a** framework user
**I want** unified cache interface for in-memory and distributed caching
**So that** I can easily switch between cache implementations

**Acceptance Criteria**:
- [ ] Can use in-memory cache (Ristretto)
- [ ] Can use Redis distributed cache
- [ ] All cache operations automatically create spans
- [ ] Support for TTL and batch operations
- [ ] Type-safe key-value operations

**Tasks**:
- [ ] Define `Cache` interface
- [ ] Implement `RistrettoCache` for in-memory
- [ ] Implement `RedisCache` for distributed cache
- [ ] Implement `WithTracing()` decorator
- [ ] Implement batch operations (MGet, MSet)
- [ ] Write unit tests with mocked cache (>80% coverage)
- [ ] Write integration tests with real Redis
- [ ] Write godoc documentation

**Technical Details**:
```go
// Unified interface
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    MGet(ctx context.Context, keys ...string) (map[string][]byte, error)
    MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error
}
```

**Estimated**: 2 days

---

## Milestone

**Deliverable**: Functional data access layer with database and cache support

**Demo Scenario**:
```go
// Service with declarative transactions
func (s *UserService) CreateUser(ctx hyperctx.Context, req *CreateUserRequest) error {
    return s.uow.WithTransaction(ctx, func(txCtx hyperctx.Context) error {
        // Create user
        user := &User{Username: req.Username}
        if err := s.userRepo.Create(txCtx, user); err != nil {
            return err
        }

        // Create profile (same transaction)
        profile := &UserProfile{UserID: user.ID}
        if err := s.profileRepo.Create(txCtx, profile); err != nil {
            return err
        }

        // Cache user
        if err := s.cache.Set(ctx, user.ID, userJSON, 1*time.Hour); err != nil {
            // Cache failure doesn't rollback transaction
            ctx.Logger().Warn("failed to cache user", "error", err)
        }

        return nil
    })
}
```

---

## Technical Notes

### Architecture Decisions

- **GORM for ORM**: Most popular Go ORM with plugin support
- **Ristretto for In-Memory**: High-performance concurrent cache from Dgraph
- **Redis for Distributed**: Industry standard with rich features

### Dependencies

- `gorm.io/gorm` - ORM
- `gorm.io/driver/postgres` - PostgreSQL driver
- `gorm.io/driver/mysql` - MySQL driver
- `gorm.io/driver/sqlite` - SQLite driver
- `github.com/dgraph-io/ristretto` - In-memory cache
- `github.com/redis/go-redis/v9` - Redis client

### Database Support Matrix

| Database | Driver | Status | Priority |
|----------|--------|--------|----------|
| PostgreSQL | gorm.io/driver/postgres | ✅ Supported | High |
| MySQL | gorm.io/driver/mysql | ✅ Supported | High |
| SQLite | gorm.io/driver/sqlite | ✅ Supported | Medium |
| SQL Server | gorm.io/driver/sqlserver | 🔄 Future | Low |

### Testing Strategy

- **Unit Tests**: Mock GORM database interface
- **Integration Tests**:
  - PostgreSQL in Docker container
  - MySQL in Docker container
  - Redis in Docker container
- **Performance Tests**: Benchmark transaction overhead

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| GORM nested transaction limitation | Medium | Document savepoint usage |
| Cache inconsistency | Medium | Implement cache-aside pattern |
| Connection pool exhaustion | High | Proper pool configuration and monitoring |

---

## Related Documentation

- [Architecture - hyperdb](../architecture.md#54-hyperdb---database--unitofwork)
- [Architecture - hypercache](../architecture.md#56-hypercache---cache-abstraction)
- [Tech Stack - Database](../architecture/tech-stack.md#orm-gorm)

---

**Last Updated**: 2025-01-XX
