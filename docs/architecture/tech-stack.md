# Hyperion Technology Stack

This document provides detailed information about the technology choices made for Hyperion framework.

---

## Core Technologies

### Dependency Injection: go.uber.org/fx

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

---

### Web Framework: Gin

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

### ORM: GORM

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

---

### Configuration: Viper

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

---

### Logging: Zap

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

---

### Tracing: OpenTelemetry

**Why OpenTelemetry?**
- Industry-standard observability framework
- Vendor-neutral (works with Jaeger, Zipkin, etc.)
- Comprehensive SDK with auto-instrumentation
- Active CNCF project with strong community
- Built-in support for traces, metrics, and logs

**Alternatives Considered:**
- **OpenTracing**: Deprecated in favor of OpenTelemetry
- **Jaeger Client**: Vendor-specific
- Custom tracing: Reinventing the wheel

---

### Cache: Ristretto (in-memory) / Redis (distributed)

**Why Ristretto?**
- High-performance concurrent cache
- Cost-based eviction for optimal memory usage
- Thread-safe operations
- Production-proven at Dgraph

**Why Redis?**
- Industry-standard distributed cache
- Rich data structures support
- Atomic operations
- Pub/sub capabilities
- Cluster mode for high availability

---

### Validation: go-playground/validator

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

### HTTP Client: Resty

**Why Resty?**
- Simple, fluent API
- Built-in retry mechanism
- Request/response middleware support
- JSON/XML handling
- Easy OpenTelemetry integration

**Alternatives Considered:**
- Standard library `http.Client`: Too low-level
- **go-resty/resty**: Winner (this is our choice)

---

## Database Support Matrix

| Database | Driver | Status |
|----------|--------|--------|
| PostgreSQL | `gorm.io/driver/postgres` | ✅ Supported |
| MySQL | `gorm.io/driver/mysql` | ✅ Supported |
| SQLite | `gorm.io/driver/sqlite` | ✅ Supported |
| SQL Server | `gorm.io/driver/sqlserver` | 🔄 Planned |

---

## Version Requirements

| Dependency | Minimum Version | Recommended |
|------------|----------------|-------------|
| Go | 1.21 | 1.22+ |
| fx | 1.20 | Latest |
| Gin | 1.9 | Latest |
| GORM | 1.25 | Latest |
| Zap | 1.26 | Latest |
| OpenTelemetry | 1.21 | Latest |

---

## Performance Characteristics

### Logging (Zap)
- **Throughput**: 1M+ logs/sec
- **Allocation**: Near-zero for structured logging
- **Latency**: <100ns per log call (cached logger)

### Cache (Ristretto)
- **Throughput**: 20M+ ops/sec
- **Hit Ratio**: 40-50% typical production
- **Memory Overhead**: ~2-4 bytes per key

### Web Framework (Gin)
- **Requests/sec**: 100K+ (single core)
- **Latency**: Sub-millisecond for simple handlers
- **Memory**: ~2KB per request

---

## Security Considerations

### Encryption (hypercrypto)
- **Algorithm**: AES-256-GCM
- **Key Management**: Environment variables (default)
- **Future**: Integration with HashiCorp Vault

### Authentication
- JWT tokens (planned for v1.2)
- OAuth 2.0 integration (planned for v1.2)

---

## Cloud Native Features

### Container Support
- Optimized for Docker containers
- Multi-stage builds for minimal image size
- Health check endpoints for Kubernetes

### Observability
- Prometheus-compatible metrics (v1.1)
- Jaeger/Zipkin trace export
- Structured JSON logs for log aggregation

### Configuration
- 12-factor app compliant
- Environment variable override
- Config hot-reload for zero-downtime updates

---

**Last Updated**: January 2025
