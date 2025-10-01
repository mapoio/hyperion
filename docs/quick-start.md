# Hyperion Quick Start Guide

Build your first Hyperion application in 10 minutes.

---

## Prerequisites

- Go 1.21+
- Basic Go programming knowledge
- (Optional) Docker - for running PostgreSQL/Redis

---

## Step 1: Create Project

```bash
mkdir my-hyperion-app
cd my-hyperion-app
go mod init github.com/yourusername/my-hyperion-app
```

---

## Step 2: Install Hyperion

```bash
go get github.com/mapoio/hyperion
```

---

## Step 3: Create Configuration File

```bash
mkdir configs
```

Create `configs/config.yaml`:

```yaml
log:
  level: info
  format: json

database:
  driver: postgres
  dsn: "host=localhost user=test password=test dbname=testdb port=5432 sslmode=disable"

web:
  host: "0.0.0.0"
  port: 8080
  mode: debug
```

---

## Step 4: Create Domain Model

Create `internal/domain/user.go`:

```go
package domain

import (
    "time"
)

type User struct {
    ID        string    `gorm:"primaryKey"`
    Username  string    `gorm:"uniqueIndex;not null"`
    Email     string    `gorm:"uniqueIndex;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

---

## Step 5: Create Repository

Create `internal/repository/user_repository.go`:

```go
package repository

import (
    "github.com/mapoio/hyperion/pkg/hyperctx"
    "github.com/mapoio/hyperion/pkg/hypererror"
    "github.com/yourusername/my-hyperion-app/internal/domain"
    "go.opentelemetry.io/otel/attribute"
    "gorm.io/gorm"
)

type UserRepository interface {
    Create(ctx hyperctx.Context, user *domain.User) error
    FindByID(ctx hyperctx.Context, id string) (*domain.User, error)
    FindAll(ctx hyperctx.Context) ([]*domain.User, error)
}

type userRepository struct{}

func NewUserRepository() UserRepository {
    return &userRepository{}
}

func (r *userRepository) Create(ctx hyperctx.Context, user *domain.User) error {
    ctx, end := ctx.StartSpan("repository", "UserRepository", "Create")
    defer end()

    if err := ctx.DB().WithContext(ctx).Create(user).Error; err != nil {
        ctx.RecordError(err)
        return hypererror.InternalWrap("failed to create user", err)
    }

    return nil
}

func (r *userRepository) FindByID(ctx hyperctx.Context, id string) (*domain.User, error) {
    ctx, end := ctx.StartSpan("repository", "UserRepository", "FindByID")
    defer end()

    ctx.SetAttributes(attribute.String("user_id", id))

    var user domain.User
    err := ctx.DB().WithContext(ctx).First(&user, "id = ?", id).Error

    if err == gorm.ErrRecordNotFound {
        return nil, hypererror.ResourceNotFound("user", id)
    }

    if err != nil {
        ctx.RecordError(err)
        return nil, hypererror.InternalWrap("failed to find user", err)
    }

    return &user, nil
}

func (r *userRepository) FindAll(ctx hyperctx.Context) ([]*domain.User, error) {
    ctx, end := ctx.StartSpan("repository", "UserRepository", "FindAll")
    defer end()

    var users []*domain.User
    if err := ctx.DB().WithContext(ctx).Find(&users).Error; err != nil {
        ctx.RecordError(err)
        return nil, hypererror.InternalWrap("failed to find users", err)
    }

    return users, nil
}
```

Create `internal/repository/module.go`:

```go
package repository

import "go.uber.org/fx"

var Module = fx.Module("repository",
    fx.Provide(
        NewUserRepository,
    ),
)
```

---

## Step 6: Create Service

Create `internal/service/user_service.go`:

```go
package service

import (
    "github.com/google/uuid"
    "github.com/mapoio/hyperion/pkg/hyperctx"
    "github.com/mapoio/hyperion/pkg/hyperdb"
    "github.com/yourusername/my-hyperion-app/internal/domain"
    "github.com/yourusername/my-hyperion-app/internal/repository"
    "go.opentelemetry.io/otel/attribute"
)

type UserService interface {
    CreateUser(ctx hyperctx.Context, username, email string) (*domain.User, error)
    GetUser(ctx hyperctx.Context, id string) (*domain.User, error)
    ListUsers(ctx hyperctx.Context) ([]*domain.User, error)
}

type userService struct {
    uow      hyperdb.UnitOfWork
    userRepo repository.UserRepository
}

func NewUserService(
    uow hyperdb.UnitOfWork,
    userRepo repository.UserRepository,
) UserService {
    return &userService{
        uow:      uow,
        userRepo: userRepo,
    }
}

func (s *userService) CreateUser(ctx hyperctx.Context, username, email string) (*domain.User, error) {
    ctx, end := ctx.StartSpan("service", "UserService", "CreateUser")
    defer end()

    ctx.SetAttributes(
        attribute.String("username", username),
        attribute.String("email", email),
    )

    var createdUser *domain.User

    err := s.uow.WithTransaction(ctx, func(txCtx hyperctx.Context) error {
        user := &domain.User{
            ID:       uuid.New().String(),
            Username: username,
            Email:    email,
        }

        if err := s.userRepo.Create(txCtx, user); err != nil {
            return err
        }

        createdUser = user
        return nil
    })

    if err != nil {
        ctx.RecordError(err)
        return nil, err
    }

    return createdUser, nil
}

func (s *userService) GetUser(ctx hyperctx.Context, id string) (*domain.User, error) {
    ctx, end := ctx.StartSpan("service", "UserService", "GetUser")
    defer end()

    user, err := s.userRepo.FindByID(ctx, id)
    if err != nil {
        ctx.RecordError(err)
        return nil, err
    }

    return user, nil
}

func (s *userService) ListUsers(ctx hyperctx.Context) ([]*domain.User, error) {
    ctx, end := ctx.StartSpan("service", "UserService", "ListUsers")
    defer end()

    users, err := s.userRepo.FindAll(ctx)
    if err != nil {
        ctx.RecordError(err)
        return nil, err
    }

    return users, nil
}
```

Create `internal/service/module.go`:

```go
package service

import "go.uber.org/fx"

var Module = fx.Module("service",
    fx.Provide(
        NewUserService,
    ),
)
```

---

## Step 7: Create Handler

Create `internal/handler/user_handler.go`:

```go
package handler

import (
    "github.com/gin-gonic/gin"
    "github.com/mapoio/hyperion/pkg/hyperctx"
    "github.com/mapoio/hyperion/pkg/hypererror"
    "github.com/yourusername/my-hyperion-app/internal/service"
)

type UserHandler struct {
    userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
    return &UserHandler{
        userService: userService,
    }
}

type CreateUserRequest struct {
    Username string `json:"username" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    ctx := c.MustGet("hyperctx").(hyperctx.Context)

    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    user, err := h.userService.CreateUser(ctx, req.Username, req.Email)
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

    c.JSON(201, user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
    ctx := c.MustGet("hyperctx").(hyperctx.Context)

    userID := c.Param("id")

    user, err := h.userService.GetUser(ctx, userID)
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

func (h *UserHandler) ListUsers(c *gin.Context) {
    ctx := c.MustGet("hyperctx").(hyperctx.Context)

    users, err := h.userService.ListUsers(ctx)
    if err != nil {
        ctx.RecordError(err)
        c.JSON(500, gin.H{"error": "internal error"})
        return
    }

    c.JSON(200, gin.H{"users": users})
}
```

Create `internal/handler/module.go`:

```go
package handler

import (
    "github.com/mapoio/hyperion/pkg/hyperweb"
    "go.uber.org/fx"
)

var Module = fx.Module("handler",
    fx.Provide(NewUserHandler),
    fx.Invoke(RegisterRoutes),
)

func RegisterRoutes(server *hyperweb.Server, userHandler *UserHandler) {
    engine := server.Engine()

    api := engine.Group("/api/v1")
    {
        users := api.Group("/users")
        {
            users.POST("", userHandler.CreateUser)
            users.GET("/:id", userHandler.GetUser)
            users.GET("", userHandler.ListUsers)
        }
    }
}
```

---

## Step 8: Create Main Program

Create `cmd/server/main.go`:

```go
package main

import (
    "github.com/mapoio/hyperion/pkg/hyperion"
    "github.com/yourusername/my-hyperion-app/internal/handler"
    "github.com/yourusername/my-hyperion-app/internal/repository"
    "github.com/yourusername/my-hyperion-app/internal/service"
    "go.uber.org/fx"
)

func main() {
    app := fx.New(
        // Import Hyperion core with web server
        hyperion.Web(),

        // Register application modules
        repository.Module,
        service.Module,
        handler.Module,
    )

    app.Run()
}
```

---

## Step 9: Database Migration

Create `scripts/migrate.go`:

```go
package main

import (
    "log"

    "github.com/yourusername/my-hyperion-app/internal/domain"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    dsn := "host=localhost user=test password=test dbname=testdb port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("failed to connect database: %v", err)
    }

    // Auto migrate
    if err := db.AutoMigrate(&domain.User{}); err != nil {
        log.Fatalf("failed to migrate: %v", err)
    }

    log.Println("Migration completed successfully")
}
```

Run migration:

```bash
go run scripts/migrate.go
```

---

## Step 10: Run Application

```bash
# Start database (if using Docker)
docker run --name postgres \
  -e POSTGRES_USER=test \
  -e POSTGRES_PASSWORD=test \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  -d postgres:15

# Run application
go run cmd/server/main.go
```

You should see output similar to:

```
{"level":"info","timestamp":"2025-01-XX...","message":"checking database connectivity..."}
{"level":"info","timestamp":"2025-01-XX...","message":"database connected successfully"}
{"level":"info","timestamp":"2025-01-XX...","message":"starting web server","addr":"0.0.0.0:8080"}
```

---

## Step 11: Test API

### Create User

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username":"john","email":"john@example.com"}'
```

Response:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "john",
  "email": "john@example.com",
  "created_at": "2025-01-XX...",
  "updated_at": "2025-01-XX..."
}
```

### Get User

```bash
curl http://localhost:8080/api/v1/users/550e8400-e29b-41d4-a716-446655440000
```

### List All Users

```bash
curl http://localhost:8080/api/v1/users
```

---

## Project Structure Overview

```
my-hyperion-app/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── configs/
│   └── config.yaml              # Configuration file
├── internal/
│   ├── domain/
│   │   └── user.go              # Domain models
│   ├── handler/
│   │   ├── user_handler.go      # HTTP handlers
│   │   └── module.go
│   ├── service/
│   │   ├── user_service.go      # Business logic
│   │   └── module.go
│   └── repository/
│       ├── user_repository.go   # Data access
│       └── module.go
├── scripts/
│   └── migrate.go               # Database migration
├── go.mod
└── go.sum
```

---

## Next Steps

Congratulations! You've successfully created your first Hyperion application.

### Learn More:

1. **Add Caching**: Integrate `hypercache` module
2. **Add Validation**: Use `hypervalidator`
3. **Error Handling**: Deep dive into `hypererror`
4. **Testing**: Write unit tests and integration tests
5. **Deployment**: Containerize with Docker

### View Complete Documentation:

- [Architecture Design](architecture.md)
- [Architecture Decisions](architecture-decisions.md)
- [Implementation Plan](implementation-plan.md)

---

## FAQ

### Q: How to enable debug logging?

Modify `configs/config.yaml`:

```yaml
log:
  level: debug
  format: console  # More readable format
```

### Q: How to add health check endpoint?

Add to `handler/module.go`:

```go
func RegisterRoutes(server *hyperweb.Server, userHandler *UserHandler) {
    engine := server.Engine()

    // Health check
    engine.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // ... other routes
}
```

### Q: How to view tracing data?

1. Install Jaeger:
   ```bash
   docker run -d -p 16686:16686 -p 14268:14268 jaegertracing/all-in-one:latest
   ```

2. Configure in `config.yaml`:
   ```yaml
   tracing:
     enabled: true
     exporter: jaeger
     endpoint: "http://localhost:14268/api/traces"
   ```

3. Visit http://localhost:16686 to view traces

---

**Happy Coding with Hyperion! 🚀**
