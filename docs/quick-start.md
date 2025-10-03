# Hyperion v2.0 Quick Start Guide

**Version**: 2.0
**Last Updated**: October 2025

Build your first Hyperion v2.0 application in 15 minutes using the Core-Adapter architecture.

---

## What You'll Build

A RESTful user management API with:
- Configuration management (Viper adapter)
- Structured logging with automatic correlation (Zap adapter)
- Distributed tracing (OpenTelemetry-compatible)
- Metrics collection with trace correlation (OpenTelemetry-compatible)
- Automatic observability via Interceptor pattern
- Database access with transactions (GORM adapter)
- Dependency injection (fx)
- Clean architecture layers

---

## Prerequisites

- **Go 1.24+** (v2.0 requires Go 1.24 for workspace support)
- Basic Go programming knowledge
- (Optional) Docker - for running PostgreSQL

---

## Understanding Hyperion v2.0

Before we start, understand the key concepts:

### Core-Adapter Pattern

```
┌─────────────────────────────────────┐
│   Your Application                   │
│   (uses interfaces only)             │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   hyperion (Core Library)            │
│   • Logger interface                 │
│   • Database interface               │
│   • Config interface                 │
│   • ZERO 3rd-party deps (except fx)  │
└──────────────┬──────────────────────┘
               │
    ┌──────────┴──────────┬─────────────┐
    │                     │             │
┌───▼──────┐    ┌────────▼───┐  ┌──────▼─────┐
│ adapter/  │    │ adapter/   │  │ adapter/   │
│ viper     │    │ zap        │  │ gorm       │
│ (Config)  │    │ (Logger)   │  │ (Database) │
└───────────┘    └────────────┘  └────────────┘
```

**Benefits**:
- Zero lock-in: Core library has NO dependency on Viper, Zap, or GORM
- Swap implementations: Replace any adapter without touching core code
- Test-friendly: Use NoOp implementations or mocks

---

## Step 1: Create Project

```bash
mkdir my-hyperion-app
cd my-hyperion-app
go mod init github.com/yourusername/my-hyperion-app
```

---

## Step 2: Install Hyperion v2.0

```bash
# Core library (interfaces only)
go get github.com/mapoio/hyperion

# Viper adapter (configuration)
go get github.com/mapoio/hyperion/adapter/viper

# Note: In v2.0, you explicitly choose which adapters to use
```

**What's different from v1.0?**
- No `pkg/` prefix in import paths
- Adapters are separate modules
- You only install what you need

---

## Step 3: Create Configuration File

```bash
mkdir configs
```

Create `configs/config.yaml`:

```yaml
# Application settings
app:
  name: my-hyperion-app
  env: development

# Server configuration (for future web module)
server:
  host: "0.0.0.0"
  port: 8080

# Database configuration (for future GORM adapter)
database:
  driver: postgres
  host: localhost
  port: 5432
  username: test
  password: test
  database: testdb
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 5
```

---

## Step 4: Create Domain Model

Create `internal/domain/user.go`:

```go
package domain

import "time"

// User represents a user entity in the domain
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

**Note**: Domain models are pure Go structs with NO framework dependencies.

---

## Step 5: Create Repository Interface

Create `internal/repository/user_repository.go`:

```go
package repository

import (
	"github.com/mapoio/hyperion"
	"github.com/yourusername/my-hyperion-app/internal/domain"
)

// UserRepository defines data access operations for users
// Note: Uses hyperion.Context (not pkg/hyperctx in v2.0)
type UserRepository interface {
	Create(ctx hyperion.Context, user *domain.User) error
	FindByID(ctx hyperion.Context, id string) (*domain.User, error)
	FindAll(ctx hyperion.Context) ([]*domain.User, error)
}
```

**v2.0 Changes**:
- Import: `github.com/mapoio/hyperion` (no `pkg/` prefix)
- Context: `hyperion.Context` interface with accessor pattern

---

## Step 6: Implement Repository (In-Memory for Demo)

Create `internal/repository/user_memory_repository.go`:

```go
package repository

import (
	"fmt"
	"sync"

	"github.com/mapoio/hyperion"
	"github.com/yourusername/my-hyperion-app/internal/domain"
)

// memoryUserRepository is an in-memory implementation for demo purposes
// In production, use GORM adapter: github.com/mapoio/hyperion/adapter/gorm
type memoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

// NewMemoryUserRepository creates a new in-memory user repository
func NewMemoryUserRepository() UserRepository {
	return &memoryUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (r *memoryUserRepository) Create(ctx hyperion.Context, user *domain.User) error {
	// v2.0 Accessor Pattern: ctx.Logger() returns the logger
	ctx.Logger().Info("creating user", "username", user.Username)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return fmt.Errorf("user with id %s already exists", user.ID)
	}

	r.users[user.ID] = user
	return nil
}

func (r *memoryUserRepository) FindByID(ctx hyperion.Context, id string) (*domain.User, error) {
	ctx.Logger().Debug("finding user by id", "user_id", id)

	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", id)
	}

	return user, nil
}

func (r *memoryUserRepository) FindAll(ctx hyperion.Context) ([]*domain.User, error) {
	ctx.Logger().Debug("finding all users")

	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*domain.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return users, nil
}
```

Create `internal/repository/module.go`:

```go
package repository

import "go.uber.org/fx"

// Module provides repository implementations
var Module = fx.Module("repository",
	fx.Provide(
		NewMemoryUserRepository,
	),
)
```

**v2.0 Key Points**:
- Use `ctx.Logger()` instead of `ctx.Info()` (accessor pattern)
- Context automatically provides logger from dependency injection
- Repository is just a regular Go interface

---

## Step 7: Create Service Layer

Create `internal/service/user_service.go`:

```go
package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mapoio/hyperion"
	"github.com/yourusername/my-hyperion-app/internal/domain"
	"github.com/yourusername/my-hyperion-app/internal/repository"
)

// UserService defines business operations for users
type UserService interface {
	CreateUser(ctx hyperion.Context, username, email string) (*domain.User, error)
	GetUser(ctx hyperion.Context, id string) (*domain.User, error)
	ListUsers(ctx hyperion.Context) ([]*domain.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new user service
// Dependencies are automatically injected by fx
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) CreateUser(ctx hyperion.Context, username, email string) (_ *domain.User, err error) {
	// v2.0: 3-Line Interceptor Pattern for automatic tracing, logging, and metrics
	ctx, end := ctx.UseIntercept("UserService", "CreateUser")
	defer end(&err)

	// Log with structured fields (automatically includes trace_id and span_id)
	ctx.Logger().Info("creating new user",
		"username", username,
		"email", email,
	)

	// Record request metric (automatically includes exemplar linking to trace)
	counter := ctx.Meter().Counter("user.requests",
		hyperion.WithMetricDescription("Total user service requests"),
		hyperion.WithMetricUnit("1"),
	)
	counter.Add(ctx, 1,
		hyperion.String("method", "CreateUser"),
		hyperion.String("status", "started"),
	)

	user := &domain.User{
		ID:        uuid.New().String(),
		Username:  username,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		ctx.Logger().Error("failed to create user", "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *userService) GetUser(ctx hyperion.Context, id string) (_ *domain.User, err error) {
	// 3-Line Interceptor Pattern
	ctx, end := ctx.UseIntercept("UserService", "GetUser")
	defer end(&err)

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		ctx.Logger().Error("failed to get user", "user_id", id, "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *userService) ListUsers(ctx hyperion.Context) (_ []*domain.User, err error) {
	// 3-Line Interceptor Pattern
	ctx, end := ctx.UseIntercept("UserService", "ListUsers")
	defer end(&err)

	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		ctx.Logger().Error("failed to list users", "error", err)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	ctx.Logger().Info("listed users", "count", len(users))
	return users, nil
}
```

Create `internal/service/module.go`:

```go
package service

import "go.uber.org/fx"

// Module provides service implementations
var Module = fx.Module("service",
	fx.Provide(
		NewUserService,
	),
)
```

**v2.0 Interceptor Pattern**:
- `ctx.UseIntercept("Service", "Method")` returns enhanced context and cleanup function
- Automatically creates trace spans, logs method entry/exit, and records metrics
- 3-line pattern: `ctx, end := ctx.UseIntercept(...); defer end(&err)`
- Named error return required: `(err error)` or `(_ *User, err error)`
- Works with NoOp implementations by default (zero overhead)

---

## Step 8: Create a Simple CLI

Create `cmd/server/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/fx"

	"github.com/mapoio/hyperion"
	"github.com/mapoio/hyperion/adapter/viper"
	"github.com/yourusername/my-hyperion-app/internal/repository"
	"github.com/yourusername/my-hyperion-app/internal/service"
)

func main() {
	app := fx.New(
		// v2.0 Core Module (provides NoOp defaults for all interfaces)
		hyperion.CoreModule,

		// v2.0 Interceptor Modules (enable automatic tracing, logging, and metrics)
		hyperion.AllInterceptorsModule, // Includes TracingInterceptor + LoggingInterceptor

		// v2.0 Viper Adapter (replaces NoOp Config with real implementation)
		viper.Module,

		// Application modules
		repository.Module,
		service.Module,

		// Demo: Create some users on startup
		fx.Invoke(demoCreateUsers),
	)

	app.Run()
}

// demoCreateUsers demonstrates using the service layer
func demoCreateUsers(
	lc fx.Lifecycle,
	userService service.UserService,
	cfg hyperion.Config,
	logger hyperion.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Create a hyperion.Context (in real apps, this comes from HTTP middleware)
			// For now, we'll simulate it with a simple implementation
			appCtx := newSimpleContext(ctx, logger)

			appName := cfg.GetString("app.name")
			logger.Info("application started", "app_name", appName)

			// Create demo users
			user1, err := userService.CreateUser(appCtx, "alice", "alice@example.com")
			if err != nil {
				return err
			}
			logger.Info("created demo user", "user_id", user1.ID, "username", user1.Username)

			user2, err := userService.CreateUser(appCtx, "bob", "bob@example.com")
			if err != nil {
				return err
			}
			logger.Info("created demo user", "user_id", user2.ID, "username", user2.Username)

			// List all users
			users, err := userService.ListUsers(appCtx)
			if err != nil {
				return err
			}

			logger.Info("total users created", "count", len(users))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("Application stopped gracefully")
			return nil
		},
	})
}

// simpleContext is a minimal hyperion.Context implementation for demo purposes
// In production, use the context implementation from hyperweb or create your own
type simpleContext struct {
	context.Context
	logger hyperion.Logger
	tracer hyperion.Tracer
}

func newSimpleContext(ctx context.Context, logger hyperion.Logger) hyperion.Context {
	return &simpleContext{
		Context: ctx,
		logger:  logger,
		tracer:  &noopTracer{}, // Use NoOp tracer for demo
	}
}

func (c *simpleContext) Logger() hyperion.Logger {
	return c.logger
}

func (c *simpleContext) DB() hyperion.Executor {
	// Not used in this demo (no database yet)
	return nil
}

func (c *simpleContext) Tracer() hyperion.Tracer {
	return c.tracer
}

func (c *simpleContext) Meter() hyperion.Meter {
	return &noopMeter{} // Use NoOp meter for demo
}

func (c *simpleContext) UseIntercept(parts ...any) (hyperion.Context, func(err *error)) {
	// For demo purposes, return self without interceptors
	return c, func(err *error) {}
}

func (c *simpleContext) WithTimeout(timeout time.Duration) (hyperion.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.Context, timeout)
	return &simpleContext{Context: ctx, logger: c.logger, tracer: c.tracer}, cancel
}

func (c *simpleContext) WithCancel() (hyperion.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c.Context)
	return &simpleContext{Context: ctx, logger: c.logger, tracer: c.tracer}, cancel
}

func (c *simpleContext) WithDeadline(deadline time.Time) (hyperion.Context, context.CancelFunc) {
	ctx, cancel := context.WithDeadline(c.Context, deadline)
	return &simpleContext{Context: ctx, logger: c.logger, tracer: c.tracer}, cancel
}

// noopTracer is a simple no-op tracer for demo
type noopTracer struct{}

func (t *noopTracer) Start(ctx context.Context, spanName string, opts ...any) (context.Context, hyperion.Span) {
	return ctx, &noopSpan{}
}

type noopSpan struct{}

func (s *noopSpan) End(options ...any)                      {}
func (s *noopSpan) AddEvent(name string, options ...any)    {}
func (s *noopSpan) RecordError(err error, options ...any)   {}
func (s *noopSpan) SetStatus(code hyperion.StatusCode, description string) {}
func (s *noopSpan) SetAttributes(attributes ...any)         {}
func (s *noopSpan) SetName(name string)                     {}
func (s *noopSpan) TracerProvider() any                     { return nil }

// noopMeter is a simple no-op meter for demo
type noopMeter struct{}

func (m *noopMeter) Counter(name string, opts ...hyperion.MetricOption) hyperion.Counter {
	return &noopCounter{}
}
func (m *noopMeter) Histogram(name string, opts ...hyperion.MetricOption) hyperion.Histogram {
	return &noopHistogram{}
}
func (m *noopMeter) Gauge(name string, opts ...hyperion.MetricOption) hyperion.Gauge {
	return &noopGauge{}
}
func (m *noopMeter) UpDownCounter(name string, opts ...hyperion.MetricOption) hyperion.UpDownCounter {
	return &noopUpDownCounter{}
}

type noopCounter struct{}
func (c *noopCounter) Add(ctx context.Context, value int64, attrs ...hyperion.Attribute) {}

type noopHistogram struct{}
func (h *noopHistogram) Record(ctx context.Context, value float64, attrs ...hyperion.Attribute) {}

type noopGauge struct{}
func (g *noopGauge) Record(ctx context.Context, value float64, attrs ...hyperion.Attribute) {}

type noopUpDownCounter struct{}
func (u *noopUpDownCounter) Add(ctx context.Context, value int64, attrs ...hyperion.Attribute) {}
```

**v2.0 Module Composition**:
```go
fx.New(
    hyperion.CoreModule,           // Provides NoOp implementations
    hyperion.AllInterceptorsModule, // Enables automatic tracing, logging, and metrics
    viper.Module,                  // Overrides Config with Viper
    // zap.Module,                 // Would override Logger with Zap (available)
    // gorm.Module,                // Would override Database with GORM (available)
    repository.Module,
    service.Module,
)
```

**Key Benefits**:
- **3-Line Interceptor Pattern**: Automatically creates spans, logs, and records metrics
- **Automatic Correlation**: TraceID and SpanID shared across logs, traces, and metrics
- **Zero Code Changes**: Switch from NoOp to real implementations without modifying business logic

---

## Step 9: Configure Environment Variables

The Viper adapter loads configuration from `configs/config.yaml` by default.

Set environment variable to specify config file:

```bash
export HYPERION_CONFIG_PATH=configs/config.yaml
```

Or rely on default search paths:
- `./configs/config.yaml`
- `./config.yaml`
- `/etc/hyperion/config.yaml`

---

## Step 10: Run the Application

```bash
go mod tidy  # Download dependencies
go run cmd/server/main.go
```

Expected output (with NoOp logger - logs go nowhere by default):

```
[Fx] PROVIDE	hyperion.Logger <= github.com/mapoio/hyperion.glob..func1()
[Fx] PROVIDE	hyperion.Config <= github.com/mapoio/hyperion/adapter/viper.NewProviderFromEnv()
[Fx] PROVIDE	repository.UserRepository <= github.com/yourusername/my-hyperion-app/internal/repository.NewMemoryUserRepository()
[Fx] PROVIDE	service.UserService <= github.com/yourusername/my-hyperion-app/internal/service.NewUserService()
[Fx] RUNNING
```

**Note**: You won't see application logs because we're using NoOp logger by default. To see logs, you would install the Zap adapter (coming in v2.1).

---

## Understanding What Just Happened

### 1. Core Module Provided NoOp Defaults

```go
hyperion.CoreModule  // Provides:
// - NoOp Logger (logs nothing)
// - NoOp Tracer (traces nothing)
// - NoOp Database (does nothing)
// - NoOp Config (returns empty values)
```

### 2. Viper Adapter Replaced Config

```go
viper.Module  // Replaces NoOp Config with real Viper-based config
```

fx automatically resolves that `viper.Module` provides `hyperion.Config`, so it uses that instead of NoOp.

### 3. Your Code Uses Interfaces Only

Your service layer only knows about `hyperion.Context` and its accessor methods. It has **zero knowledge** of Viper, Zap, or any concrete implementation.

---

## Next Steps: Add Real Logger (Zap Adapter)

When the Zap adapter is available (v2.1), you would simply:

```go
import "github.com/mapoio/hyperion/adapter/zap"

fx.New(
    hyperion.CoreModule,
    viper.Module,
    zap.Module,           // Add this line
    repository.Module,
    service.Module,
)
```

Your application code **doesn't change at all**. The logger used by `ctx.Logger()` automatically becomes a real Zap logger.

---

## Next Steps: Add Database (GORM Adapter)

When the GORM adapter is available (v2.1), you would:

1. Install adapter:
   ```bash
   go get github.com/mapoio/hyperion/adapter/gorm
   ```

2. Add to application:
   ```go
   import "github.com/mapoio/hyperion/adapter/gorm"

   fx.New(
       hyperion.CoreModule,
       viper.Module,
       gorm.Module,      // Add this line
       repository.Module,
       service.Module,
   )
   ```

3. Replace in-memory repository with GORM-based implementation:
   ```go
   func (r *gormUserRepository) Create(ctx hyperion.Context, user *domain.User) error {
       // ctx.DB() now returns a real GORM database
       return ctx.DB().Create(user).Error
   }
   ```

---

## Project Structure

```
my-hyperion-app/
├── cmd/
│   └── server/
│       └── main.go              # Application entry (fx.New)
├── configs/
│   └── config.yaml              # Configuration file (Viper reads this)
├── internal/
│   ├── domain/
│   │   └── user.go              # Pure domain models (no framework deps)
│   ├── repository/
│   │   ├── user_repository.go   # Repository interface
│   │   ├── user_memory_repository.go  # In-memory implementation
│   │   └── module.go            # fx.Module
│   └── service/
│       ├── user_service.go      # Business logic (uses interfaces)
│       └── module.go            # fx.Module
├── go.mod
└── go.sum
```

**Key Points**:
- `internal/domain`: Pure Go, no framework dependencies
- `internal/service`: Uses `hyperion.Context` interface only
- `internal/repository`: Implements your data access interfaces
- `cmd/server`: Composes everything with fx modules

---

## FAQ

### Q: Why don't I see any logs?

**A**: You're using the NoOp logger from `hyperion.CoreModule`. Install the Zap adapter when available (v2.1):

```go
import "github.com/mapoio/hyperion/adapter/zap"

fx.New(
    hyperion.CoreModule,
    viper.Module,
    zap.Module,  // Real logger
    // ...
)
```

### Q: How do I test my services?

**A**: Inject mock implementations:

```go
func TestUserService(t *testing.T) {
    mockRepo := &mockUserRepository{}
    service := NewUserService(mockRepo)

    // Create mock context
    ctx := &mockContext{
        logger: &mockLogger{},
        tracer: &mockTracer{},
    }

    user, err := service.CreateUser(ctx, "test", "test@example.com")
    // assertions...
}
```

Since your service uses interfaces, testing is straightforward.

### Q: What's the difference from v1.0?

**v1.0**:
- Import: `github.com/mapoio/hyperion/pkg/hyperlog`
- Bundled implementations (forced to use Zap, GORM, etc.)
- Context exposed all methods directly

**v2.0**:
- Import: `github.com/mapoio/hyperion`
- Choose your adapters (or write your own)
- Context uses accessor pattern (`ctx.Logger()`, `ctx.Tracer()`)
- Core library has ZERO lock-in (only depends on fx)

### Q: Can I use sqlx instead of GORM?

**A**: Yes! Write your own adapter:

```go
// adapter/sqlx/module.go
package sqlx

import (
    "go.uber.org/fx"
    "github.com/mapoio/hyperion"
)

var Module = fx.Module("hyperion.adapter.sqlx",
    fx.Provide(
        fx.Annotate(
            NewSqlxDatabase,
            fx.As(new(hyperion.Database)),
        ),
    ),
)

func NewSqlxDatabase(cfg hyperion.Config) (hyperion.Database, error) {
    // Your sqlx implementation
}
```

This is the **power of v2.0**: complete flexibility.

---

## Learn More

### Core Concepts
- [Architecture Overview](architecture.md) - Complete v2.0 architecture with Interceptor and Meter
- [Architecture Decisions](architecture-decisions.md) - ADRs explaining design choices
- [Coding Standards](architecture/coding-standards.md) - Best practices and conventions

### Observability Guides
- [Interceptor Pattern](interceptor.md) - Complete guide to 3-line interceptor pattern
- [Observability Architecture](observability.md) - Unified Logs, Traces, and Metrics correlation

### Implementation Details
- [Source Tree Guide](architecture/source-tree.md) - Monorepo structure
- [Tech Stack](architecture/tech-stack.md) - Technology choices

### Advanced Topics
- **✅ Interceptor Pattern**: Automatic tracing, logging, and metrics (v2.0-2.2)
- **✅ Metrics Collection**: OpenTelemetry-compatible Meter interface (v2.0-2.2)
- **✅ Transaction Management**: UnitOfWork with GORM adapter (v2.0-2.2)
- **🔜 OpenTelemetry Integration**: Full OTel adapter with exemplars (Epic 3, Q1 2026)
- Web Server (hyperweb with Gin) - v2.3
- gRPC Server (hypergrpc) - v2.3

---

## What's Next?

Now that you understand the basics:

1. **Explore the Viper adapter source** to see how adapters work
2. **Write your own adapter** for a library you prefer
3. **Build a real application** with multiple layers
4. **Wait for v2.1** for Zap and GORM adapters

---

**Welcome to Hyperion v2.0 - Zero Lock-in, Maximum Flexibility! 🚀**

**Last Updated**: October 2025
**Version**: 2.0 (Core-Adapter Architecture)
