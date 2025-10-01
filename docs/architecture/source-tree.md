# Hyperion Framework Source Tree Guide

This document explains the organization of the Hyperion framework source code and provides guidelines for navigating the codebase.

---

## Framework Source Structure

```
hyperion/
├── pkg/                           # Framework core packages
│   ├── hyperion/                  # Framework entry point and module aggregation
│   │   ├── hyperion.go           # Core(), Web(), GRPC(), FullStack() constructors
│   │   └── module.go             # fx.Module aggregation
│   │
│   ├── hyperctx/                  # Context abstraction (CORE)
│   │   ├── context.go            # Context interface definition
│   │   ├── implementation.go     # Default implementation
│   │   ├── user.go               # User interface and default implementation
│   │   └── module.go             # fx.Module provider
│   │
│   ├── hyperconfig/               # Configuration management
│   │   ├── config.go             # Config interface
│   │   ├── viper.go              # Viper-based implementation
│   │   ├── watcher.go            # Hot reload support
│   │   └── module.go
│   │
│   ├── hyperlog/                  # Structured logging
│   │   ├── logger.go             # Logger interface
│   │   ├── zap.go                # Zap implementation
│   │   ├── level.go              # Dynamic level adjustment
│   │   └── module.go
│   │
│   ├── hyperdb/                   # Database + UnitOfWork
│   │   ├── db.go                 # Database interface
│   │   ├── gorm.go               # GORM implementation
│   │   ├── uow.go                # UnitOfWork interface
│   │   ├── transaction.go        # Transaction implementation
│   │   ├── trace_plugin.go       # GORM tracing plugin
│   │   └── module.go
│   │
│   ├── hypercache/                # Cache abstraction
│   │   ├── cache.go              # Cache interface
│   │   ├── ristretto.go          # In-memory cache (Ristretto)
│   │   ├── redis.go              # Distributed cache (Redis)
│   │   └── module.go
│   │
│   ├── hyperstore/                # Object storage
│   │   ├── store.go              # Store interface
│   │   ├── s3.go                 # S3 implementation
│   │   └── module.go
│   │
│   ├── hypercrypto/               # Encryption utilities
│   │   ├── crypter.go            # Crypter interface
│   │   ├── aes.go                # AES-256-GCM implementation
│   │   └── module.go
│   │
│   ├── hypererror/                # Error handling
│   │   ├── error.go              # Error interface and Code type
│   │   ├── codes.go              # Predefined error codes
│   │   ├── constructors.go       # Convenient error constructors
│   │   └── conversion.go         # HTTP/gRPC status conversion
│   │
│   ├── hypervalidator/            # Request validation
│   │   ├── validator.go          # Validator interface
│   │   ├── playground.go         # go-playground/validator implementation
│   │   └── module.go
│   │
│   ├── hyperhttp/                 # HTTP client
│   │   ├── client.go             # Client interface
│   │   ├── resty.go              # Resty-based implementation
│   │   ├── trace.go              # OpenTelemetry tracing middleware
│   │   └── module.go
│   │
│   ├── hyperweb/                  # Web server (Gin)
│   │   ├── server.go             # Server interface
│   │   ├── gin.go                # Gin implementation
│   │   ├── middleware/           # Built-in middleware
│   │   │   ├── trace.go          # Tracing middleware
│   │   │   ├── recovery.go       # Panic recovery
│   │   │   ├── request_log.go    # Request logging
│   │   │   └── error_handler.go  # Error response conversion
│   │   └── module.go
│   │
│   └── hypergrpc/                 # gRPC server
│       ├── server.go             # Server interface
│       ├── grpc.go               # gRPC implementation
│       ├── interceptor/          # Built-in interceptors
│       │   ├── trace.go          # Tracing interceptor
│       │   ├── recovery.go       # Panic recovery
│       │   ├── request_log.go    # Request logging
│       │   └── error_handler.go  # Error status conversion
│       └── module.go
│
├── examples/                      # Example applications
│   ├── simple-api/               # Minimal REST API example
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   └── repository/
│   │   └── configs/config.yaml
│   │
│   └── fullstack/                # Complete example (Web + gRPC + DB)
│       ├── cmd/server/main.go
│       ├── internal/
│       │   ├── domain/           # Domain models
│       │   ├── handler/          # HTTP handlers
│       │   ├── grpc/             # gRPC services
│       │   ├── service/          # Business logic
│       │   └── repository/       # Data access
│       ├── api/
│       │   ├── proto/            # Protobuf definitions
│       │   └── openapi/          # OpenAPI specs
│       └── configs/config.yaml
│
├── docs/                         # Documentation
│   ├── architecture.md           # Main architecture document
│   ├── architecture/             # Detailed architecture docs
│   │   ├── coding-standards.md
│   │   ├── tech-stack.md
│   │   └── source-tree.md       # This file
│   ├── quick-start.md            # Getting started guide
│   ├── architecture-decisions.md # ADRs
│   └── implementation-plan.md    # Development roadmap
│
├── .golangci.yml                 # Linter configuration
├── Makefile                      # Build and development tasks
├── go.mod
├── go.sum
└── README.md                     # Project overview
```

---

## Application Project Structure

When building applications with Hyperion, follow this standard layout:

```
your-app/
├── cmd/
│   └── server/
│       └── main.go               # Application entry point
│
├── configs/
│   ├── config.yaml               # Base configuration
│   ├── config.dev.yaml           # Development overrides
│   └── config.prod.yaml          # Production overrides
│
├── internal/                     # Private application code
│   ├── domain/                   # Domain models (optional, for complex apps)
│   │   ├── user/
│   │   │   ├── user.go          # Domain entity
│   │   │   ├── repository.go    # Repository interface
│   │   │   └── errors.go        # Domain-specific errors
│   │   └── ...
│   │
│   ├── handler/                  # Presentation layer
│   │   ├── user_handler.go      # HTTP request handlers
│   │   ├── dto.go               # Request/Response DTOs
│   │   └── module.go            # fx.Module for handlers
│   │
│   ├── service/                  # Application service layer
│   │   ├── user_service.go      # Business use cases
│   │   ├── interfaces.go        # Service interfaces
│   │   └── module.go            # fx.Module for services
│   │
│   └── repository/               # Infrastructure layer
│       ├── user_repository.go   # Data access implementation
│       ├── models.go            # Database models (GORM)
│       └── module.go            # fx.Module for repositories
│
├── api/                          # API definitions
│   ├── proto/                    # gRPC Protobuf files
│   │   └── user/v1/
│   │       └── user.proto
│   └── openapi/                  # OpenAPI/Swagger specs
│       └── api.yaml
│
├── pkg/                          # Exportable packages (optional)
│   └── client/                   # API client libraries
│
├── migrations/                   # Database migrations
│   └── 001_create_users.sql
│
├── scripts/                      # Build and deployment scripts
│   ├── build.sh
│   └── deploy.sh
│
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

---

## Package Naming Conventions

### Framework Packages (pkg/hyper*)

All framework core packages follow the `hyper*` naming convention:

| Package | Purpose | Naming Rationale |
|---------|---------|------------------|
| `hyperion` | Framework entry | Greek mythology reference |
| `hyperctx` | Context abstraction | **hyper** + **ctx**(context) |
| `hyperconfig` | Configuration | **hyper** + **config** |
| `hyperlog` | Logging | **hyper** + **log** |
| `hyperdb` | Database | **hyper** + **db** |
| `hypercache` | Cache | **hyper** + **cache** |
| `hyperstore` | Object storage | **hyper** + **store** |
| `hypercrypto` | Encryption | **hyper** + **crypto** |
| `hypererror` | Error handling | **hyper** + **error** |
| `hypervalidator` | Validation | **hyper** + **validator** |
| `hyperhttp` | HTTP client | **hyper** + **http** |
| `hyperweb` | Web server | **hyper** + **web** |
| `hypergrpc` | gRPC server | **hyper** + **grpc** |

### Application Packages (internal/*)

Application packages use descriptive, domain-specific names without prefixes:

- `internal/domain`: Domain entities and business rules
- `internal/handler`: HTTP/gRPC request handlers
- `internal/service`: Application use case implementations
- `internal/repository`: Data access layer

---

## Module Organization

Each package exports an `fx.Module` for dependency injection:

```go
// pkg/hyperlog/module.go
package hyperlog

import "go.uber.org/fx"

var Module = fx.Module("hyperlog",
    fx.Provide(NewZapLogger),
)
```

Application modules follow the same pattern:

```go
// internal/service/module.go
package service

import "go.uber.org/fx"

var Module = fx.Module("service",
    fx.Provide(
        NewUserService,
        NewOrderService,
        // ... more services
    ),
)
```

---

## File Naming Conventions

### Framework Packages

- `{component}.go`: Core interface definition (e.g., `logger.go`, `cache.go`)
- `{impl}.go`: Implementation (e.g., `zap.go`, `ristretto.go`)
- `module.go`: fx.Module provider
- `{feature}.go`: Specific feature implementation (e.g., `trace_plugin.go`)

### Application Packages

- `{entity}_handler.go`: HTTP handlers for entity (e.g., `user_handler.go`)
- `{entity}_service.go`: Business logic for entity (e.g., `user_service.go`)
- `{entity}_repository.go`: Data access for entity (e.g., `user_repository.go`)
- `module.go`: fx.Module provider
- `interfaces.go`: Interface definitions
- `dto.go`: Data transfer objects
- `models.go`: Database models

### Test Files

- `{name}_test.go`: Unit tests
- `{name}_integration_test.go`: Integration tests
- `mock_{name}.go`: Mock implementations

---

## Import Path Structure

### Framework Imports

```go
import (
    "github.com/mapoio/hyperion/pkg/hyperctx"
    "github.com/mapoio/hyperion/pkg/hyperlog"
    "github.com/mapoio/hyperion/pkg/hyperdb"
)
```

### Application Imports

Follow three-group ordering:

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"

    // Third-party
    "go.uber.org/fx"
    "go.uber.org/zap"

    // Local (framework + application)
    "github.com/mapoio/hyperion/pkg/hyperctx"
    "github.com/mapoio/hyperion/pkg/hyperlog"
    "github.com/your-app/internal/domain/user"
    "github.com/your-app/internal/service"
)
```

---

## Navigation Tips

### Finding Component Implementations

1. **Interface Definition**: Always in `{component}.go`
   - Example: `pkg/hyperlog/logger.go` contains `Logger` interface

2. **Default Implementation**: Named after the underlying library
   - Example: `pkg/hyperlog/zap.go` contains Zap-based implementation

3. **Module Registration**: Always in `module.go`
   - Example: `pkg/hyperlog/module.go` exports `Module` variable

### Finding Usage Examples

1. **Simple Examples**: `examples/simple-api/`
   - Minimal setup with basic CRUD operations

2. **Complete Examples**: `examples/fullstack/`
   - Production-like structure with all layers
   - Includes domain modeling, validation, error handling

### Finding Documentation

1. **Package Documentation**: Each package has a `doc.go` file
   - Example: `pkg/hyperlog/doc.go` explains logging capabilities

2. **Architecture Docs**: `docs/architecture/` directory
   - `coding-standards.md`: Development guidelines
   - `tech-stack.md`: Technology choices and rationale
   - `source-tree.md`: This file

---

## Code Generation Locations

### Protobuf Generated Code

```
api/proto/{domain}/{version}/{file}.pb.go        # Message definitions
api/proto/{domain}/{version}/{file}_grpc.pb.go   # gRPC service stubs
```

### Mock Generated Code

```
internal/{layer}/mock_{interface}.go             # Mockery-generated mocks
```

### Migration Files

```
migrations/{timestamp}_{description}.sql         # SQL migrations
migrations/{timestamp}_{description}.go          # Go migrations (GORM)
```

---

## Configuration File Locations

### Framework Configuration

Framework components read configuration from these keys:

```yaml
# Example: configs/config.yaml
log:
  level: info
  encoding: json
  output: stdout

database:
  driver: postgres
  dsn: postgres://localhost/mydb
  max_open_conns: 100

cache:
  type: redis
  address: localhost:6379
```

### Application Configuration

Applications can define custom config sections:

```yaml
# Application-specific config
app:
  name: my-service
  version: 1.0.0

auth:
  jwt_secret: ${JWT_SECRET}
  token_expiry: 24h
```

---

## Testing Structure

### Unit Tests

Located alongside source files:

```
pkg/hyperlog/
├── logger.go
├── zap.go
├── zap_test.go          # Unit tests for Zap logger
└── module_test.go       # Module initialization tests
```

### Integration Tests

Located in separate directory or marked with build tags:

```
pkg/hyperdb/
├── integration_test.go  # Integration tests (database required)
└── testdata/            # Test fixtures
    └── schema.sql
```

### Test Utilities

```
internal/testutil/       # Shared test utilities
├── fixtures.go          # Test data builders
├── assertions.go        # Custom assertions
└── mock_context.go      # Mock hyperctx.Context
```

---

## Development Workflow Locations

### Pre-commit Hooks

```
.git/hooks/
├── commit-msg           # Validates commit message format
└── pre-commit           # Runs linter and tests
```

### CI/CD Configuration

```
.github/workflows/
├── ci.yml               # Continuous integration
├── release.yml          # Release automation
└── deploy.yml           # Deployment pipeline
```

### Build Artifacts

```
build/                   # Build output directory
├── bin/                 # Compiled binaries
│   └── server
└── docker/              # Docker build context
    └── Dockerfile
```

---

**Last Updated**: January 2025
