# Hyperion Framework Architecture Design

**Version**: 2.0
**Date**: 2025-01-XX
**Status**: Design Complete

---

## Table of Contents

1. [Overview](#1-overview)
2. [Core Architectural Principles](#2-core-architectural-principles)
3. [System Architecture](#3-system-architecture)
4. [Core Components](#4-core-components)
5. [Component Details](#5-component-details)
6. [Application Template](#6-application-template)
7. [Best Practices](#7-best-practices)
8. [Testing Strategy](#8-testing-strategy)
9. [Roadmap](#9-roadmap)

---

## 1. Overview

**Hyperion** is a production-ready, microkernel-based Go backend framework built on `go.uber.org/fx` dependency injection. It provides a modular, pluggable architecture with comprehensive observability, type-safe context management, and declarative transaction handling.

### 1.1 Key Features

- ✅ **Modular Architecture**: All features provided as independent `fx.Module`
- ✅ **Type-Safe Context**: `hyperctx.Context` with integrated tracing, logging, and database access
- ✅ **OpenTelemetry Integration**: Automatic distributed tracing across all layers
- ✅ **Declarative Transactions**: UnitOfWork pattern with automatic transaction propagation
- ✅ **Hot Reload Configuration**: Support for file and remote config sources (Consul/Etcd)
- ✅ **Production-Ready Defaults**: Structured logging, graceful shutdown, health checks
- ✅ **Comprehensive Error Handling**: Typed error codes with multi-layer wrapping

### 1.2 Technology Stack

| Layer | Technology |
|-------|-----------|
| **DI Framework** | go.uber.org/fx |
| **Web Framework** | Gin |
| **RPC Framework** | gRPC |
| **ORM** | GORM |
| **Config** | Viper |
| **Logging** | Zap |
| **Tracing** | OpenTelemetry |
| **Cache** | Ristretto / Redis |
| **Validation** | go-playground/validator |
| **HTTP Client** | Resty |

---

## 2. Core Architectural Principles

### 2.1 Design Philosophy

1. **Modularity over Monolith**
   All functionality is provided as independent `fx.Module`. Applications import only what they need.

2. **Convention over Configuration**
   Production-grade defaults for all components. Override only when necessary.

3. **Explicit over Implicit**
   Dependencies are declared explicitly through function signatures and fx constructors.

4. **Interface-Driven Design**
   Core components expose capabilities through interfaces, enabling decoupling and testability.

5. **Production-Ready by Default**
   Built-in structured logging, graceful shutdown, health checks, metrics, and tracing.

### 2.2 Architectural Style

Hyperion adopts **Layered Architecture** with Clean Architecture principles:

```
┌─────────────────────────────────────────────────────────────┐
│  Presentation Layer (hyperweb/hypergrpc)                    │
│  - HTTP Handlers, gRPC Services                              │
│  - Request validation, serialization                         │
│  - Calls application services                                │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│  Application Service Layer (internal/service)               │
│  - Business use cases implementation                         │
│  - Transaction orchestration (UnitOfWork)                   │
│  - Coordinates domain and data access                        │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│  Domain Layer (internal/domain) - Optional                  │
│  - Core business entities, value objects                     │
│  - Domain services, domain events                            │
│  - No external dependencies                                  │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│  Infrastructure Layer (internal/repository, pkg/*)          │
│  - Repository implementations                                │
│  - External service clients                                  │
│  - Cache, storage, message queue                             │
└─────────────────────────────────────────────────────────────┘
```

**Dependency Rule**: Outer layers depend on inner layers. Infrastructure details are injected at runtime.

---

## 3. System Architecture

### 3.1 Project Structure

```
hyperion/
├── pkg/                           # Framework core components
│   ├── hyperion/                  # Framework entry point
│   ├── hyperctx/                  # Context abstraction (CORE)
│   ├── hyperconfig/               # Configuration management
│   ├── hyperlog/                  # Structured logging
│   ├── hyperdb/                   # Database + UnitOfWork
│   ├── hypercache/                # Cache abstraction
│   ├── hyperstore/                # Object storage (S3)
│   ├── hypercrypto/               # Encryption utilities
│   ├── hypererror/                # Error handling
│   ├── hypervalidator/            # Request validation
│   ├── hyperhttp/                 # HTTP client
│   ├── hyperweb/                  # Web server (Gin)
│   └── hypergrpc/                 # gRPC server
│
├── examples/                      # Example applications
│   ├── simple-api/                # REST API example
│   └── fullstack/                 # Web + gRPC + DB example
│
├── .doc/                          # Design documentation
├── go.mod
└── README.md
```

### 3.2 Module Path

```
github.com/mapoio/hyperion
```

### 3.3 Component Dependency Graph

```
                    ┌─────────────┐
                    │  hyperion   │  (Entry point)
                    └──────┬──────┘
                           │
        ┌──────────────────┼──────────────────┐
        ▼                  ▼                  ▼
   hyperconfig        hyperlog           hyperdb
        │                  │                  │
        │                  │                  │
        │                  └────────┬─────────┘
        │                           │
        └───────────────────────────▼
                            hyperctx  (Core)
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
   hyperweb               hypergrpc              hypercache
        │                       │                       │
        │                       │                       │
        └───────────────────────┴───────────────────────┘
                                │
                          (Applications)
```

---

## 4. Core Components

### 4.1 Component Matrix

| Component | Purpose | Key Dependencies | Exportable |
|-----------|---------|------------------|------------|
| `hyperctx` | Type-safe context with trace/log/db | `hyperlog`, OpenTelemetry | ✅ |
| `hyperconfig` | Configuration management | Viper | ✅ |
| `hyperlog` | Structured logging | Zap | ✅ |
| `hyperdb` | Database + UnitOfWork | GORM, `hyperctx` | ✅ |
| `hypererror` | Typed error handling | - | ✅ |
| `hypercache` | Cache abstraction | Ristretto/Redis | ✅ |
| `hypervalidator` | Request validation | go-playground/validator | ✅ |
| `hyperhttp` | HTTP client with tracing | Resty, `hyperctx` | ✅ |
| `hyperweb` | Web server | Gin, `hyperctx` | ✅ |
| `hypergrpc` | gRPC server | gRPC, `hyperctx` | ✅ |

---

## 5. Component Details

### 5.1 hyperctx - Context Abstraction (CORE)

**The most critical innovation of Hyperion.**

#### Interface

```go
type Context interface {
    context.Context

    // Core dependencies
    Logger() hyperlog.Logger
    DB() *gorm.DB
    User() User

    // OpenTelemetry tracing
    Span() trace.Span
    TraceID() string
    SpanID() string
    StartSpan(layer, component, operation string) (Context, func())
    RecordError(err error)
    SetAttributes(attrs ...attribute.KeyValue)
    AddEvent(name string, attrs ...attribute.KeyValue)

    // Baggage (cross-service data)
    Baggage() baggage.Baggage
    WithBaggage(key, value string) Context
    GetBaggage(key string) string

    // Context management
    WithValue(key ContextKey[any], val any) Context
    Value(key ContextKey[any]) (any, bool)
    WithTimeout(timeout time.Duration) (Context, context.CancelFunc)
    WithCancel() (Context, context.CancelFunc)
    WithDeadline(deadline time.Time) (Context, context.CancelFunc)
}
```

#### Key Features

1. **Type-Safe Dependencies**: Direct access to `Logger()`, `DB()`, `User()`
2. **Integrated Tracing**: `StartSpan()` creates child spans automatically
3. **Automatic Propagation**: Trace context propagates across service boundaries
4. **Immutable Design**: All `WithXxx()` methods return new instances
5. **Transaction-Aware**: `DB()` returns transaction handle when inside `WithTransaction()`

#### Usage Example

```go
func (s *UserService) GetByID(ctx hyperctx.Context, id string) (*User, error) {
    // Start span (one line)
    ctx, end := ctx.StartSpan("service", "UserService", "GetByID")
    defer end()

    // Add attributes
    ctx.SetAttributes(attribute.String("user_id", id))

    // Access logger (with trace context)
    ctx.Logger().Info("fetching user")

    // Call repository (span auto-propagates)
    user, err := s.userRepo.FindByID(ctx, id)
    if err != nil {
        ctx.RecordError(err) // Auto-record in span
        return nil, err
    }

    return user, nil
}
```

---

### 5.2 hyperconfig - Configuration Management

#### Interface

```go
type Provider interface {
    Unmarshal(key string, rawVal any) error
    Get(key string) any
    GetString(key string) string
    GetInt(key string) int
    // ... other getters
}

type Watcher interface {
    Watch(callback func(event ChangeEvent)) (stop func(), err error)
}
```

#### Implementations

1. **ViperProvider**: File-based config (YAML/JSON/TOML)
2. **ConsulProvider**: Remote config with hot reload
3. **EtcdProvider**: (Future support)

#### Configuration Example

```yaml
# configs/config.yaml
log:
  level: info
  format: json
  output:
    - type: stdout
    - type: file
      path: /var/log/app/app.log
      max_size: 100
      max_backups: 3

database:
  driver: postgres
  dsn: "host=localhost user=test password=test dbname=testdb"
  max_open_conns: 100
  max_idle_conns: 10

cache:
  type: redis
  addr: "localhost:6379"
```

#### Hot Reload

```go
// Automatic hot reload for log level
cfgProvider.Watch(func(event hyperconfig.ChangeEvent) {
    var cfg hyperlog.Config
    cfgProvider.Unmarshal("log", &cfg)
    logger.SetLevel(parseLevel(cfg.Level))
})
```

---

### 5.3 hyperlog - Structured Logging

#### Interface

```go
type Logger interface {
    Debug(msg string, fields ...any)
    Info(msg string, fields ...any)
    Warn(msg string, fields ...any)
    Error(msg string, fields ...any)
    Fatal(msg string, fields ...any)

    With(fields ...any) Logger
    WithError(err error) Logger

    SetLevel(level Level)
    GetLevel() Level
    Sync() error
}
```

#### Implementation

- Based on `go.uber.org/zap`
- JSON or console output
- Dynamic log level adjustment
- File rotation support (lumberjack)
- Automatic trace context injection

#### Usage

```go
logger.Info("user created",
    "user_id", user.ID,
    "email", user.Email,
    "trace_id", ctx.TraceID(),
)
```

---

### 5.4 hyperdb - Database + UnitOfWork

#### Interfaces

```go
type DB interface {
    GetDB() *gorm.DB
    Health(ctx context.Context) error
    Close() error
}

type UnitOfWork interface {
    WithTransaction(ctx hyperctx.Context, fn func(txCtx hyperctx.Context) error) error
    WithTransactionOptions(ctx hyperctx.Context, opts *TransactionOptions, fn func(txCtx hyperctx.Context) error) error
}
```

#### Key Features

1. **GORM Integration**: Full GORM capabilities
2. **Connection Pool Management**: Auto-configured with defaults
3. **UnitOfWork Pattern**: Declarative transaction boundaries
4. **Automatic Tracing**: GORM plugin creates spans for all queries
5. **Transaction Propagation**: `ctx.DB()` automatically resolves transaction handle

#### Usage Example

```go
// Service layer
func (s *UserService) CreateUser(ctx hyperctx.Context, req *CreateUserRequest) (*User, error) {
    var createdUser *User

    // Declarative transaction
    err := s.uow.WithTransaction(ctx, func(txCtx hyperctx.Context) error {
        // Create user
        user := &User{Username: req.Username}
        if err := s.userRepo.Create(txCtx, user); err != nil {
            return err // Auto-rollback
        }

        // Create profile (same transaction)
        profile := &UserProfile{UserID: user.ID}
        if err := s.profileRepo.Create(txCtx, profile); err != nil {
            return err // Auto-rollback
        }

        createdUser = user
        return nil // Auto-commit
    })

    return createdUser, err
}

// Repository layer
func (r *UserRepository) Create(ctx hyperctx.Context, user *User) error {
    // ctx.DB() automatically returns tx handle if in transaction
    return ctx.DB().WithContext(ctx).Create(user).Error
}
```

---

### 5.5 hypererror - Error Handling

#### Design

**Type-safe error codes with multi-layer wrapping support.**

#### Core Types

```go
type Code struct {
    Code       string     // "USER_NOT_FOUND"
    HTTPStatus int        // 404
    GRPCCode   codes.Code // codes.NotFound
}

type Error struct {
    code    Code
    message string
    cause   error          // Underlying error (wrapping)
    fields  map[string]any // Context fields
}
```

#### Predefined Codes

```go
var (
    CodeBadRequest          = Code{"BAD_REQUEST", 400, codes.InvalidArgument}
    CodeUnauthorized        = Code{"UNAUTHORIZED", 401, codes.Unauthenticated}
    CodeForbidden           = Code{"FORBIDDEN", 403, codes.PermissionDenied}
    CodeNotFound            = Code{"NOT_FOUND", 404, codes.NotFound}
    CodeInternal            = Code{"INTERNAL_ERROR", 500, codes.Internal}
    CodeValidationFailed    = Code{"VALIDATION_FAILED", 400, codes.InvalidArgument}
    // ...
)
```

#### Usage

```go
// Simple error
return hypererror.NotFound("user not found")

// Error with fields
return hypererror.ResourceNotFound("user", userID)

// Wrap underlying error
return hypererror.InternalWrap("database query failed", err)

// Multi-layer wrapping
err := hypererror.Wrap(
    hypererror.CodeInternal,
    "failed to create user",
    err,
).WithField("email", email)

// Error checking
if hypererror.IsNotFound(err) {
    // handle not found
}

// Extract error chain
if hyperErr, ok := hypererror.As(err); ok {
    chain := hyperErr.Chain()
    cause := hyperErr.Cause()
}
```

---

### 5.6 hypercache - Cache Abstraction

#### Interface

```go
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    MGet(ctx context.Context, keys ...string) (map[string][]byte, error)
    MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error
}
```

#### Implementations

1. **RistrettoCache**: In-memory cache (high performance)
2. **RedisCache**: Distributed cache

#### Automatic Tracing

```go
cache = hypercache.WithTracing(cache)
// All operations automatically create spans
```

---

### 5.7 hypervalidator - Request Validation

#### Usage

```go
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=32"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"gte=18,lte=120"`
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Validate
    if err := h.validator.Struct(&req); err != nil {
        // Auto-converts to hypererror with field details
        c.JSON(400, err)
        return
    }

    // Process...
}
```

---

### 5.8 hyperhttp - HTTP Client

#### Features

- Built on `go-resty/resty`
- Automatic tracing (trace context injection)
- Configurable timeouts and retries

#### Usage

```go
func (c *Client) Get(ctx hyperctx.Context, url string) (*resty.Response, error) {
    ctx, end := ctx.StartSpan("http.client", "GET", url)
    defer end()

    req := c.client.R().SetContext(ctx)

    // Auto-inject trace context
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

    resp, err := req.Get(url)
    if err != nil {
        ctx.RecordError(err)
        return nil, err
    }

    ctx.SetAttributes(attribute.Int("http.status_code", resp.StatusCode()))
    return resp, nil
}
```

---

### 5.9 hyperweb - Web Server (Gin)

#### Core Middleware

1. **TraceMiddleware**: Creates `hyperctx.Context` with trace extraction
2. **RecoveryMiddleware**: Panic recovery with logging
3. **LoggerMiddleware**: Request/response logging
4. **CORSMiddleware**: CORS handling

#### Span Hierarchy

```
HTTP Request
├── handler.UserHandler.GetUser
│   └── service.UserService.GetByID
│       └── repository.UserRepository.FindByID
│           └── db.query.users
```

#### Usage

```go
func NewUserHandler(server *hyperweb.Server, userService *service.UserService) {
    engine := server.Engine()

    v1 := engine.Group("/api/v1")
    {
        v1.GET("/users/:id", h.GetUser)
        v1.POST("/users", h.CreateUser)
    }
}

func (h *UserHandler) GetUser(c *gin.Context) {
    ctx := c.MustGet("hyperctx").(hyperctx.Context)

    userID := c.Param("id")
    user, err := h.userService.GetByID(ctx, userID)

    if err != nil {
        ctx.RecordError(err)
        status := hypererror.GetHTTPStatus(err)

        if hyperErr, ok := hypererror.As(err); ok {
            c.JSON(status, hyperErr.ToResponse())
        } else {
            c.JSON(500, gin.H{"error": "internal error"})
        }
        return
    }

    c.JSON(200, user)
}
```

---

### 5.10 hypergrpc - gRPC Server

#### Core Interceptors

1. **UnaryInterceptor**: Creates `hyperctx.Context` with trace extraction
2. **StreamInterceptor**: Streaming RPC support

#### Health Check

- Automatic `grpc_health_v1.Health` service registration

#### Usage

```go
func RegisterUserService(server *hypergrpc.Server, userService *service.UserService) {
    pb.RegisterUserServiceServer(server.Server(), &UserServiceImpl{
        userService: userService,
    })
}
```

---

## 6. Application Template

### 6.1 Standard Project Layout

```
your-app/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── configs/
│   └── config.yaml              # Configuration file
├── internal/
│   ├── domain/                  # Domain models (optional)
│   ├── handler/                 # HTTP/gRPC handlers
│   │   ├── user_handler.go
│   │   └── module.go
│   ├── service/                 # Business logic
│   │   ├── user_service.go
│   │   └── module.go
│   └── repository/              # Data access
│       ├── user_repository.go
│       └── module.go
├── api/                         # API definitions
│   ├── proto/                   # Protobuf files
│   └── openapi/                 # OpenAPI/Swagger specs
├── go.mod
└── Dockerfile
```

### 6.2 Main Entry Point

```go
// cmd/server/main.go
package main

import (
    "github.com/mapoio/hyperion/pkg/hyperion"
    "github.com/your-app/internal/handler"
    "github.com/your-app/internal/repository"
    "github.com/your-app/internal/service"
    "go.uber.org/fx"
)

func main() {
    app := fx.New(
        // Import Hyperion core
        hyperion.Web(), // or hyperion.GRPC() or hyperion.FullStack()

        // Register application modules
        repository.Module,
        service.Module,
        handler.Module,
    )

    app.Run()
}
```

### 6.3 Module Definition

```go
// internal/service/module.go
package service

import "go.uber.org/fx"

var Module = fx.Module("service",
    fx.Provide(
        NewUserService,
        NewOrderService,
        // ...
    ),
)
```

---

## 7. Best Practices

### 7.1 Code Organization

1. **Module Encapsulation**: Group related functionality in `fx.Module`
2. **Interface Segregation**: Define small, focused interfaces
3. **Dependency Injection**: Never use global variables; inject dependencies via constructors

### 7.2 Error Handling

1. **Service Layer**: Return wrapped `hypererror` with business context
2. **Repository Layer**: Return `hypererror` for known errors, wrap unexpected errors
3. **Handler Layer**: Convert `hypererror` to HTTP/gRPC responses

```go
// Repository
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, hypererror.ResourceNotFound("user", id)
}
return nil, hypererror.InternalWrap("query failed", err)

// Service
if err != nil {
    return nil, hypererror.Wrap(
        hypererror.CodeInternal,
        "failed to create user",
        err,
    ).WithField("email", email)
}

// Handler
if err != nil {
    status := hypererror.GetHTTPStatus(err)
    c.JSON(status, hypererror.As(err).ToResponse())
}
```

### 7.3 Transaction Management

1. **Service Layer**: Define transaction boundaries with `UnitOfWork`
2. **Repository Layer**: Use `ctx.DB()` transparently
3. **Nested Transactions**: Not supported; use savepoints if needed

### 7.4 Logging

1. **Use `ctx.Logger()`**: Always use context logger (includes trace ID)
2. **Structured Fields**: Use key-value pairs
3. **Log Levels**: Debug for development, Info for production events, Error for failures

### 7.5 Testing

1. **Unit Tests**: Mock interfaces for dependencies
2. **Integration Tests**: Use test database and fx testing utilities
3. **Table-Driven Tests**: Follow Go conventions

---

## 8. Testing Strategy

### 8.1 Unit Testing

```go
// Service unit test with mocks
type mockUserRepository struct {
    mock.Mock
}

func (m *mockUserRepository) FindByID(ctx hyperctx.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*User), args.Error(1)
}

func TestUserService_GetByID(t *testing.T) {
    mockRepo := new(mockUserRepository)
    service := NewUserService(mockRepo)

    mockRepo.On("FindByID", mock.Anything, "123").Return(&User{ID: "123"}, nil)

    user, err := service.GetByID(ctx, "123")

    assert.NoError(t, err)
    assert.Equal(t, "123", user.ID)
    mockRepo.AssertExpectations(t)
}
```

### 8.2 Integration Testing

```go
func TestIntegration_UserFlow(t *testing.T) {
    app := fx.New(
        hyperion.Core(),
        repository.Module,
        service.Module,
        fx.NopLogger,
    )

    app.Run()
    defer app.Stop(context.Background())

    // Test with real components
}
```

---

## 9. Roadmap

### v1.0 (Current Design)

- ✅ Core framework (fx-based)
- ✅ hyperctx with integrated tracing
- ✅ Configuration management
- ✅ Structured logging
- ✅ Database + UnitOfWork
- ✅ Error handling
- ✅ Cache abstraction
- ✅ Validation
- ✅ HTTP client
- ✅ Web server (Gin)
- ✅ gRPC server

### v1.1 (Q1 2025)

- OpenTelemetry exporter configuration
- Prometheus metrics integration
- Health check endpoints
- Production examples

### v1.2 (Q2 2025)

- Message queue component (`hypermq`)
  - Kafka support
  - RabbitMQ support
- Object storage enhancements
- Authentication/Authorization helpers

### v1.3 (Q3 2025)

- Distributed task scheduling (`hypercron`)
- Rate limiting component
- Circuit breaker pattern

### v2.0 (Q4 2025)

- Generic repository and service patterns
- Code generation tools
- Admin dashboard

---

## Appendix A: Configuration Reference

### Complete Configuration Example

```yaml
app:
  name: "my-app"
  env: "production"
  version: "1.0.0"

log:
  level: info
  format: json
  output:
    - type: stdout
    - type: file
      path: /var/log/app/app.log
      max_size: 100
      max_backups: 3
      max_age: 7
      compress: true

database:
  driver: postgres
  dsn: "host=localhost user=test password=test dbname=testdb port=5432 sslmode=disable"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 1h
  conn_max_idle_time: 10m
  log_level: warn

cache:
  type: redis
  addr: "localhost:6379"
  password: ""
  db: 0
  pool_size: 10

web:
  host: "0.0.0.0"
  port: 8080
  mode: release
  read_timeout: 30s
  write_timeout: 30s

grpc:
  port: 9090

tracing:
  enabled: true
  exporter: jaeger
  endpoint: "http://localhost:14268/api/traces"
  sample_rate: 1.0
```

---

## Appendix B: Error Code Reference

| Code | HTTP | gRPC | Description |
|------|------|------|-------------|
| `BAD_REQUEST` | 400 | InvalidArgument | Invalid request parameters |
| `UNAUTHORIZED` | 401 | Unauthenticated | Authentication required |
| `FORBIDDEN` | 403 | PermissionDenied | Insufficient permissions |
| `NOT_FOUND` | 404 | NotFound | Resource not found |
| `CONFLICT` | 409 | AlreadyExists | Resource conflict |
| `VALIDATION_FAILED` | 400 | InvalidArgument | Request validation failed |
| `INTERNAL_ERROR` | 500 | Internal | Internal server error |
| `SERVICE_UNAVAILABLE` | 503 | Unavailable | Service temporarily unavailable |

---

## Appendix C: Span Naming Convention

```
{layer}.{component}.{operation}

Examples:
- handler.UserHandler.GetUser
- service.UserService.GetByID
- repository.UserRepository.FindByID
- db.query.users
- db.create.users
- cache.Get.user:123
- http.client.GET.https://api.example.com/users
```

---

**End of Architecture Document**
