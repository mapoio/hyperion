# Epic 6: Advanced Features

**Priority**: ⭐⭐⭐ (Medium)
**Estimated Duration**: 1 week
**Status**: Not Started
**Dependencies**: Epic 1 (Core Foundation)

---

## Overview

Implement advanced features including remote configuration support and comprehensive example applications.

---

## Goals

- Support remote configuration sources (Consul, Etcd)
- Provide production-ready example applications
- Demonstrate all framework capabilities
- Enable zero-downtime configuration updates

---

## User Stories

### Story 6.1: Remote Configuration Support

**As a** framework user
**I want** remote configuration support
**So that** I can manage configuration centrally across multiple services

**Acceptance Criteria**:
- [ ] Can load configuration from Consul
- [ ] Configuration changes trigger callbacks in real-time
- [ ] Support for long polling / watch mechanism
- [ ] Fallback to local config if remote unavailable
- [ ] Secure connection to remote config source

**Tasks**:
- [ ] Define `RemoteProvider` interface
- [ ] Implement `ConsulProvider`
  - [ ] Implement key-value read
  - [ ] Implement Watch with long polling
  - [ ] Implement connection pooling
  - [ ] Implement retry logic
- [ ] Implement fallback mechanism
- [ ] Implement secure connection (TLS)
- [ ] Write unit tests (>80% coverage)
- [ ] Write integration tests with real Consul
- [ ] Write godoc documentation

**Technical Details**:
```go
// Remote config provider
type RemoteProvider interface {
    Provider
    Watcher
    Connect() error
    Disconnect() error
}

// Consul implementation
type ConsulProvider struct {
    client *consul.Client
    prefix string
}

func (p *ConsulProvider) Watch(callback func(ChangeEvent)) (func(), error) {
    // Long polling
    go func() {
        index := uint64(0
        for {
            kv, meta, err := p.client.KV().Get(p.prefix, &consul.QueryOptions{
                WaitIndex: index,
                WaitTime:  5 * time.Minute,
            })

            if err != nil {
                // Retry with backoff
                continue
            }

            if meta.LastIndex != index {
                index = meta.LastIndex
                callback(ChangeEvent{Key: p.prefix, Value: kv.Value})
            }
        }
    }()

    return func() { /* stop watching */ }, nil
}
```

**Configuration**:
```yaml
config:
  sources:
    # Primary: Remote (Consul)
    - type: consul
      address: "localhost:8500"
      prefix: "hyperion/config"
      watch: true
      tls:
        enabled: true
        cert_file: "/path/to/cert.pem"

    # Fallback: Local file
    - type: file
      path: "configs/config.yaml"
```

**Estimated**: 3 days

---

### Story 6.2: Example Applications

**As a** framework user
**I want** comprehensive example applications
**So that** I can learn framework best practices quickly

**Acceptance Criteria**:
- [ ] Simple API example demonstrates basic CRUD
- [ ] Full-stack example demonstrates all features
- [ ] Examples include README with setup instructions
- [ ] Examples include Dockerfile for containerization
- [ ] Examples demonstrate testing strategies

**Tasks**:

**Simple API Example**:
- [ ] Create project structure
- [ ] Implement user CRUD (Create, Read, Update, Delete)
- [ ] Implement repository layer
- [ ] Implement service layer
- [ ] Implement handler layer
- [ ] Add configuration files
- [ ] Add database migrations
- [ ] Add README with instructions
- [ ] Add Dockerfile

**Full-Stack Example**:
- [ ] Create project structure (Web + gRPC)
- [ ] Implement domain models
- [ ] Implement user service (with transactions)
- [ ] Implement order service (demonstrating cache)
- [ ] Implement HTTP handlers
- [ ] Implement gRPC services
- [ ] Add validation examples
- [ ] Add error handling examples
- [ ] Add tracing demonstration
- [ ] Add configuration examples (local + remote)
- [ ] Add docker-compose.yml (app + PostgreSQL + Redis + Jaeger + Consul)
- [ ] Add comprehensive README
- [ ] Add unit tests
- [ ] Add integration tests

**Project Structure**:
```
examples/
├── simple-api/              # Basic CRUD API
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── handler/
│   │   ├── service/
│   │   └── repository/
│   ├── configs/config.yaml
│   ├── Dockerfile
│   └── README.md
│
└── fullstack/               # Complete example
    ├── cmd/server/main.go
    ├── internal/
    │   ├── domain/
    │   ├── handler/         # HTTP
    │   ├── grpc/            # gRPC services
    │   ├── service/
    │   └── repository/
    ├── api/
    │   ├── proto/           # Protobuf definitions
    │   └── openapi/         # OpenAPI specs
    ├── configs/
    │   ├── config.yaml
    │   └── config.consul.yaml
    ├── scripts/
    │   ├── migrate.go
    │   └── seed.go
    ├── docker-compose.yml
    ├── Dockerfile
    └── README.md
```

**Example Features**:

**Simple API**:
- User CRUD operations
- Database persistence (PostgreSQL)
- Automatic tracing
- Error handling
- Request validation

**Full-Stack**:
- User management (HTTP + gRPC)
- Order processing with transactions
- Redis caching
- Distributed tracing (Jaeger)
- Remote configuration (Consul)
- Health checks
- Metrics (Prometheus - future)
- Comprehensive testing

**Estimated**: 2 days

---

## Milestone

**Deliverable**: Production-ready advanced features and comprehensive examples

**Demo Scenario**:

**Remote Configuration**:
```go
func main() {
    // Load config from Consul with fallback
    fx.New(
        hyperion.Core(
            hyperconfig.WithConsul("localhost:8500", "hyperion/config"),
            hyperconfig.WithFileFallback("configs/config.yaml"),
        ),
        fx.Invoke(StartApplication),
    ).Run()
}

// Dynamic configuration update
func StartApplication(cfg hyperconfig.Provider, logger hyperlog.Logger) {
    cfg.Watch(func(event hyperconfig.ChangeEvent) {
        var logCfg hyperlog.Config
        cfg.Unmarshal("log", &logCfg)

        logger.SetLevel(parseLevel(logCfg.Level))
        logger.Info("log level updated", "level", logCfg.Level)
    })
}
```

**Example Application** (`examples/fullstack`):
```bash
# Start infrastructure
docker-compose up -d

# Run migrations
go run scripts/migrate.go

# Start application
go run cmd/server/main.go

# Test HTTP API
curl http://localhost:8080/api/v1/users

# Test gRPC
grpcurl -plaintext localhost:9090 user.v1.UserService/GetUser

# View traces in Jaeger
open http://localhost:16686

# Update config in Consul (triggers hot reload)
consul kv put hyperion/config/log.level debug
```

---

## Technical Notes

### Architecture Decisions

- **Consul for Remote Config**: Industry standard, feature-rich, proven at scale
- **Watch Mechanism**: Long polling for real-time updates
- **Fallback Strategy**: Local file as backup for resilience

### Dependencies

- `github.com/hashicorp/consul/api` - Consul client
- Infrastructure:
  - PostgreSQL - Database
  - Redis - Cache
  - Jaeger - Tracing
  - Consul - Configuration

### Remote Config Providers

| Provider | Status | Features |
|----------|--------|----------|
| Consul | ✅ Implemented | KV store, Watch, ACL, TLS |
| Etcd | 🔄 Future | KV store, Watch, TLS |
| AWS Secrets Manager | 🔄 Future | Secrets management |
| Vault | 🔄 Future | Secrets + dynamic credentials |

### Example Applications Matrix

| Example | HTTP | gRPC | DB | Cache | Tracing | Remote Config | Tests |
|---------|------|------|-------|-------|---------|---------------|-------|
| simple-api | ✅ | ❌ | ✅ | ❌ | ✅ | ❌ | ✅ |
| fullstack | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

### Testing Strategy

- **Unit Tests**:
  - Consul provider logic
  - Watch mechanism
  - Fallback behavior
- **Integration Tests**:
  - Real Consul instance
  - Configuration hot reload
  - Example applications end-to-end

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Consul unavailability | Medium | Fallback to local config |
| Network partition | Medium | Retry with exponential backoff |
| Configuration corruption | High | Validation before applying |
| Example complexity | Low | Comprehensive documentation |

---

## Related Documentation

- [Architecture - hyperconfig](../architecture.md#52-hyperconfig---configuration-management)
- [Quick Start Guide](../quick-start.md)
- [Coding Standards](../architecture/coding-standards.md)

---

**Last Updated**: 2025-01-XX
