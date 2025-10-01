# Hyperion Coding Standards

This document defines the coding standards and best practices for Hyperion framework development.

---

## General Principles

Hyperion follows the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) with additional framework-specific conventions.

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
3. Local packages (prefixed with `github.com/mapoio/hyperion`)

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"

    // Third-party
    "go.uber.org/fx"
    "go.uber.org/zap"

    // Local
    "github.com/mapoio/hyperion/pkg/hyperctx"
    "github.com/mapoio/hyperion/pkg/hyperlog"
)
```

---

## Naming Conventions

### Package Names

- Use lowercase, single-word package names
- Framework core components use `hyper*` prefix
- Example: `hyperlog`, `hyperdb`, `hypercache`

### Interface Names

- Use simple, descriptive names without `Interface` or `I` prefix
- Examples: `Logger`, `Cache`, `Crypter` (not `ILogger`, `CacheInterface`)
- Exception: When interface and implementation have the same concept, add `er` suffix

```go
// Good
type Logger interface { ... }
type ZapLogger struct { ... }

// Bad
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

- Use `fmt.Errorf` with `%w` verb (not `github.com/pkg/errors`)
- Service layer: Return `hypererror` with business context
- Repository layer: Wrap unexpected errors

```go
// Service layer
if err != nil {
    return nil, hypererror.Wrap(
        hypererror.CodeInternal,
        "failed to create user",
        err,
    ).WithField("email", email)
}

// Repository layer
if err := db.Create(user).Error; err != nil {
    return hypererror.InternalWrap("database operation failed", err)
}
```

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

### Context as First Parameter

Always pass `context.Context` (or `hyperctx.Context`) as the first parameter:

```go
func (s *UserService) GetByID(ctx hyperctx.Context, id string) (*User, error)
```

### Never Store Context

Never store context in structs:

```go
// Bad
type Service struct {
    ctx hyperctx.Context
}

// Good
func (s *Service) DoWork(ctx hyperctx.Context) error
```

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

func CreateUser(ctx hyperctx.Context, req CreateUserRequest) error

// Bad
func CreateUser(ctx hyperctx.Context, username, email, password string, metadata map[string]any) error
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

- Every package should have a `doc.go` file
- Explain the package's purpose and usage

```go
// Package hyperlog provides structured logging capabilities for Hyperion applications.
//
// It wraps go.uber.org/zap with a simplified interface and automatic trace context
// injection. All loggers support dynamic level adjustment and multiple output targets.
//
// Example usage:
//
//     logger, err := hyperlog.NewZapLogger(config)
//     logger.Info("server started", "port", 8080)
//
package hyperlog
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

### Examples

```
feat(hyperlog): add file rotation support

fix(hyperdb): correct transaction rollback handling

docs: update quick start guide

refactor(hypererror): simplify error wrapping logic
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

**Last Updated**: January 2025
