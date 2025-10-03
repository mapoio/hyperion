# Hyperion Coding Standards

**Version**: 2.0
**Date**: October 2025
**Status**: Updated for Core-Adapter Architecture

This document defines the coding standards and best practices for Hyperion framework v2.0 development.

---

## General Principles

Hyperion follows the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) with additional framework-specific conventions for the v2.0 core-adapter architecture.

### Core Values

1. **Simplicity (KISS)**: Favor simple solutions over clever ones
2. **Do Not Repeat Yourself (DRY)**: Extract common patterns
3. **You Aren't Gonna Need It (YAGNI)**: Implement only what's needed now
4. **SOLID Principles**: Single responsibility, open-closed, Liskov substitution, interface segregation, dependency inversion

---

## Code Formatting

### Formatting Tools

```bash
# Format code
gofmt -w .
goimports -w .

# Or use the Makefile
make fmt
```

### Import Order

Imports should be organized in three groups:

1. Standard library packages
2. Third-party packages
3. Framework and application packages

**v2.0 Import Structure**:

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"

    // Third-party
    "go.uber.org/fx"

    // Framework core (NO pkg/ prefix in v2.0)
    "github.com/mapoio/hyperion"

    // Framework adapters
    "github.com/mapoio/hyperion/adapter/viper"
    "github.com/mapoio/hyperion/adapter/zap"

    // Application packages
    "github.com/your-app/internal/domain/user"
    "github.com/your-app/internal/service"
)
```

**Key Changes from v1.0**:
- ❌ No `pkg/` prefix: `github.com/mapoio/hyperion/pkg/hyperlog`
- ✅ Direct import: `github.com/mapoio/hyperion`
- ✅ Adapters: `github.com/mapoio/hyperion/adapter/{name}`

---

## Naming Conventions

### Package Names

**v2.0 Structure**:

- **Core library** (`hyperion/`): Single package, no prefix
  - All core interfaces in one package: `hyperion.Logger`, `hyperion.Tracer`, `hyperion.Database`
- **Adapters** (`adapter/*`): Named after underlying library
  - `adapter/viper`, `adapter/zap`, `adapter/otel`, `adapter/gorm`
- **Application packages**: Use descriptive domain names
  - `internal/domain/user`, `internal/service`, `internal/repository`

**Naming Rules**:
- Use lowercase, single-word package names
- No underscores or hyphens
- Package name should match directory name

### Interface Names

- Use simple, descriptive names without `Interface` or `I` prefix
- Examples: `Logger`, `Cache`, `Crypter` (not `ILogger`, `CacheInterface`)
- Exception: When interface and implementation have the same concept, add `er` suffix

**v2.0 Interface Design**:

```go
// Core library defines interfaces (hyperion/)
package hyperion

type Logger interface {
    Debug(msg string, fields ...any)
    Info(msg string, fields ...any)
    // ...
}

// Adapters implement interfaces (adapter/zap/)
package zap

import "github.com/mapoio/hyperion"

type zapLogger struct { /* implementation */ }

func (l *zapLogger) Info(msg string, fields ...any) { /* ... */ }

// Ensure interface compliance
var _ hyperion.Logger = (*zapLogger)(nil)
```

**Good Practices**:
```go
// Good - Simple interface names
type Logger interface { ... }
type Tracer interface { ... }
type Database interface { ... }

// Bad - Unnecessary prefixes/suffixes
type ILogger interface { ... }
type LoggerInterface interface { ... }
```

### Variable Names

- Use camelCase for local variables
- Use descriptive names, avoid single-letter variables except for:
  - Loop counters: `i`, `j`, `k`
  - Context: `ctx`
  - Error: `err`
  - Common abbreviations: `db`, `tx`, `cfg`

```go
// Good
userService := NewUserService(repo)
for i := 0; i < len(items); i++ { ... }

// Bad
us := NewUserService(repo)
for index := 0; index < len(items); index++ { ... }
```

### Constant Names

- Use MixedCaps, not SCREAMING_SNAKE_CASE
- Exceptions: Environment variables and exported constants

```go
// Good
const MaxRetries = 3
const DefaultTimeout = 30 * time.Second

// Bad
const MAX_RETRIES = 3
const default_timeout = 30 * time.Second
```

### Acronyms

- Keep acronyms consistent
- `ID` not `Id`, `HTTP` not `Http`, `URL` not `Url`

```go
// Good
type UserID string
func ServeHTTP(w http.ResponseWriter, r *http.Request)

// Bad
type UserId string
func ServeHttp(w http.ResponseWriter, r *http.Request)
```

---

## Error Handling

### Always Check Errors

```go
// Good
user, err := repo.FindByID(ctx, id)
if err != nil {
    return nil, fmt.Errorf("failed to find user: %w", err)
}

// Bad
user, _ := repo.FindByID(ctx, id)
```

### Error Wrapping

- Use `fmt.Errorf` with `%w` verb (standard library approach)
- Service layer: Return errors with business context
- Repository layer: Wrap database errors

**v2.0 Error Handling**:

```go
// Service layer - Add business context
func (s *UserService) CreateUser(ctx hyperion.Context, req CreateUserRequest) error {
    if err := s.userRepo.Save(ctx, user); err != nil {
        // Log with structured fields
        ctx.Logger().Error("failed to create user",
            "email", req.Email,
            "error", err,
        )
        // Wrap with business context
        return fmt.Errorf("failed to create user %s: %w", req.Email, err)
    }
    return nil
}

// Repository layer - Wrap database errors
func (r *UserRepository) Save(ctx hyperion.Context, user *User) error {
    if err := ctx.DB().Create(user).Error; err != nil {
        return fmt.Errorf("database create operation failed: %w", err)
    }
    return nil
}

// Transaction handling with UnitOfWork
func (s *UserService) RegisterUser(ctx hyperion.Context, req RegisterRequest) error {
    return s.uow.WithTransaction(ctx, func(txCtx hyperion.Context) error {
        // All operations within this function use txCtx.DB()
        // which automatically uses the transaction
        if err := s.userRepo.Save(txCtx, user); err != nil {
            return fmt.Errorf("failed to save user: %w", err)
        }
        if err := s.profileRepo.Create(txCtx, profile); err != nil {
            return fmt.Errorf("failed to create profile: %w", err)
        }
        return nil  // Commit on success
    })  // Automatic rollback on error
}
```

**Note**: Hyperion v2.0 does NOT mandate a specific error library. Use standard `fmt.Errorf` or integrate your own error handling library as needed.

### Error Messages

- Use lowercase for error messages (except proper nouns)
- Be specific about what failed
- Include relevant context

```go
// Good
return fmt.Errorf("failed to connect to database: %w", err)

// Bad
return fmt.Errorf("Error!") // Not specific, not lowercase
```

---

## Context Handling

### v2.0 Context Design

Hyperion v2.0 uses **accessor pattern** for the Context interface:

```go
// Core interface (hyperion/context.go)
type Context interface {
    context.Context  // Embedded standard context

    // Accessor methods - retrieve dependencies
    Logger() Logger
    DB() Executor
    Tracer() Tracer

    // Context management
    WithTimeout(timeout time.Duration) (Context, context.CancelFunc)
    WithCancel() (Context, context.CancelFunc)
}
```

### Context as First Parameter

Always pass `hyperion.Context` as the first parameter:

```go
// Service layer - Use Interceptor Pattern (recommended)
func (s *UserService) GetByID(ctx hyperion.Context, id string) (_ *User, err error) {
    // 3-Line Interceptor Pattern for automatic tracing, logging, and metrics
    ctx, end := ctx.UseIntercept("UserService", "GetByID")
    defer end(&err)

    ctx.Logger().Info("fetching user", "user_id", id)

    return s.userRepo.FindByID(ctx, id)
}

// Alternative: Manual tracing (only when you need fine-grained control)
func (s *UserService) GetByIDManual(ctx hyperion.Context, id string) (*User, error) {
    _, span := ctx.Tracer().Start(ctx, "UserService.GetByIDManual")
    defer span.End()

    ctx.Logger().Info("fetching user", "user_id", id)

    return s.userRepo.FindByID(ctx, id)
}

// Repository layer
func (r *UserRepository) FindByID(ctx hyperion.Context, id string) (*User, error) {
    var user User
    // ctx.DB() returns Executor (handles transaction propagation)
    if err := ctx.DB().First(&user, "id = ?", id).Error; err != nil {
        return nil, fmt.Errorf("failed to find user: %w", err)
    }
    return &user, nil
}
```

### Never Store Context

Never store context in structs:

```go
// Bad
type Service struct {
    ctx hyperion.Context
}

// Good
type Service struct {
    logger hyperion.Logger  // Store dependencies, not context
}

func (s *Service) DoWork(ctx hyperion.Context) error {
    // Use context from parameter, not stored field
    ctx.Logger().Info("working...")
    return nil
}
```

### Interceptor Pattern (3-Line Pattern)

**Recommended for Service Layer**: Use the 3-line interceptor pattern for automatic observability:

```go
// Good - 3-Line Interceptor Pattern
func (s *UserService) CreateUser(ctx hyperion.Context, req CreateUserRequest) (_ *User, err error) {
    // Line 1: Start interceptor chain
    ctx, end := ctx.UseIntercept("UserService", "CreateUser")
    // Line 2: Defer cleanup (MUST pass error pointer)
    defer end(&err)

    // Line 3+: Your business logic
    ctx.Logger().Info("creating user", "email", req.Email)

    // Record metrics
    counter := ctx.Meter().Counter("user.operations")
    counter.Add(ctx, 1, hyperion.String("operation", "create"))

    user, err := s.userRepo.Save(ctx, &User{...})
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    return user, nil
}

// Bad - Missing named error return
func (s *UserService) CreateUser(ctx hyperion.Context, req CreateUserRequest) (*User, error) {
    ctx, end := ctx.UseIntercept("UserService", "CreateUser")
    defer end(&err)  // ❌ 'err' is not defined as named return
    // ...
}

// Bad - Not passing error pointer
func (s *UserService) CreateUser(ctx hyperion.Context, req CreateUserRequest) (_ *User, err error) {
    ctx, end := ctx.UseIntercept("UserService", "CreateUser")
    defer end(nil)  // ❌ Should be end(&err)
    // ...
}
```

**Key Requirements**:
1. **Named Error Return**: Use `err error` or `_ *Type, err error` in return signature
2. **Error Pointer**: Pass `&err` to `end()` function for automatic error recording
3. **Service Layer Only**: Apply to service methods, not repository methods (too granular)

**What You Get Automatically**:
- ✅ Distributed trace span created and ended
- ✅ Method entry/exit logs with trace context
- ✅ Metrics recorded with duration and status
- ✅ Automatic error recording in traces
- ✅ TraceID/SpanID correlation across logs and metrics

---

## Function Design

### Function Length

- Keep functions focused and concise
- Cyclomatic complexity: ≤ 15
- Cognitive complexity: ≤ 20
- Nesting depth: ≤ 5 levels

### Function Parameters

- Limit to 3-4 parameters
- Use structs for functions requiring many parameters

```go
// Good
type CreateUserRequest struct {
    Username string
    Email    string
    Password string
    Metadata map[string]any
}

func CreateUser(ctx hyperion.Context, req CreateUserRequest) error

// Bad
func CreateUser(ctx hyperion.Context, username, email, password string, metadata map[string]any) error
```

### Return Values

- Return errors as the last return value
- Use named return values sparingly (only for documentation)

```go
// Good
func GetUser(id string) (*User, error)

// Acceptable (for documentation)
func GetUser(id string) (user *User, err error)
```

---

## Struct Design

### Struct Tags

- Use consistent tag ordering: `json`, `xml`, `gorm`, `validate`

```go
type User struct {
    ID       string `json:"id" gorm:"primaryKey"`
    Username string `json:"username" gorm:"uniqueIndex;not null" validate:"required,min=3"`
    Email    string `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
}
```

### Struct Initialization

- Use field names for clarity
- Zero values are valid where appropriate

```go
// Good
user := &User{
    Username: "john",
    Email:    "john@example.com",
}

// Bad
user := &User{"", "john", "john@example.com"} // Positional
```

---

## Concurrency

### Use sync.Mutex or sync.RWMutex

```go
type Cache struct {
    mu    sync.RWMutex
    items map[string]any
}

func (c *Cache) Get(key string) (any, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.items[key]
    return val, ok
}
```

### Avoid Data Races

- Run tests with `-race` flag: `go test -race ./...`
- Document non-thread-safe types

---

## Testing

### Test File Naming

- Test files: `*_test.go`
- Test functions: `TestFunctionName`
- Benchmark functions: `BenchmarkFunctionName`

### Table-Driven Tests

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 1, 2, 3},
        {"negative numbers", -1, -2, -3},
        {"mixed", 1, -1, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Test Helpers

- Use `t.Helper()` to mark helper functions

```go
func assertUser(t *testing.T, expected, actual *User) {
    t.Helper()
    assert.Equal(t, expected.ID, actual.ID)
    assert.Equal(t, expected.Username, actual.Username)
}
```

---

## Documentation

### Package Documentation

**v2.0 Documentation Standards**:

- Every package should have a `doc.go` file (or package comment in main file)
- Explain the package's purpose and usage

**Core Library Example**:

```go
// Package hyperion provides the core interfaces and types for building
// modular Go applications using dependency injection.
//
// Hyperion v2.0 follows a core-adapter pattern where this package defines
// ONLY interfaces with zero 3rd-party dependencies (except go.uber.org/fx).
// Concrete implementations are provided via adapters in the adapter/ directory.
//
// Basic usage with default NoOp implementations:
//
//     app := fx.New(
//         hyperion.CoreModule,  // Provides NoOp defaults
//         fx.Invoke(func(logger hyperion.Logger) {
//             logger.Info("hello world")  // NoOp - does nothing
//         }),
//     )
//
// Usage with real implementations:
//
//     app := fx.New(
//         hyperion.CoreWithoutDefaultsModule,  // No defaults
//         viper.Module,  // Real Config implementation
//         zap.Module,    // Real Logger implementation
//         fx.Invoke(func(logger hyperion.Logger) {
//             logger.Info("hello world")  // Real logging via Zap
//         }),
//     )
//
package hyperion
```

**Adapter Example**:

```go
// Package viper provides a Viper-based implementation of hyperion.Config
// and hyperion.ConfigWatcher interfaces.
//
// This adapter wraps github.com/spf13/viper to provide configuration
// management with support for multiple sources (files, environment variables,
// remote config) and hot-reloading via file watching.
//
// Example usage:
//
//     app := fx.New(
//         viper.Module,  // Provides Config and ConfigWatcher
//         fx.Invoke(func(cfg hyperion.Config) {
//             port := cfg.GetInt("server.port")
//             fmt.Printf("Port: %d\n", port)
//         }),
//     )
//
package viper
```

### Function Documentation

- Document exported functions, types, and constants
- Start comments with the name being documented
- End with a period

```go
// NewUserService creates a new UserService with the given dependencies.
// It returns an error if the repository is nil.
func NewUserService(repo UserRepository) (*UserService, error) {
    ...
}
```

---

## Code Quality Tools

### Linting

Use `golangci-lint` with the project configuration:

```bash
# Run linter
make lint

# Run linter with auto-fix
make lint-fix
```

### Enabled Linters

- `errcheck`: Check unchecked errors
- `gosimple`: Simplify code
- `govet`: Go vet examines Go source code
- `staticcheck`: Static analysis
- `revive`: Fast, configurable, extensible linter
- `gofmt/goimports`: Format checking
- `gocyclo`: Cyclomatic complexity
- `gosec`: Security checker
- Plus 20+ more (see `.golangci.yml`)

---

## Performance Guidelines

### Avoid Premature Optimization

- Write clear code first
- Profile before optimizing
- Benchmark to verify improvements

### Common Performance Pitfalls

1. **String Concatenation**: Use `strings.Builder` for multiple concatenations
2. **Defer in Loops**: Avoid defer in tight loops
3. **Map Allocation**: Pre-allocate maps when size is known
4. **Slice Growth**: Use `make([]T, 0, capacity)` to avoid reallocations

---

## Git Commit Standards

Hyperion follows the [AngularJS Commit Message Convention](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#commit).

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Test-related changes
- `build`: Build system changes
- `ci`: CI configuration changes
- `chore`: Other changes

### Scopes (v2.0)

**Core Library**:
- `core`: Core interfaces and types
- `context`: Context abstraction
- `interceptor`: Interceptor pattern and built-in interceptors
- `meter`: Meter interface and metrics
- `module`: Module system

**Adapters**:
- `adapter/viper`: Viper config adapter
- `adapter/zap`: Zap logger adapter
- `adapter/otel`: OpenTelemetry tracer and meter adapter
- `adapter/gorm`: GORM database adapter

**Documentation**:
- `docs`: General documentation
- `arch`: Architecture documentation

### Examples

```
feat(adapter/viper): add hot reload support

fix(core): correct Context interface accessor pattern

docs(arch): update architecture.md for v2.0

refactor(adapter/zap): simplify logger initialization

feat(adapter/otel): add OpenTelemetry tracer implementation
```

---

## Review Checklist

Before submitting a PR, ensure:

- [ ] Code follows Uber Go Style Guide
- [ ] All tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Code coverage ≥ 90%
- [ ] Documentation is updated
- [ ] Commit messages follow convention
- [ ] No unnecessary dependencies added

---

## v2.0 Migration Checklist

When migrating existing code to v2.0:

- [ ] Update import paths (remove `pkg/` prefix)
  - `github.com/mapoio/hyperion/pkg/hyperlog` → `github.com/mapoio/hyperion`
  - Add adapter imports: `github.com/mapoio/hyperion/adapter/viper`
- [ ] Replace bundled implementations with adapters
  - Remove direct Viper/Zap/GORM dependencies from main module
  - Import adapters instead
- [ ] Update context usage to accessor pattern
  - Replace `ctx.Info(...)` with `ctx.Logger().Info(...)`
  - Replace direct span creation with `ctx.Tracer().Start(...)`
- [ ] Update module composition
  - Use `hyperion.CoreModule` or `hyperion.CoreWithoutDefaultsModule`
  - Add adapter modules (e.g., `viper.Module`, `zap.Module`)
- [ ] Update error handling (if using custom error library)
  - Adapt to standard `fmt.Errorf` or keep your error library
- [ ] Review transaction management
  - Use `UnitOfWork.WithTransaction(...)` for declarative transactions
  - Access DB via `ctx.DB()` for automatic transaction propagation

---

**Last Updated**: October 2025
**Version**: 2.0 (Core-Adapter Architecture)
