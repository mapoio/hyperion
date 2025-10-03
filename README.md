# Hyperion

<div align="center">

**A production-ready, microkernel-based Go backend framework built on uber/fx dependency injection**

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-latest-brightgreen.svg)](docs/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

**Architecture**: v2.0 Monorepo | **Status**: Epic 2 In Progress

</div>

---

## 🚀 Overview

Hyperion is a **zero lock-in** Go backend framework built on the **core-adapter pattern**. The ultra-lightweight core (`hyperion/`) defines pure interfaces with **zero third-party dependencies**, while independent adapters (`adapter/`) provide swappable implementations. Built on top of `go.uber.org/fx`, it delivers production-ready defaults while ensuring you're never locked into any specific technology.

### Architecture Philosophy (v2.0)

**Core-Adapter Pattern**: Hyperion's revolutionary architecture separates interface definitions from implementations:

- **🎯 Core (`hyperion/`)**: Pure Go interfaces with zero dependencies (except `fx`)
- **🔌 Adapters (`adapter/`)**: Independent, swappable implementations
- **📦 Monorepo Workspace**: Unified development with independent versioning

**Why This Matters**: You can replace Zap with Logrus, GORM with sqlx, or Viper with your own config—without touching application code.

### Key Features

- ✅ **Zero Lock-In**: Core interfaces with NoOp implementations, swap adapters at will
- ✅ **Modular Architecture**: All features delivered as independent `fx.Module` packages
- ✅ **Type-Safe Context**: `hyperion.Context` with integrated tracing, logging, metrics, and database access
- ✅ **Interceptor Pattern**: 3-line pattern for automatic tracing, logging, and metrics
- ✅ **Unified Observability**: Automatic correlation between Logs, Traces, and Metrics via OpenTelemetry
- ✅ **Production-Ready Adapters**: Viper (config), Zap (logging), GORM (database) with 80%+ test coverage
- ✅ **Declarative Transactions**: UnitOfWork pattern with automatic commit/rollback and panic recovery
- ✅ **Hot Configuration Reload**: Viper-based config with file watching support
- ✅ **Transaction Propagation**: Type-safe context-based transaction propagation via `hyperion.WithDB()`
- ✅ **Interface-Driven Design**: Every component is mockable and testable

---

## 📦 Quick Start

### Installation

```bash
# Add to your go.mod
go get github.com/mapoio/hyperion/hyperion
go get github.com/mapoio/hyperion/adapter/viper  # Configuration
go get github.com/mapoio/hyperion/adapter/zap    # Logging
go get github.com/mapoio/hyperion/adapter/gorm   # Database (optional)
```

### Minimal Example

```go
package main

import (
    "go.uber.org/fx"

    "github.com/mapoio/hyperion/hyperion"
    "github.com/mapoio/hyperion/adapter/viper"
    "github.com/mapoio/hyperion/adapter/zap"
)

func main() {
    fx.New(
        // Core provides interface definitions
        hyperion.CoreModule,

        // Adapters provide implementations
        viper.Module,  // Config from files/env
        zap.Module,    // Structured logging

        // Your application logic
        fx.Invoke(run),
    ).Run()
}

func run(logger hyperion.Logger, cfg hyperion.Config) {
    logger.Info("application started",
        "env", cfg.GetString("app.env"),
        "version", "1.0.0",
    )
}
```

### Configuration (config.yaml)

```yaml
app:
  env: production

log:
  level: info
  encoding: json
  output: stdout
```

### Run the Application

```bash
go run main.go
```

For a complete CRUD application example with HTTP server, see the [Quick Start Guide](docs/quick-start.md) (coming soon in Epic 2.4).

---

## 🏗️ Architecture

### v2.0 Monorepo Structure

```
hyperion/                          # Monorepo root
├── go.work                        # Go workspace definition
├── QUICK_START.md                 # Quick start guide
├── docs/                          # 📚 Documentation
│   ├── interceptor.md             # Interceptor pattern guide
│   ├── observability.md           # Observability guide
│   └── architecture.md            # Architecture documentation
│
├── hyperion/                      # 🎯 Core (zero dependencies)
│   ├── go.mod                     # Only depends on: go.uber.org/fx
│   ├── README.md                  # Core library documentation
│   ├── logger.go                  # Logger interface
│   ├── config.go                  # Config interface
│   ├── database.go                # Database interface
│   ├── tracer.go                  # Tracer interface
│   ├── metric.go                  # Meter interface
│   ├── cache.go                   # Cache interface
│   ├── context.go                 # Context interface
│   └── interceptor.go             # Interceptor interface
│
└── adapter/                       # 🔌 Adapters (independent modules)
    ├── viper/                     # ✅ Config adapter (Implemented)
    │   ├── go.mod                 # Depends on: spf13/viper
    │   └── provider.go
    │
    ├── zap/                       # ✅ Logger adapter (Implemented)
    │   ├── go.mod                 # Depends on: uber-go/zap
    │   ├── logger.go
    │   └── module.go
    │
    ├── gorm/                      # ✅ Database adapter (Implemented)
    │   ├── go.mod                 # Depends on: gorm.io/gorm
    │   ├── database.go
    │   ├── executor.go
    │   ├── unit_of_work.go
    │   └── module.go
    │
    ├── otel/                      # 🔜 Tracer adapter (Planned)
    ├── ristretto/                 # 🔜 Cache adapter (Planned)
    └── redis/                     # 🔜 Cache adapter (Planned)
```

### Core Interfaces

| Interface | Status | Adapter | Documentation |
|-----------|--------|---------|---------------|
| `Config` | ✅ Implemented | [adapter/viper](adapter/viper) | Configuration with file watching |
| `ConfigWatcher` | ✅ Implemented | [adapter/viper](adapter/viper) | Hot config reload |
| `Logger` | ✅ Implemented | [adapter/zap](adapter/zap) | Structured logging with Zap |
| `Database` | ✅ Implemented | [adapter/gorm](adapter/gorm) | Database access with GORM |
| `Executor` | ✅ Implemented | [adapter/gorm](adapter/gorm) | Query execution with transaction tracking |
| `UnitOfWork` | ✅ Implemented | [adapter/gorm](adapter/gorm) | Declarative transaction management |
| `Tracer` | ✅ Implemented | [hyperion/tracer.go](hyperion/tracer.go) | Distributed tracing (NoOp default) |
| `Meter` | ✅ Implemented | [hyperion/metric.go](hyperion/metric.go) | Metrics collection (NoOp default) |
| `Interceptor` | ✅ Implemented | [hyperion/interceptor.go](hyperion/interceptor.go) | Cross-cutting concerns pattern |
| `Cache` | 🔜 Planned | `adapter/ristretto` | In-memory caching |
| `Context` | ✅ Implemented | [hyperion/context.go](hyperion/context.go) | Type-safe request context |

---

## 🎯 Design Principles

1. **Zero Lock-In**: Core defines interfaces, adapters are swappable
2. **Interface-Driven Design**: Every dependency is an interface
3. **Modularity over Monolith**: Independent modules with independent versioning
4. **Convention over Configuration**: Production-grade defaults with override capability
5. **Explicit over Implicit**: Clear dependency declarations via fx
6. **Production-Ready by Default**: All adapters ship with 90%+ test coverage

For detailed design rationale, see [Architecture Decisions](docs/architecture-decisions.md).

---

## 📚 Documentation

### Core Documentation
- **[Quick Start Guide](QUICK_START.md)**: 5-minute tutorial with complete CRUD example
- **[Hyperion Core README](hyperion/README.md)**: Core library overview and usage patterns
- **[Interceptor Guide](docs/interceptor.md)**: Complete interceptor pattern documentation
- **[Observability Guide](docs/observability.md)**: Unified observability with Logs, Traces, and Metrics
- **[Architecture Guide](docs/architecture.md)**: Comprehensive framework design document
- **[Coding Standards](docs/architecture/coding-standards.md)**: Development guidelines and best practices
- **[Tech Stack](docs/architecture/tech-stack.md)**: Technology choices and rationale
- **[Source Tree Guide](docs/architecture/source-tree.md)**: Navigate the codebase
- **[Architecture Decisions](docs/architecture-decisions.md)**: ADRs explaining key design choices
- **[Implementation Plan](docs/implementation-plan.md)**: Development roadmap

### Design Documents
- **[Interceptor Architecture](.design/interceptor-architecture.md)**: Deep dive into interceptor pattern design
- **[Observability Architecture](.design/observability-architecture.md)**: Deep dive into observability correlation design

### Adapter Documentation
- **[Adapter Overview](docs/adapters)**: Complete guide to all official adapters
- **[Viper Adapter](adapter/viper/README.md)**: Configuration management guide
- **[Zap Adapter](adapter/zap/README.md)**: Structured logging guide
- **[GORM Adapter](adapter/gorm/README.md)**: Database access and transactions guide
- **[Implementation Reports](docs/adapters/reports)**: Detailed implementation metrics and decisions

---

## 🛠️ Current Adapter Implementations

| Adapter | Status | Version | Test Coverage | Purpose |
|---------|--------|---------|---------------|---------|
| **[Viper](adapter/viper)** | ✅ Implemented | v1.20.0 | 84.4% | Config management with hot reload |
| **[Zap](adapter/zap)** | ✅ Implemented | v1.27.0 | 93.9% | High-performance structured logging |
| **[GORM](adapter/gorm)** | ✅ Implemented | v1.25.12 | 82.1% | Database access with declarative transactions |
| **OpenTelemetry** | 🔜 Planned | v1.33.0+ | - | Distributed tracing |
| **Ristretto** | 🔜 Planned | v1.3.0+ | - | In-memory caching |
| **Redis** | 🔜 Planned | v9.0.0+ | - | Distributed caching |

### Why These Technologies?

- **Viper**: De-facto standard for Go configuration with hot reload support
- **Zap**: Blazing fast (1M+ logs/sec), zero-allocation structured logging
- **GORM**: Most popular Go ORM with excellent plugin ecosystem
- **OpenTelemetry**: Industry-standard observability framework
- **Ristretto**: High-performance, concurrent in-memory cache
- **Redis**: Battle-tested distributed cache and data store

For detailed technology rationale, see [Tech Stack Documentation](docs/architecture/tech-stack.md).

---

## 🚦 Development Workflow

### Prerequisites

- **Go 1.24+** (required for workspace features)
- Git with hooks support
- PostgreSQL/MySQL (for database adapter testing, optional)
- Redis (for cache adapter testing, optional)

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/mapoio/hyperion.git
cd hyperion

# Install development dependencies and Git hooks
make setup

# Verify workspace setup
go work sync

# Run all tests across all modules
make test

# Run linter
make lint
```

### Working with the Monorepo

```bash
# Test a specific adapter
cd adapter/zap && go test -v ./...

# Test the core
cd hyperion && go test -v ./...

# Test everything with coverage
make test-coverage

# Format all code
make fmt

# Run full verification (format + lint + test)
make verify
```

### Adding a New Adapter

```bash
# Create adapter directory
mkdir -p adapter/newadapter

# Initialize module
cd adapter/newadapter
go mod init github.com/mapoio/hyperion/adapter/newadapter

# Add to workspace
cd ../..
go work use ./adapter/newadapter
```

For complete development guidelines, see [Coding Standards](docs/architecture/coding-standards.md).

---

## 📝 Git Commit Standards

This project follows the [AngularJS Commit Message Convention](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#commit).

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Examples

```bash
# Feature
feat(hyperlog): add file rotation support

# Bug fix
fix(hyperdb): correct transaction rollback handling

# Documentation
docs: update quick start guide

# Refactoring
refactor(hypererror): simplify error wrapping logic
```

Git hooks will automatically validate your commit messages. For details, see [Coding Standards - Git Commit Standards](docs/architecture/coding-standards.md#git-commit-standards).

---

## 🗂️ Project Structure

### Monorepo Structure (v2.0)

```
hyperion/                          # Monorepo root
├── go.work                        # Workspace definition
├── go.work.sum                    # Workspace checksums
├── Makefile                       # Unified build system
├── .golangci.yml                  # Shared linter config
│
├── hyperion/                      # 🎯 Core library
│   ├── go.mod                     # Minimal deps (fx only)
│   ├── README.md                  # Core library documentation
│   ├── logger.go                  # Logger interface + NoOp
│   ├── config.go                  # Config interface + NoOp
│   ├── database.go                # Database interface + NoOp
│   ├── tracer.go                  # Tracer interface + NoOp
│   ├── metric.go                  # Meter interface + NoOp
│   ├── cache.go                   # Cache interface + NoOp
│   ├── context.go                 # Context interface
│   ├── interceptor.go             # Interceptor interface
│   ├── module.go                  # CoreModule definition
│   └── defaults.go                # Default NoOp providers
│
├── adapter/                       # 🔌 Adapter implementations
│   ├── viper/                     # Config adapter
│   │   ├── go.mod                 # Independent versioning
│   │   ├── provider.go
│   │   ├── module.go
│   │   └── provider_test.go
│   │
│   ├── zap/                       # Logger adapter
│   │   ├── go.mod
│   │   ├── logger.go
│   │   ├── module.go
│   │   ├── logger_test.go
│   │   └── integration_test.go
│   │
│   └── ...                        # Other adapters
│
├── docs/                          # Documentation
│   ├── prd/                       # Product requirements
│   ├── stories/                   # User stories
│   └── architecture/              # Technical docs
│
└── .github/                       # CI/CD workflows
    ├── workflows/
    │   └── pr-checks.yml         # Automated testing
    └── labeler.yml               # Auto-labeling
```

### Application Structure (Recommended)

```
your-app/
├── cmd/server/main.go     # Entry point
├── internal/
│   ├── handler/          # HTTP/gRPC handlers
│   ├── service/          # Business logic
│   └── repository/       # Data access
├── configs/config.yaml   # Configuration
└── go.mod                # Dependencies
```

For detailed structure guide, see [Source Tree Guide](docs/architecture/source-tree.md).

---

## 🧪 Testing

Hyperion emphasizes comprehensive testing with clear separation of concerns:

```bash
# Run all tests
make test

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./pkg/hyperlog/...

# Generate coverage report
make test-coverage
```

### Testing Guidelines

- **Unit Tests**: Mock dependencies using interfaces
- **Integration Tests**: Use real implementations with Docker containers
- **Table-Driven Tests**: Recommended for testing multiple scenarios
- **Test Helpers**: Mark with `t.Helper()` for better error reporting

For testing best practices, see [Architecture Guide - Testing Strategy](docs/architecture.md#8-testing-strategy).

---

## 🗺️ Development Status & Roadmap

### Current Phase: **Epic 2 - Essential Adapters** (v2.1)

**Progress**: 🟢🟢🟢🟢⚪⚪ (3/5 stories completed)

| Story | Status | Deliverable | Completion |
|-------|--------|-------------|------------|
| 2.0 | ✅ Complete | v2.0 Monorepo Migration | Oct 2, 2025 |
| 2.1 | ✅ Complete | Zap Logger Adapter (93.9% coverage) | Oct 2, 2025 |
| 2.2 | ✅ Complete | GORM Database Adapter (82.1% coverage) | Oct 2, 2025 |
| 2.3 | 🔜 Planned | Ristretto Cache Adapter | Dec 2025 |
| 2.4 | 🔜 Planned | Example CRUD Application | Dec 2025 |

### Epic Overview

**✅ Epic 1: Core Interfaces** (Completed Sept 2025)
- Zero-dependency core with pure interfaces
- NoOp implementations for all interfaces
- fx.Module integration
- Comprehensive documentation

**✅ Epic 2: Essential Adapters** (60% Complete)
- ✅ Viper adapter (Config + ConfigWatcher, 84.4% coverage)
- ✅ Zap adapter (Logger, 93.9% coverage)
- ✅ GORM adapter (Database + Executor + UnitOfWork, 82.1% coverage)
- 🔜 Ristretto adapter (Cache, planned)
- 🔜 Example CRUD Application (planned)

**🔜 Epic 3: Observability** (Planned Q1 2026)
- OpenTelemetry tracer adapter
- Metrics collection
- Distributed tracing examples

**🔜 Epic 4: Web & RPC** (Planned Q2 2026)
- HTTP server framework integration
- gRPC server support
- Middleware/interceptor system

For detailed implementation plan, see [Implementation Plan](docs/implementation-plan.md).

---

## 🤝 Contributing

We welcome contributions! Before submitting a PR, please ensure:

### Code Quality Checklist

- [ ] Code follows [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [ ] All tests pass (`make test`) across all affected modules
- [ ] Linter passes (`make lint`) with zero warnings
- [ ] Test coverage ≥ 80% (90%+ for core components)
- [ ] No race conditions (`go test -race ./...`)
- [ ] Documentation is updated (godoc + README)
- [ ] Commit messages follow [AngularJS Convention](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#commit)
- [ ] PR targets the `develop` branch (not `main`)

### Contribution Types

**Bug Fixes**: Target `develop` branch with `fix(scope):` commits

**New Features**: Target `develop` branch with `feat(scope):` commits

**New Adapters**: Follow the [adapter implementation guide](docs/architecture/source-tree.md#adding-a-new-adapter)

**Documentation**: Target `develop` branch with `docs:` commits

For detailed contribution guidelines, see [Coding Standards](docs/architecture/coding-standards.md).

### PR Review Process

1. Automated CI checks must pass (tests, lint, coverage)
2. Code review by at least one maintainer
3. All conversations must be resolved
4. Squash and merge to `develop`
5. Release to `main` happens at epic completion

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

Hyperion is built on the shoulders of giants:

- [uber-go/fx](https://github.com/uber-go/fx): Dependency injection and lifecycle management
- [gin-gonic/gin](https://github.com/gin-gonic/gin): High-performance HTTP framework
- [uber-go/zap](https://github.com/uber-go/zap): Blazing fast structured logging
- [go-gorm/gorm](https://github.com/go-gorm/gorm): Comprehensive ORM
- [open-telemetry/opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go): Observability framework

---

## 📞 Contact & Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/mapoio/hyperion/issues)
- **Discussions**: [GitHub Discussions](https://github.com/mapoio/hyperion/discussions)

---

<div align="center">

**Built with ❤️ for the Go community**

</div>
