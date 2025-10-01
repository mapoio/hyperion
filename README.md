# Hyperion

<div align="center">

**A production-ready, microkernel-based Go backend framework built on uber/fx dependency injection**

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-latest-brightgreen.svg)](docs/)

</div>

---

## 🚀 Overview

Hyperion is a modular Go backend framework that provides comprehensive observability, type-safe context management, and declarative transaction handling out of the box. Built on top of `go.uber.org/fx`, it delivers production-ready defaults while maintaining flexibility and extensibility.

### Key Features

- ✅ **Modular Architecture**: All features delivered as independent `fx.Module` packages
- ✅ **Type-Safe Context**: `hyperctx.Context` with integrated tracing, logging, and database access
- ✅ **OpenTelemetry Integration**: Automatic distributed tracing across all architectural layers
- ✅ **Declarative Transactions**: UnitOfWork pattern with seamless transaction propagation
- ✅ **Hot Configuration Reload**: Support for both file-based and remote config sources (Consul/Etcd)
- ✅ **Production-Ready Defaults**: Structured logging, graceful shutdown, and health checks out of the box
- ✅ **Comprehensive Error Handling**: Typed error codes with multi-layer error wrapping

---

## 📦 Quick Start

### Installation

```bash
go get github.com/mapoio/hyperion
```

### Minimal Example

```go
package main

import (
    "github.com/mapoio/hyperion/pkg/hyperion"
    "github.com/mapoio/hyperion/pkg/hyperctx"
    "github.com/mapoio/hyperion/pkg/hyperweb"
    "go.uber.org/fx"
)

func main() {
    fx.New(
        // Import Hyperion web stack
        hyperion.Web(),

        // Register your handlers
        fx.Invoke(registerRoutes),
    ).Run()
}

func registerRoutes(server hyperweb.Server) {
    server.GET("/hello", func(ctx hyperctx.Context) (any, error) {
        ctx.Logger().Info("handling hello request")
        return map[string]string{"message": "Hello, Hyperion!"}, nil
    })
}
```

### Run the Server

```bash
go run cmd/server/main.go
```

Visit `http://localhost:8080/hello` to see the response!

For a complete CRUD application example, see the [Quick Start Guide](docs/quick-start.md).

---

## 🏗️ Architecture

Hyperion follows a **layered architecture** with clear dependency rules:

```
Presentation Layer (hyperweb/hypergrpc)
           ↓
Application Service Layer (internal/service)
           ↓
Domain Layer (internal/domain) - Optional
           ↓
Infrastructure Layer (internal/repository, pkg/*)
```

### Core Components

| Component | Purpose | Documentation |
|-----------|---------|---------------|
| `hyperctx` | Type-safe context with trace/log/db | [Architecture](docs/architecture.md#51-hyperctx---context-abstraction-core) |
| `hyperconfig` | Configuration management | [Tech Stack](docs/architecture/tech-stack.md#configuration-viper) |
| `hyperlog` | Structured logging | [Tech Stack](docs/architecture/tech-stack.md#logging-zap) |
| `hyperdb` | Database + UnitOfWork | [Architecture](docs/architecture.md#hyperdb) |
| `hypererror` | Typed error handling | [Architecture](docs/architecture.md#52-hypererror---error-handling) |
| `hypercache` | Cache abstraction | [Tech Stack](docs/architecture/tech-stack.md#cache-ristretto--redis) |
| `hypervalidator` | Request validation | [Tech Stack](docs/architecture/tech-stack.md#validation-go-playgroundvalidator) |
| `hyperhttp` | HTTP client with tracing | [Tech Stack](docs/architecture/tech-stack.md#http-client-resty) |
| `hyperweb` | Web server (Gin) | [Tech Stack](docs/architecture/tech-stack.md#web-framework-gin) |
| `hypergrpc` | gRPC server | [Architecture](docs/architecture.md#hypergrpc) |

---

## 🎯 Design Principles

1. **Modularity over Monolith**: Import only what you need
2. **Convention over Configuration**: Production-grade defaults
3. **Explicit over Implicit**: Clear dependency declarations
4. **Interface-Driven Design**: Loose coupling and testability
5. **Production-Ready by Default**: Observability built-in

For detailed design rationale, see [Architecture Decisions](docs/architecture-decisions.md).

---

## 📚 Documentation

- **[Architecture Guide](docs/architecture.md)**: Comprehensive framework design document
- **[Quick Start](docs/quick-start.md)**: 10-minute tutorial with complete CRUD example
- **[Coding Standards](docs/architecture/coding-standards.md)**: Development guidelines and best practices
- **[Tech Stack](docs/architecture/tech-stack.md)**: Technology choices and rationale
- **[Source Tree Guide](docs/architecture/source-tree.md)**: Navigate the codebase
- **[Architecture Decisions](docs/architecture-decisions.md)**: ADRs explaining key design choices
- **[Implementation Plan](docs/implementation-plan.md)**: Development roadmap

---

## 🛠️ Technology Stack

| Layer | Technology | Why? |
|-------|------------|------|
| **DI Framework** | go.uber.org/fx | Production-proven lifecycle management |
| **Web Framework** | Gin | High performance, mature ecosystem |
| **RPC Framework** | gRPC | Standard for microservices communication |
| **ORM** | GORM | Most popular Go ORM with plugin support |
| **Configuration** | Viper | Multi-source config with hot reload |
| **Logging** | Zap | Blazing fast structured logging |
| **Tracing** | OpenTelemetry | Industry-standard observability |
| **Cache** | Ristretto/Redis | In-memory + distributed caching |
| **Validation** | go-playground/validator | Tag-based validation |
| **HTTP Client** | Resty | Simple API with built-in retry |

For detailed technology rationale, see [Tech Stack Documentation](docs/architecture/tech-stack.md).

---

## 🚦 Development Workflow

### Prerequisites

- Go 1.21+ (1.22+ recommended)
- PostgreSQL/MySQL (for database examples)
- Redis (optional, for cache examples)

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/mapoio/hyperion.git
cd hyperion

# Install development dependencies and Git hooks
make setup

# Run tests
make test

# Run linter
make lint
```

### Common Commands

```bash
# Format code
make fmt

# Run tests with coverage
make test-coverage

# Run linter with auto-fix
make lint-fix

# Full verification (format + lint + test)
make verify
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

### Framework Structure

```
hyperion/
├── pkg/                    # Framework core components
│   ├── hyperion/          # Framework entry point
│   ├── hyperctx/          # Context abstraction
│   ├── hyperlog/          # Structured logging
│   ├── hyperdb/           # Database + UnitOfWork
│   └── ...                # Other core components
├── examples/              # Example applications
├── docs/                  # Documentation
└── README.md
```

### Application Structure

```
your-app/
├── cmd/server/main.go     # Entry point
├── internal/
│   ├── handler/          # HTTP/gRPC handlers
│   ├── service/          # Business logic
│   └── repository/       # Data access
├── configs/config.yaml   # Configuration
└── go.mod
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

## 🗺️ Roadmap

### v1.0 (Current)
- ✅ Core framework design complete
- ⏳ Implementation in progress

### v1.1 (Q1 2025)
- OpenTelemetry exporter configuration
- Prometheus metrics integration
- Production examples

### v1.2 (Q2 2025)
- Message queue component (`hypermq`)
- Enhanced object storage
- Authentication/Authorization helpers

### v2.0 (Q4 2025)
- Generic repository and service patterns
- Code generation tools
- Admin dashboard

For detailed implementation plan, see [Implementation Plan](docs/implementation-plan.md).

---

## 🤝 Contributing

We welcome contributions! Before submitting a PR, please ensure:

- [ ] Code follows [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [ ] All tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Code coverage ≥ 90%
- [ ] Documentation is updated
- [ ] Commit messages follow convention

For detailed contribution guidelines, see [Coding Standards](docs/architecture/coding-standards.md).

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
