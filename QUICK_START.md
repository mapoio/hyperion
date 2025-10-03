# Hyperion Quick Start Guide

This guide shows you how to build a production-ready Go web application with Hyperion in 5 minutes.

## What You'll Build

A RESTful API with:
- ✅ Structured logging (Zap)
- ✅ Distributed tracing (OpenTelemetry)
- ✅ Metrics collection (OpenTelemetry)
- ✅ Database access (GORM)
- ✅ Automatic transaction management
- ✅ Dependency injection (fx)
- ✅ Clean architecture (Handler → Service → Repository)

## Table of Contents

1. [Installation](#installation)
2. [Project Structure](#project-structure)
3. [Step 1: Setup Main Application](#step-1-setup-main-application)
4. [Step 2: Create Middleware](#step-2-create-middleware)
5. [Step 3: Build Handler Layer](#step-3-build-handler-layer)
6. [Step 4: Build Service Layer](#step-4-build-service-layer)
7. [Step 5: Build Repository Layer](#step-5-build-repository-layer)
8. [Run Your Application](#run-your-application)
9. [Advanced Features](#advanced-features)

---

## Installation

```bash
go get github.com/mapoio/hyperion
go get github.com/gin-gonic/gin
```

---

## Project Structure

```
myapp/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── handler/
│   │   └── user_handler.go      # HTTP handlers
│   ├── service/
│   │   └── user_service.go      # Business logic
│   ├── repository/
│   │   └── user_repository.go   # Data access
│   ├── middleware/
│   │   └── hyperion.go          # Gin middleware
│   └── module.go                # fx modules
└── go.mod
```

---

## Step 1: Setup Main Application

**`cmd/api/main.go`**

```go
package main

import (
	"context"

	"github.com/gin-gonic/gin"
	hyperion "github.com/mapoio/hyperion"
	"go.uber.org/fx"
	"myapp/internal/handler"
	"myapp/internal/middleware"
	"myapp/internal/service"
	"myapp/internal/repository"
)

func main() {
	fx.New(
		// Hyperion core infrastructure
		hyperion.CoreModule,                  // Core with no-op defaults
		hyperion.TracingInterceptorModule,    // Enable OpenTelemetry tracing
		hyperion.LoggingInterceptorModule,    // Enable structured logging

		// Optional: Use real adapters
		// zap.Module,   // Real structured logging
		// otel.Module,  // Real OpenTelemetry tracing
		// gorm.Module,  // Real database with GORM

		// Application modules
		fx.Provide(service.NewUserService),
		fx.Provide(repository.NewUserRepository),
		fx.Provide(handler.NewUserHandler),

		// Start HTTP server
		fx.Invoke(StartServer),
	).Run()
}

// StartServer starts the Gin HTTP server
func StartServer(
	lc fx.Lifecycle,
	factory hyperion.ContextFactory,
	userHandler *handler.UserHandler,
) {
	r := gin.New()

	// Register middleware
	r.Use(middleware.HyperionContext(factory))
	r.Use(middleware.RequestID(factory))
	r.Use(gin.Recovery())

	// Register routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/users/:id", userHandler.GetUser)
		v1.POST("/users", userHandler.CreateUser)
		v1.PUT("/users/:id", userHandler.UpdateUser)
		v1.DELETE("/users/:id", userHandler.DeleteUser)
	}

	// Lifecycle management
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go r.Run(":8080")
			return nil
		},
	})
}
```

---

## Step 2: Create Middleware

**`internal/middleware/hyperion.go`**

```go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	hyperion "github.com/mapoio/hyperion"
)

// HyperionContext creates hyperion.Context from Gin request context
// and stores it in Gin context for downstream handlers.
func HyperionContext(factory hyperion.ContextFactory) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create hyperion.Context from request context
		hctx := factory.New(c.Request.Context())

		// Store in Gin context
		c.Set("hyperion.context", hctx)

		c.Next()
	}
}

// RequestID adds a unique request ID to the logger.
// This middleware MUST run after HyperionContext middleware.
func RequestID(factory hyperion.ContextFactory) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get or generate request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Get hyperion.Context
		hctx, exists := c.Get("hyperion.context")
		if !exists {
			c.Next()
			return
		}

		// Add request ID to logger
		ctx := hctx.(hyperion.Context)
		logger := ctx.Logger().With(
			"requestID", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)

		// Create new context with enriched logger
		ctx = hyperion.WithLogger(ctx, logger)

		// Update in Gin context
		c.Set("hyperion.context", ctx)

		// Set response header
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// GetHyperionContext extracts hyperion.Context from Gin context.
// Helper function for handlers.
func GetHyperionContext(c *gin.Context) hyperion.Context {
	return c.MustGet("hyperion.context").(hyperion.Context)
}
```

---

## Step 3: Build Handler Layer

**`internal/handler/user_handler.go`**

```go
package handler

import (
	"github.com/gin-gonic/gin"
	"myapp/internal/middleware"
	"myapp/internal/service"
)

// UserHandler handles HTTP requests for user resources.
type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// GetUser handles GET /api/v1/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	// Extract hyperion.Context
	ctx := middleware.GetHyperionContext(c)

	// Get user ID from URL
	userID := c.Param("id")

	// Call service layer
	user, err := h.service.GetUser(ctx, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, user)
}

// CreateUser handles POST /api/v1/users
func (h *UserHandler) CreateUser(c *gin.Context) {
	ctx := middleware.GetHyperionContext(c)

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.CreateUser(ctx, req.Name, req.Email)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, user)
}

// UpdateUser handles PUT /api/v1/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	ctx := middleware.GetHyperionContext(c)
	userID := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.UpdateUser(ctx, userID, req.Name, req.Email)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, user)
}

// DeleteUser handles DELETE /api/v1/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	ctx := middleware.GetHyperionContext(c)
	userID := c.Param("id")

	err := h.service.DeleteUser(ctx, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(204, nil)
}

// Request/Response DTOs
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email" binding:"omitempty,email"`
}
```

---

## Step 4: Build Service Layer

**`internal/service/user_service.go`**

```go
package service

import (
	"errors"
	"fmt"

	hyperion "github.com/mapoio/hyperion"
	"myapp/internal/repository"
)

// UserService handles user business logic.
type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetUser retrieves a user by ID.
//
// Key pattern: All service methods accept hyperion.Context as first parameter.
func (s *UserService) GetUser(ctx hyperion.Context, userID string) (_ *User, err error) {
	// Apply interceptors (automatic tracing + logging)
	ctx, end := ctx.UseIntercept("UserService", "GetUser")
	defer end(&err)

	// Validate input
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	// Log with context (automatically includes requestID)
	ctx.Logger().Info("fetching user", "userID", userID)

	// Call repository
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return user, nil
}

// CreateUser creates a new user.
func (s *UserService) CreateUser(ctx hyperion.Context, name, email string) (_ *User, err error) {
	ctx, end := ctx.UseIntercept("UserService", "CreateUser")
	defer end(&err)

	// Validate
	if name == "" || email == "" {
		return nil, errors.New("name and email are required")
	}

	ctx.Logger().Info("creating user", "name", name, "email", email)

	// Create user
	user := &User{
		Name:  name,
		Email: email,
	}

	// Save to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user.
func (s *UserService) UpdateUser(ctx hyperion.Context, userID, name, email string) (_ *User, err error) {
	ctx, end := ctx.UseIntercept("UserService", "UpdateUser")
	defer end(&err)

	ctx.Logger().Info("updating user", "userID", userID)

	// Fetch existing user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Update fields
	if name != "" {
		user.Name = name
	}
	if email != "" {
		user.Email = email
	}

	// Save changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user by ID.
func (s *UserService) DeleteUser(ctx hyperion.Context, userID string) (err error) {
	ctx, end := ctx.UseIntercept("UserService", "DeleteUser")
	defer end(&err)

	ctx.Logger().Info("deleting user", "userID", userID)

	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// User domain model
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
```

---

## Step 5: Build Repository Layer

**`internal/repository/user_repository.go`**

```go
package repository

import (
	"errors"

	hyperion "github.com/mapoio/hyperion"
	"myapp/internal/service"
)

// UserRepository handles user data access.
type UserRepository struct {
	// In real application, inject GORM DB or other database client
}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// FindByID retrieves a user by ID from database.
func (r *UserRepository) FindByID(ctx hyperion.Context, id string) (_ *service.User, err error) {
	ctx, end := ctx.UseIntercept("UserRepository", "FindByID")
	defer end(&err)

	// Use ctx.DB() to access database
	// In real application with GORM adapter:
	//
	// var user service.User
	// err := ctx.DB().QueryRow(ctx,
	//     "SELECT id, name, email FROM users WHERE id = ?", id,
	// ).Scan(&user.ID, &user.Name, &user.Email)
	//
	// Or with GORM:
	// err := ctx.DB().First(&user, "id = ?", id).Error

	// Mock implementation
	if id == "1" {
		return &service.User{
			ID:    "1",
			Name:  "John Doe",
			Email: "john@example.com",
		}, nil
	}

	return nil, errors.New("user not found")
}

// Create inserts a new user into database.
func (r *UserRepository) Create(ctx hyperion.Context, user *service.User) (err error) {
	ctx, end := ctx.UseIntercept("UserRepository", "Create")
	defer end(&err)

	// Use ctx.DB() for database operations
	// Real implementation:
	// err := ctx.DB().Exec(ctx,
	//     "INSERT INTO users (name, email) VALUES (?, ?)",
	//     user.Name, user.Email,
	// ).Error

	ctx.Logger().Info("user created", "name", user.Name)
	user.ID = "generated-id"
	return nil
}

// Update updates an existing user in database.
func (r *UserRepository) Update(ctx hyperion.Context, user *service.User) (err error) {
	ctx, end := ctx.UseIntercept("UserRepository", "Update")
	defer end(&err)

	// Real implementation:
	// err := ctx.DB().Exec(ctx,
	//     "UPDATE users SET name = ?, email = ? WHERE id = ?",
	//     user.Name, user.Email, user.ID,
	// ).Error

	ctx.Logger().Info("user updated", "userID", user.ID)
	return nil
}

// Delete removes a user from database.
func (r *UserRepository) Delete(ctx hyperion.Context, id string) (err error) {
	ctx, end := ctx.UseIntercept("UserRepository", "Delete")
	defer end(&err)

	// Real implementation:
	// err := ctx.DB().Exec(ctx,
	//     "DELETE FROM users WHERE id = ?", id,
	// ).Error

	ctx.Logger().Info("user deleted", "userID", id)
	return nil
}
```

---

## Run Your Application

```bash
# Start the server
go run cmd/api/main.go

# Test the API
curl http://localhost:8080/api/v1/users/1

# Create a user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Doe","email":"jane@example.com"}'
```

**Output logs** (with tracing and logging interceptors enabled):

```
[DEBUG] Method started path=UserService.GetUser
[INFO] fetching user userID=1 requestID=550e8400-e29b-41d4-a716-446655440000
[DEBUG] Method started path=UserRepository.FindByID
[DEBUG] Method completed path=UserRepository.FindByID duration=1.2ms
[DEBUG] Method completed path=UserService.GetUser duration=2.5ms
```

---

## Advanced Features

### 1. Using Transactions

```go
// In service layer
func (s *UserService) TransferMoney(ctx hyperion.Context, fromID, toID string, amount float64) (err error) {
	ctx, end := ctx.UseIntercept("UserService", "TransferMoney")
	defer end(&err)

	// Use UnitOfWork for transaction management
	return s.uow.WithTransaction(ctx, func(txCtx hyperion.Context) error {
		// txCtx.DB() now returns transaction executor
		// All operations use the same transaction

		// Deduct from sender
		if err := s.userRepo.DeductBalance(txCtx, fromID, amount); err != nil {
			return err // Transaction auto-rollback
		}

		// Add to receiver
		if err := s.userRepo.AddBalance(txCtx, toID, amount); err != nil {
			return err // Transaction auto-rollback
		}

		return nil // Transaction auto-commit
	})
}
```

### 2. Selective Interceptor Application

```go
// Only apply tracing (exclude logging for high-frequency calls)
func (s *UserService) HighFrequencyOperation(ctx hyperion.Context) (err error) {
	ctx, end := ctx.UseIntercept("UserService", "HighFrequencyOperation",
		hyperion.WithOnly("tracing"))  // Only tracing, no logging
	defer end(&err)

	// Business logic...
	return nil
}

// Only apply logging (exclude tracing for sensitive operations)
func (s *UserService) SensitiveOperation(ctx hyperion.Context) (err error) {
	ctx, end := ctx.UseIntercept("UserService", "SensitiveOperation",
		hyperion.WithExclude("tracing"))  // Only logging, no tracing
	defer end(&err)

	// Business logic...
	return nil
}
```

### 3. Custom Interceptors

```go
// Create a metrics interceptor
type MetricsInterceptor struct {
	registry *prometheus.Registry
}

func (m *MetricsInterceptor) Name() string {
	return "metrics"
}

func (m *MetricsInterceptor) Intercept(
	ctx hyperion.Context,
	fullPath string,
) (hyperion.Context, func(err *error), error) {
	start := time.Now()

	end := func(errPtr *error) {
		duration := time.Since(start)
		// Record metrics
		m.recordDuration(fullPath, duration)
		if errPtr != nil && *errPtr != nil {
			m.recordError(fullPath)
		}
	}

	return ctx, end, nil
}

func (m *MetricsInterceptor) Order() int {
	return 300 // After tracing and logging
}

// Register in main.go
fx.Provide(
	fx.Annotate(
		NewMetricsInterceptor,
		fx.ResultTags(`group:"hyperion.interceptors"`),
	),
)
```

### 4. Recording Metrics

```go
// In service layer - track request counts and latency
func (s *UserService) GetUser(ctx hyperion.Context, userID string) (_ *User, err error) {
	ctx, end := ctx.UseIntercept("UserService", "GetUser")
	defer end(&err)

	// Record request counter with attributes
	requestCounter := ctx.Meter().Counter("user.requests",
		hyperion.WithMetricDescription("Total user requests"),
		hyperion.WithMetricUnit("1"),
	)
	requestCounter.Add(ctx, 1,
		hyperion.String("method", "GetUser"),
		hyperion.String("status", "started"),
	)

	// Measure latency
	start := time.Now()
	defer func() {
		latency := ctx.Meter().Histogram("user.latency",
			hyperion.WithMetricDescription("User request latency"),
			hyperion.WithMetricUnit("ms"),
		)
		latency.Record(ctx, float64(time.Since(start).Milliseconds()))
	}()

	// Business logic
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		// Record error counter
		errorCounter := ctx.Meter().Counter("user.errors")
		errorCounter.Add(ctx, 1,
			hyperion.String("method", "GetUser"),
			hyperion.String("error_type", "not_found"),
		)
		return nil, err
	}

	// Record success counter
	requestCounter.Add(ctx, 1,
		hyperion.String("method", "GetUser"),
		hyperion.String("status", "success"),
	)

	return user, nil
}

// Track active connections with Gauge
func (s *ConnectionPool) TrackActiveConnections(ctx hyperion.Context) {
	activeConnections := ctx.Meter().Gauge("db.active_connections",
		hyperion.WithMetricDescription("Number of active database connections"),
	)
	activeConnections.Record(ctx, float64(s.pool.ActiveCount()))
}

// Track queue size with UpDownCounter
func (s *MessageQueue) TrackQueueSize(ctx hyperion.Context, delta int64) {
	queueSize := ctx.Meter().UpDownCounter("queue.size",
		hyperion.WithMetricDescription("Message queue size"),
	)
	queueSize.Add(ctx, delta,
		hyperion.String("queue_name", "user_events"),
	)
}
```

**Key Benefits**:
- **Automatic trace correlation**: When using OpenTelemetry adapter, metrics automatically include exemplars linking to traces
- **Context-aware**: Passing `ctx` enables automatic correlation without manual TraceID extraction
- **Type-safe attributes**: Use `hyperion.String()`, `hyperion.Int()`, etc. for type-safe metric attributes

### 5. Using Real Adapters

```go
// main.go
import (
	"github.com/mapoio/hyperion/adapter/zap"
	"github.com/mapoio/hyperion/adapter/otel"
	"github.com/mapoio/hyperion/adapter/gorm"
)

fx.New(
	hyperion.CoreModule,
	hyperion.AllInterceptorsModule,

	// Real adapters
	zap.Module,   // Structured logging with Zap
	otel.Module,  // OpenTelemetry tracing + metrics
	gorm.Module,  // GORM database

	// Your modules...
)
```

---

## Key Takeaways

1. **ContextFactory** is injected by fx and creates `hyperion.Context` from standard `context.Context`

2. **Middleware pattern** - Create `hyperion.Context` once in middleware, reuse in all handlers

3. **Service layer pattern** - All methods accept `hyperion.Context` as first parameter

4. **UseIntercept pattern** - Apply interceptors with 3 lines:
   ```go
   ctx, end := ctx.UseIntercept("Service", "Method")
   defer end(&err)
   ```

5. **Automatic features**:
   - Structured logging with request context
   - Distributed tracing spans
   - Transaction propagation
   - Error recording

6. **Clean architecture** - Handler → Service → Repository, each layer uses `hyperion.Context`

---

## Next Steps

- Read [Hyperion Core README](hyperion/README.md) for core library overview
- Explore [Interceptor Guide](docs/interceptor.md) for advanced interceptor patterns
- Check [Observability Guide](docs/observability.md) for unified observability architecture
- Read [Architecture Guide](docs/architecture.md) for comprehensive design patterns
- See [Adapter Documentation](docs/adapters/) for real implementations

---

**Questions?** Open an issue at https://github.com/mapoio/hyperion/issues
