# Hyperion Technology Stack

**Version**: 2.0
**Date**: October 2025
**Status**: Updated for Core-Adapter Architecture

This document provides detailed information about the technology choices made for Hyperion framework.

---

## Core Technology Philosophy (v2.0)

**Zero Lock-in Principle**: The core library (`hyperion/`) has ZERO 3rd-party dependencies except `go.uber.org/fx`. All concrete implementations are provided via optional adapters.

```
Core Library (hyperion/)
    ↓ depends on
go.uber.org/fx ONLY
    ↓
  (NO other dependencies)

Adapters (adapter/*)
    ↓ depends on
Specific implementations (Viper, Zap, GORM, etc.)
```

---

## Core Technologies

### Dependency Injection: go.uber.org/fx

**Status**: ✅ Core Dependency (ONLY dependency in core library)

**Why fx?**
- Production-proven at Uber scale
- Explicit dependency declaration through function signatures
- Built-in lifecycle management (OnStart/OnStop hooks)
- Native support for modular architecture via `fx.Module`
- Compile-time dependency resolution with clear error messages
- Visualization support for dependency graphs

**Alternatives Considered:**
- **Wire (Google)**: Compile-time DI but lacks lifecycle management
- **Dig (Uber)**: Low-level library that fx builds upon
- **Manual DI**: Maximum flexibility but high maintenance cost

**v2.0 Impact**: fx is the ONLY dependency in `hyperion/go.mod`, enabling true zero lock-in.

---

## Adapter Technologies

### Configuration: Viper (adapter/viper)

**Status**: ✅ Implemented

**Why Viper?**
- Multi-source configuration (files, env vars, remote)
- Hot reload support via file watching
- Wide format support (YAML, JSON, TOML, etc.)
- Remote config support (Consul, etcd)
- De facto standard in Go ecosystem

**Alternatives Considered:**
- **envconfig**: Simple but limited to environment variables
- **koanf**: Modern alternative, smaller community
- Custom solution: Higher maintenance burden

**Implementation**:
- **Package**: `github.com/mapoio/hyperion/adapter/viper`
- **Implements**: `hyperion.Config` and `hyperion.ConfigWatcher`
- **Dependencies**: `github.com/spf13/viper v1.21.0`

---

### Logging: Zap (adapter/zap)

**Status**: 🔜 Planned

**Why Zap?**
- Blazing fast structured logging
- Zero allocation logging paths
- Flexible output configuration
- Production-proven at Uber
- Dynamic level adjustment support

**Alternatives Considered:**
- **Logrus**: Slower performance, older design
- **Zerolog**: Fast but less flexible
- Standard library: Insufficient for production

**Planned Implementation**:
- **Package**: `github.com/mapoio/hyperion/adapter/zap`
- **Implements**: `hyperion.Logger`
- **Dependencies**: `go.uber.org/zap v1.27.0`

---

### Observability: OpenTelemetry (adapter/otel)

**Status**: 🔜 Planned for full integration (Epic 3, Q1 2026)
**Current**: ✅ Core interfaces implemented (Tracer, Meter), NoOp defaults available

**Why OpenTelemetry?**
- Industry-standard observability framework
- Vendor-neutral (works with Jaeger, Zipkin, Prometheus, etc.)
- Comprehensive SDK with auto-instrumentation
- Active CNCF project with strong community
- Built-in support for traces, metrics, and logs
- **Automatic correlation** between all three pillars

**Alternatives Considered:**
- **OpenTracing**: Deprecated in favor of OpenTelemetry
- **Jaeger Client**: Vendor-specific, traces only
- **Prometheus Client**: Metrics only, no trace correlation
- Custom observability: Reinventing the wheel

**Core Design**:
- Core library provides OTel-compatible interfaces WITHOUT depending on OTel
- Adapter wraps actual OpenTelemetry SDK
- **Interceptor Pattern**: Automatic tracing, logging, and metrics via `ctx.UseIntercept()`
- **Automatic Correlation**: TraceID and SpanID shared across logs, traces, and metrics

**Current Implementation (v2.0-2.2)**:
- **Interfaces**: `hyperion.Tracer`, `hyperion.Meter` (OTel-compatible)
- **Built-in Interceptors**: `TracingInterceptor`, `LoggingInterceptor`
- **NoOp Defaults**: Zero overhead when OTel adapter not installed
- **Zap Logger**: OTel Logs Bridge for automatic trace context

**Planned Full OTel Integration (Epic 3)**:
- **Package**: `github.com/mapoio/hyperion/adapter/otel`
- **Implements**: `hyperion.Tracer` + `hyperion.Meter`
- **Dependencies**: `go.opentelemetry.io/otel v1.24.0`
- **Features**:
  - Real distributed tracing with span export
  - Real metrics collection with exemplar support
  - Automatic trace correlation for logs (via OTel Logs Bridge)
  - Exporters: Jaeger, Prometheus, OTLP
  - Sampling strategies configuration

---

### ORM: GORM (adapter/gorm)

**Status**: ✅ Implemented (v2.0-2.2)

**Why GORM?**
- Most popular Go ORM with comprehensive features
- Plugin architecture for tracing integration
- Support for multiple databases (PostgreSQL, MySQL, SQLite)
- Auto-migration and schema management
- Active development and community support

**Alternatives Considered:**
- **sqlx**: Lightweight but requires more manual work
- **ent**: Type-safe but more opinionated
- **sqlc**: Code generation approach, different paradigm

**Core Design**:
- Core library provides generic `Database` and `Executor` interfaces
- NOT tied to GORM specifically - can implement with sqlx, ent, etc.

**Implementation**:
- **Package**: `github.com/mapoio/hyperion/adapter/gorm`
- **Implements**: `hyperion.Database`, `hyperion.Executor`, and `hyperion.UnitOfWork`
- **Test Coverage**: 82.1%
- **Dependencies**: `gorm.io/gorm v1.25.0`

---

### Cache: Ristretto (adapter/ristretto) / Redis (adapter/redis)

**Status**: 🔜 Planned

**Why Ristretto (In-Memory)?**
- High-performance concurrent cache
- Cost-based eviction for optimal memory usage
- Thread-safe operations
- Production-proven at Dgraph

**Why Redis (Distributed)?**
- Industry-standard distributed cache
- Rich data structures support
- Atomic operations
- Pub/sub capabilities
- Cluster mode for high availability

**Core Design**:
- Core library provides byte-slice based `Cache` interface
- Works with any backend (in-memory, distributed, etc.)

**Planned Implementations**:
- **Package**: `github.com/mapoio/hyperion/adapter/ristretto`
- **Implements**: `hyperion.Cache`
- **Dependencies**: `github.com/dgraph-io/ristretto v0.1.1`

- **Package**: `github.com/mapoio/hyperion/adapter/redis`
- **Implements**: `hyperion.Cache`
- **Dependencies**: `github.com/redis/go-redis/v9 v9.5.0`

---

## Web Framework Technologies (Planned)

### Web Framework: Gin (hyperweb - planned)

**Status**: 🔜 Planned for v2.3

**Why Gin?**
- High performance with minimal overhead
- Mature ecosystem with extensive middleware support
- Excellent documentation and community
- Easy integration with OpenTelemetry
- Familiar API for developers

**Alternatives Considered:**
- **Echo**: Similar performance, smaller ecosystem
- **Fiber**: Fastest but Express.js-like API
- **Chi**: Minimalist but requires more boilerplate

---

### Validation: go-playground/validator (hypervalidator - planned)

**Status**: 🔜 Planned for v2.3

**Why validator?**
- Tag-based validation (struct tags)
- Comprehensive validation rules
- Custom validator support
- Excellent performance
- Most popular Go validation library

**Alternatives Considered:**
- **ozzo-validation**: Programmatic approach, more verbose
- Custom validation: Maintenance overhead

---

### HTTP Client: Resty (hyperhttp - planned)

**Status**: 🔜 Planned for v2.3

**Why Resty?**
- Simple, fluent API
- Built-in retry mechanism
- Request/response middleware support
- JSON/XML handling
- Easy OpenTelemetry integration

**Alternatives Considered:**
- Standard library `http.Client`: Too low-level
- Custom wrapper: Maintenance overhead

---

## Database Support Matrix

### Core Support (v2.0)

The core `hyperion.Database` interface is database-agnostic.

### Planned Adapter Support

| Database | Adapter | Driver | Status |
|----------|---------|--------|--------|
| PostgreSQL | adapter/gorm | `gorm.io/driver/postgres` | 🔜 Planned |
| MySQL | adapter/gorm | `gorm.io/driver/mysql` | 🔜 Planned |
| SQLite | adapter/gorm | `gorm.io/driver/sqlite` | 🔜 Planned |
| SQL Server | adapter/gorm | `gorm.io/driver/sqlserver` | 🔄 Future |

**Note**: Users can implement `hyperion.Database` with ANY ORM or database library (sqlx, ent, etc.)

---

## Version Requirements

### Core Library

| Dependency | Minimum Version | Current |
|------------|----------------|---------|
| Go | 1.24 | 1.24+ |
| fx | 1.22 | 1.24.0 |

### Adapter Dependencies

#### Viper Adapter (✅ Implemented)

| Dependency | Version |
|------------|---------|
| github.com/spf13/viper | 1.21.0 |
| github.com/fsnotify/fsnotify | 1.9.0 |

#### Zap Adapter (🔜 Planned)

| Dependency | Planned Version |
|------------|----------------|
| go.uber.org/zap | 1.27.0+ |

#### OpenTelemetry Adapter (🔜 Planned)

| Dependency | Planned Version |
|------------|----------------|
| go.opentelemetry.io/otel | 1.24.0+ |
| go.opentelemetry.io/otel/trace | 1.24.0+ |

#### GORM Adapter (🔜 Planned)

| Dependency | Planned Version |
|------------|----------------|
| gorm.io/gorm | 1.25.0+ |
| gorm.io/driver/postgres | 1.5.0+ |
| gorm.io/driver/mysql | 1.5.0+ |

---

## Performance Characteristics

### Core Library

- **Overhead**: Near-zero (interfaces only)
- **NoOp Implementations**: < 10ns per call (inline-able)
- **Module Resolution**: ~5-10ms (fx dependency graph construction)

### Adapter Performance (Expected)

#### Logging (Zap)
- **Throughput**: 1M+ logs/sec
- **Allocation**: Near-zero for structured logging
- **Latency**: <100ns per log call (cached logger)

#### Cache (Ristretto)
- **Throughput**: 20M+ ops/sec
- **Hit Ratio**: 40-50% typical production
- **Memory Overhead**: ~2-4 bytes per key

#### Web Framework (Gin)
- **Requests/sec**: 100K+ (single core)
- **Latency**: Sub-millisecond for simple handlers
- **Memory**: ~2KB per request

---

## Security Considerations

### Core Library

- **Dependencies**: ONLY `go.uber.org/fx` - minimal attack surface
- **No Network Code**: Core library makes no network calls
- **No File I/O**: Core library does no file operations

### Adapter Security

#### Encryption (hypercrypto - planned)
- **Algorithm**: AES-256-GCM
- **Key Management**: Environment variables (default), HashiCorp Vault (planned)
- **Best Practices**: Key rotation, secure key storage

#### Authentication (planned for v2.4)
- JWT tokens
- OAuth 2.0 integration
- Session management

---

## Cloud Native Features

### Container Support

- Optimized for Docker containers
- Multi-stage builds for minimal image size
- Health check endpoints (planned)
- Graceful shutdown support (via fx lifecycle)

### Observability

- **Tracing**: OpenTelemetry-compatible Tracer interface (✅ v2.0-2.2, NoOp default)
- **Metrics**: OpenTelemetry-compatible Meter interface (✅ v2.0-2.2, NoOp default)
- **Logging**: Structured JSON logs with trace correlation (✅ adapter/zap, 93.9% coverage)
- **Interceptor Pattern**: Automatic observability via `ctx.UseIntercept()` (✅ v2.0-2.2)
- **Automatic Correlation**: TraceID/SpanID shared across logs, traces, metrics (✅ v2.0-2.2)
- **Full OTel Integration**: Real exporters for Jaeger, Prometheus, OTLP (🔜 Epic 3, Q1 2026)
- **Health Checks**: Kubernetes-ready endpoints (🔜 planned)

### Configuration

- **12-Factor App Compliant**: Environment variable support
- **Hot Reload**: Zero-downtime config updates (adapter/viper)
- **Secret Management**: External secret sources (planned)

---

## Dependency Graph Visualization

### Core Library (hyperion/)

```
hyperion
    └── go.uber.org/fx v1.24.0
        └── go.uber.org/dig v1.19.0
```

**Total Transitive Dependencies**: 3 (fx, dig, multierr)

### Viper Adapter (adapter/viper)

```
adapter/viper
    ├── github.com/mapoio/hyperion v0.0.0
    └── github.com/spf13/viper v1.21.0
        ├── github.com/fsnotify/fsnotify v1.9.0
        ├── github.com/spf13/afero v1.15.0
        └── ... (Viper's dependencies)
```

---

## Technology Selection Criteria

When selecting technologies for Hyperion, we prioritize:

1. **Production Maturity**: Battle-tested in production environments
2. **Community Support**: Active development and large user base
3. **Performance**: Minimal overhead and high throughput
4. **Flexibility**: Can be swapped/replaced (adapter pattern)
5. **Go Idioms**: Follows Go best practices and conventions

---

## Adapter Development Guidelines

### Creating a New Adapter

To add support for a new library:

1. **Create Module**: `mkdir adapter/{name}`
2. **Initialize**: `go mod init github.com/mapoio/hyperion/adapter/{name}`
3. **Implement Interface**: Implement one or more `hyperion.*` interfaces
4. **Export Module**: Provide `fx.Module` in `module.go`
5. **Add Tests**: Unit and integration tests
6. **Document**: README with usage examples

### Example Adapter Structure

```
adapter/mylib/
├── go.mod                  # Independent module
├── implementation.go       # Interface implementation
├── module.go              # fx.Module export
├── implementation_test.go # Tests
└── README.md              # Usage documentation
```

---

## Future Technology Considerations

### Message Queue (v3.0)
- **Kafka**: Distributed streaming
- **RabbitMQ**: Traditional message broker
- **NATS**: Cloud-native messaging

### Task Scheduling (v3.0)
- **Asynq**: Redis-backed task queue
- **Machinery**: Distributed task queue

### Service Mesh (v3.0)
- **Istio**: Full-featured service mesh
- **Linkerd**: Lightweight service mesh

---

## Version History

### v1.0 (Deprecated)
- Bundled implementations (Zap, GORM, Viper)
- Single module structure
- Tight coupling to specific libraries

### v2.0 (Current)
- Core-adapter pattern
- Zero lock-in (core has only fx dependency)
- Independent adapter modules
- Viper adapter implemented

### v2.1 (Planned)
- Zap adapter (Logger)
- GORM adapter (Database)
- Full Context implementation

### v2.2 (Planned)
- OpenTelemetry adapter (Tracer)
- Prometheus metrics
- Cache adapters (Ristretto, Redis)

---

**Last Updated**: October 2025
**Version**: 2.0 (Core-Adapter Architecture)
