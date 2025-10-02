# Hyperion Framework Source Tree Guide

**Version**: 2.0
**Date**: October 2025
**Status**: Updated for Monorepo Architecture

This document explains the organization of the Hyperion framework source code and provides guidelines for navigating the codebase.

---

## Framework Source Structure (v2.0 Monorepo)

```
hyperion/                          # Monorepo root
├── go.work                        # Go workspace definition
├── Makefile                       # Unified build system
├── .golangci.yml                  # Linter configuration
│
├── hyperion/                      # 🎯 Core Library (ZERO 3rd-party deps)
│   ├── go.mod                     # Dependencies: go.uber.org/fx ONLY
│   ├── go.sum
│   │
│   ├── logger.go                  # Logger interface
│   ├── logger_noop.go             # NoOp Logger implementation
│   │
│   ├── tracer.go                  # Tracer interface (OTel-compatible)
│   ├── tracer_noop.go             # NoOp Tracer implementation
│   │
│   ├── database.go                # Database + Executor interfaces
│   ├── database_noop.go           # NoOp Database implementation
│   │
│   ├── config.go                  # Config + ConfigWatcher interfaces
│   ├── config_noop.go             # NoOp Config implementation
│   │
│   ├── cache.go                   # Cache interface
│   ├── cache_noop.go              # NoOp Cache implementation
│   │
│   ├── context.go                 # Context interface (type-safe)
│   ├── defaults.go                # Default modules (NoOp providers)
│   ├── module.go                  # CoreModule definitions
│   └── hyperion_test.go           # Core tests
│
└── adapter/                       # 🔌 Adapter Implementations
    │
    ├── viper/                     # ✅ Config Adapter (Implemented)
    │   ├── go.mod                 # Independent module
    │   ├── go.sum
    │   ├── provider.go            # ConfigWatcher implementation
    │   ├── module.go              # fx.Module export
    │   └── provider_test.go       # Unit tests
    │
    ├── zap/                       # 🔜 Logger Adapter (Planned)
    │   ├── go.mod
    │   ├── logger.go              # Zap-based Logger
    │   ├── module.go
    │   └── logger_test.go
    │
    ├── otel/                      # 🔜 Tracer Adapter (Planned)
    │   ├── go.mod
    │   ├── tracer.go              # OpenTelemetry integration
    │   ├── module.go
    │   └── tracer_test.go
    │
    ├── gorm/                      # 🔜 Database Adapter (Planned)
    │   ├── go.mod
    │   ├── database.go            # GORM integration
    │   ├── unit_of_work.go        # Transaction management
    │   ├── module.go
    │   └── database_test.go
    │
    ├── ristretto/                 # 🔜 In-Memory Cache (Planned)
    │   ├── go.mod
    │   ├── cache.go
    │   └── module.go
    │
    └── redis/                     # 🔜 Distributed Cache (Planned)
        ├── go.mod
        ├── cache.go
        └── module.go
```

---

## Key Architectural Changes (v1.0 → v2.0)

| Aspect | v1.0 | v2.0 |
|--------|------|------|
| **Structure** | Single module `pkg/hyper*` | Monorepo with `hyperion/` + `adapter/*` |
| **Dependencies** | Bundled implementations | Core: zero deps, Adapters: specific deps |
| **Modules** | One `go.mod` | Multiple independent `go.mod` files |
| **Versioning** | Monolithic | Independent per module |
| **NoOp Location** | Separate package | Same package as interface |

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
│       ├── models.go            # Database models
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

### Framework Core (hyperion/)

All core interfaces use simple, descriptive names:

| File | Purpose | Naming Rationale |
|------|---------|------------------|
| `logger.go` | Logger interface | Core capability |
| `tracer.go` | Tracer interface | Distributed tracing |
| `database.go` | Database interface | Data access |
| `config.go` | Configuration interface | Config management |
| `cache.go` | Cache interface | Caching layer |
| `context.go` | Context interface | Type-safe context |

### Adapter Packages (adapter/*)

Adapters are named after the underlying library:

| Adapter | Implements | Library Used |
|---------|------------|--------------|
| `viper` | Config/ConfigWatcher | github.com/spf13/viper |
| `zap` | Logger | go.uber.org/zap |
| `otel` | Tracer | go.opentelemetry.io/otel |
| `gorm` | Database | gorm.io/gorm |
| `ristretto` | Cache | github.com/dgraph-io/ristretto |
| `redis` | Cache | github.com/redis/go-redis |

### Application Packages (internal/*)

Application packages use descriptive, domain-specific names without prefixes:

- `internal/domain`: Domain entities and business rules
- `internal/handler`: HTTP/gRPC request handlers
- `internal/service`: Application use case implementations
- `internal/repository`: Data access layer

---

## Module Organization

### Core Module Structure

The core library exports fx modules for dependency injection:

```go
// hyperion/module.go
package hyperion

import "go.uber.org/fx"

// CoreModule - Provides all interfaces with NoOp defaults
var CoreModule = fx.Module("hyperion.core",
    fx.Options(
        DefaultLoggerModule,
        DefaultTracerModule,
        DefaultDatabaseModule,
        DefaultConfigModule,
        DefaultCacheModule,
    ),
)

// CoreWithoutDefaultsModule - Strict mode, requires all adapters
var CoreWithoutDefaultsModule = fx.Module("hyperion.core.minimal",
    // No default implementations
)
```

### Adapter Module Pattern

Each adapter exports a module that provides interface implementation:

```go
// adapter/viper/module.go
package viper

import (
    "go.uber.org/fx"
    "github.com/mapoio/hyperion"
)

var Module = fx.Module("hyperion.adapter.viper",
    fx.Provide(
        fx.Annotate(
            NewProviderFromEnv,
            fx.As(new(hyperion.Config)),
            fx.As(new(hyperion.ConfigWatcher)),
        ),
    ),
)
```

### Application Module Pattern

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

### Framework Core

- `{interface}.go`: Core interface definition (e.g., `logger.go`, `cache.go`)
- `{interface}_noop.go`: NoOp implementation (e.g., `logger_noop.go`)
- `module.go`: fx.Module provider
- `defaults.go`: Default module definitions
- `{interface}_test.go`: Unit tests

**Example**:
```
hyperion/
├── logger.go          # Logger interface
├── logger_noop.go     # NoOp implementation
├── logger_test.go     # Tests
```

### Adapter Packages

- `{impl}.go`: Implementation file (e.g., `provider.go` for Viper, `logger.go` for Zap)
- `module.go`: fx.Module export
- `{impl}_test.go`: Unit tests
- `integration_test.go`: Integration tests (if applicable)

**Example**:
```
adapter/viper/
├── provider.go        # Viper implementation
├── module.go          # fx.Module
├── provider_test.go   # Unit tests
```

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
- `mock_{name}.go`: Mock implementations (or use mockery)

---

## Import Path Structure

### Framework Imports (v2.0)

```go
import (
    // Core library (interfaces)
    "github.com/mapoio/hyperion"

    // Adapters (implementations)
    "github.com/mapoio/hyperion/adapter/viper"
    "github.com/mapoio/hyperion/adapter/zap"
    "github.com/mapoio/hyperion/adapter/otel"
    "github.com/mapoio/hyperion/adapter/gorm"
)
```

**Note**: In v2.0, there is NO `pkg/` prefix. Import directly from `github.com/mapoio/hyperion`.

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

    // Framework + Application
    "github.com/mapoio/hyperion"
    "github.com/mapoio/hyperion/adapter/viper"
    "github.com/your-app/internal/domain/user"
    "github.com/your-app/internal/service"
)
```

---

## Navigation Tips

### Finding Component Implementations

1. **Interface Definition**: Always in `hyperion/{component}.go`
   - Example: `hyperion/logger.go` contains `Logger` interface

2. **NoOp Implementation**: In `hyperion/{component}_noop.go`
   - Example: `hyperion/logger_noop.go` contains NoOp Logger

3. **Adapter Implementation**: In `adapter/{name}/`
   - Example: `adapter/viper/provider.go` contains Viper-based Config

4. **Module Registration**: Always in `module.go`
   - Core: `hyperion/module.go` exports `CoreModule`
   - Adapters: `adapter/{name}/module.go` exports `Module`

### Finding Usage Examples

**Note**: Examples are planned but not yet implemented. Current reference:

1. **Core Tests**: `hyperion/hyperion_test.go`
   - Shows how to use CoreModule
   - Demonstrates NoOp implementations

2. **Adapter Tests**: `adapter/viper/provider_test.go`
   - Shows how to use Viper adapter
   - Demonstrates configuration loading

**Planned Examples**:
1. **Simple Examples**: `examples/simple-api/` (planned)
   - Minimal setup with basic CRUD operations

2. **Complete Examples**: `examples/fullstack/` (planned)
   - Production-like structure with all layers
   - Includes domain modeling, validation, error handling

### Finding Documentation

1. **Main Architecture**: `docs/architecture.md`
   - Complete v2.0 architecture documentation (2531 lines)
   - Source of truth for all architectural decisions

2. **Architecture Decisions**: `docs/architecture-decisions.md`
   - ADRs explaining key design choices
   - Rationale for v2.0 changes

3. **Implementation Review**: `docs/architecture-review-v2.md`
   - Detailed review of v2.0 implementation
   - Technical implementation notes

4. **Additional Docs**: `docs/architecture/` directory
   - `coding-standards.md`: Development guidelines
   - `tech-stack.md`: Technology choices and rationale
   - `source-tree.md`: This file

---

## Workspace Management

### Go Workspace Commands

```bash
# Sync workspace (after adding new modules)
go work sync

# Build all modules
go build ./...

# Test all modules
go test ./...

# Update workspace to include new module
go work use ./adapter/newadapter
```

### Adding a New Adapter

1. Create adapter directory: `mkdir -p adapter/newadapter`
2. Initialize module: `cd adapter/newadapter && go mod init github.com/mapoio/hyperion/adapter/newadapter`
3. Add to workspace: `go work use ./adapter/newadapter`
4. Implement interface and module
5. Update root Makefile if needed

---

## Configuration File Locations

### Framework Configuration

Framework components read configuration from standard keys:

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

### Core Library Tests

Located alongside source files:

```
hyperion/
├── logger.go
├── logger_noop.go
├── logger_test.go      # Tests for Logger interface and NoOp
├── module.go
└── module_test.go      # Tests for module system
```

### Adapter Tests

Each adapter has its own tests:

```
adapter/viper/
├── provider.go
├── provider_test.go    # Unit tests
└── integration_test.go # Integration tests (file watching, etc.)
```

### Test Utilities

```
internal/testutil/       # Shared test utilities (planned)
├── fixtures.go          # Test data builders
├── assertions.go        # Custom assertions
└── mock_context.go      # Mock hyperion.Context
```

---

## Development Workflow Locations

### Build System

**Root Makefile**: Unified build system for all modules

```makefile
# All workspace modules
MODULES := hyperion adapter/viper

.PHONY: test
test: ## Run tests across all modules
	@for module in $(MODULES); do \
		echo "Testing $$module..."; \
		(cd $$module && go test -v -race ./...) || exit 1; \
	done
```

### CI/CD Configuration (Planned)

```
.github/workflows/
├── ci.yml               # Continuous integration
├── release.yml          # Release automation
└── deploy.yml           # Deployment pipeline
```

### Build Artifacts

```
build/                   # Build output directory (planned)
├── bin/                 # Compiled binaries
└── docker/              # Docker build context
```

---

## Version Comparison

### v1.0 Structure (Deprecated)

```
hyperion/
└── pkg/
    ├── hyperion/
    ├── hyperctx/
    ├── hyperlog/
    ├── hyperdb/
    └── ...
```

### v2.0 Structure (Current)

```
hyperion/
├── hyperion/           # Core interfaces + NoOp
└── adapter/
    ├── viper/          # Config adapter
    ├── zap/            # Logger adapter (planned)
    └── ...
```

**Key Differences**:
- No `pkg/` prefix in v2.0
- Adapters are separate modules
- NoOp implementations in same package as interfaces
- Independent versioning per adapter

---

**Last Updated**: October 2025
**Version**: 2.0 (Monorepo Architecture)
