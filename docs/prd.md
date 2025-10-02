# Hyperion v2.0 Product Requirements

**Version**: 2.0
**Status**: Active Development
**Last Updated**: October 2025
**Architecture**: Core-Adapter Pattern

---

## Overview

Hyperion v2.0 is a modular Go framework built on the **Core-Adapter architecture** principle, providing **zero lock-in** for developers. Unlike traditional frameworks, Hyperion's core library contains ONLY interfaces and NoOp implementations, while concrete implementations are provided through optional adapters.

**Key Innovation**: Choose your own libraries (Zap, Zerolog, GORM, sqlx, etc.) or write your own adapters—no vendor lock-in.

---

## Product Vision

**Mission**: Build the most flexible Go framework that never forces technology choices on developers.

**Core Values**:
1. **Zero Lock-in**: Core library has ZERO 3rd-party dependencies (except fx)
2. **Maximum Flexibility**: Users choose adapters or write their own
3. **Developer-Friendly**: NoOp defaults allow instant prototyping
4. **Production-Ready**: Real adapters provide enterprise-grade capabilities
5. **Community-Driven**: Open ecosystem for adapter contributions

---

## Architecture Overview

```
┌─────────────────────────────────────┐
│   Your Application                   │
│   (uses interfaces only)             │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   hyperion/ (Core Library)           │
│   • Logger interface                 │
│   • Database interface               │
│   • Config interface                 │
│   • Tracer interface                 │
│   • Cache interface                  │
│   • Context interface                │
│   • ZERO 3rd-party deps (except fx)  │
└──────────────┬──────────────────────┘
               │
    ┌──────────┴──────────────────────┐
    │                                  │
┌───▼────────────┐          ┌─────────▼─────────┐
│ adapter/viper   │          │ adapter/zap        │
│ adapter/sqlx    │          │ adapter/zerolog    │
│ adapter/gorm    │          │ adapter/otel       │
│ adapter/redis   │          │ adapter/gin        │
│ (choose yours)  │          │ (or write yours)   │
└─────────────────┘          └────────────────────┘
```

---

## Release Status

### ✅ v2.0-alpha (October 2025) - **CURRENT**

**Status**: ✅ **RELEASED**

**Delivered**:
- ✅ Core library with all interfaces
- ✅ NoOp implementations (zero-config prototyping)
- ✅ Viper adapter (Config + ConfigWatcher)
- ✅ Module system (CoreModule + CoreWithoutDefaultsModule)
- ✅ Complete documentation (6,600+ lines)
- ✅ Quick start guide
- ✅ Monorepo infrastructure (Go workspace)

**Key Achievement**: Zero lock-in architecture validated

---

### 🔜 v2.1 (December 2025) - **PLANNED**

**Focus**: Production-Ready Logging & Database

**Planned Features**:
- 🔜 **Zap Adapter** - Production logger implementation
  - JSON and Console encoders
  - Dynamic level adjustment
  - File rotation support
- 🔜 **GORM Adapter** - Database with transactions
  - PostgreSQL, MySQL, SQLite support
  - UnitOfWork pattern implementation
  - Automatic transaction propagation
  - Query tracing integration
- 🔜 **Context Implementation** - Production-ready context
  - Logger, Tracer, DB integration
  - Accessor pattern implementation
- 🔜 **Example Application** - Simple CRUD API
  - Demonstrates all v2.1 features
  - Dockerfile and deployment guide

**Timeline**: 6 weeks development

---

### 🔜 v2.2 (February 2026) - **PLANNED**

**Focus**: Observability & Caching

**Planned Features**:
- 🔜 **OpenTelemetry Adapter** - Distributed tracing
  - Jaeger, OTLP, Zipkin exporters
  - Automatic span creation
  - Trace context propagation
- 🔜 **Ristretto Adapter** - In-memory cache
  - High-performance caching (20M+ ops/sec)
  - Cost-based eviction
- 🔜 **Redis Adapter** - Distributed cache
  - Redis standalone and cluster support
  - Connection pooling and retry logic
- 🔜 **Prometheus Metrics** - Metrics integration
  - Automatic metric collection
  - Custom metric support

**Timeline**: 4 weeks development

---

### 🔜 v2.3 (April 2026) - **PLANNED**

**Focus**: Web Framework Integration

**Planned Features**:
- 🔜 **Gin Adapter** - HTTP server
  - Context injection middleware
  - Automatic tracing and logging
  - Graceful shutdown
  - CORS, Recovery, Metrics middleware
- 🔜 **Error Utilities** (Optional) - Convenience package
  - HTTP/gRPC status code mapping
  - Error code constants
- 🔜 **Validation Utilities** (Optional) - Convenience package
  - go-playground/validator wrapper
- 🔜 **Full-Stack Example** - Complete web application
  - Multi-layer architecture
  - Authentication, authorization
  - Production deployment

**Timeline**: 3 weeks development

---

### 🔜 v2.4 (June 2026) - **PLANNED**

**Focus**: Microservices Support

**Planned Features**:
- 🔜 **gRPC Adapter** - gRPC server
  - Unary and Stream interceptors
  - Health check service
  - Automatic tracing
- 🔜 **HTTP Client Adapter** - HTTP client with tracing
  - Retry logic
  - Circuit breaker (optional)
- 🔜 **Service Discovery** - Integration with Consul/etcd
- 🔜 **Microservices Example** - Multi-service architecture

**Timeline**: 3 weeks development

---

### 🔄 v3.0 (2027+) - **FUTURE**

**Focus**: Distributed Systems

**Vision**:
- 🔄 Message queue interfaces (Kafka, RabbitMQ, NATS adapters)
- 🔄 Task scheduling (Asynq, Machinery adapters)
- 🔄 Service mesh integration (Istio, Linkerd)
- 🔄 Event sourcing support
- 🔄 GraphQL adapter

---

## Core Components (v2.0)

### Interface Library (`hyperion/`)

All components exist as **interfaces only** in the core library:

| Interface | Purpose | NoOp Default | Adapters |
|-----------|---------|--------------|----------|
| **Logger** | Structured logging | ✅ Silent | viper (✅), zap (🔜), zerolog (🔄 community) |
| **Config** | Configuration | ✅ Empty | viper (✅) |
| **ConfigWatcher** | Hot reload | ✅ No-op | viper (✅) |
| **Tracer** | Distributed tracing | ✅ No-op | otel (🔜) |
| **Database** | Database access | ✅ No-op | gorm (🔜), sqlx (🔄 community) |
| **Executor** | Query execution | ✅ No-op | gorm (🔜), sqlx (🔄 community) |
| **UnitOfWork** | Transactions | ✅ No-op | gorm (🔜), sqlx (🔄 community) |
| **Cache** | Key-value cache | ✅ No-op | ristretto (🔜), redis (🔜) |
| **Context** | Request context | ✅ Basic | Full impl (🔜) |

**Design Principles**:
- Every interface has a NoOp implementation
- NoOp implementations co-located with interfaces
- No 3rd-party dependencies in core (except fx)

---

## Adapter Development

### Official Adapters (Maintained by Core Team)

**Priority 1** (v2.1):
- `adapter/zap` - Zap logger adapter
- `adapter/gorm` - GORM database adapter

**Priority 2** (v2.2):
- `adapter/otel` - OpenTelemetry tracer adapter
- `adapter/ristretto` - In-memory cache adapter
- `adapter/redis` - Redis cache adapter

**Priority 3** (v2.3):
- `adapter/gin` - Gin web framework adapter

**Priority 4** (v2.4):
- `adapter/grpc` - gRPC server adapter
- `adapter/resty` - HTTP client adapter

### Community Adapters (Wanted)

We welcome community contributions for:

**High Priority**:
- `adapter/sqlx` - sqlx database adapter (alternative to GORM)
- `adapter/zerolog` - Zerolog logger adapter (alternative to Zap)
- `adapter/chi` - Chi router adapter (alternative to Gin)
- `adapter/fiber` - Fiber web framework adapter

**Medium Priority**:
- `adapter/nats` - NATS message queue adapter
- `adapter/kafka` - Kafka adapter
- `adapter/rabbitmq` - RabbitMQ adapter
- `adapter/temporal` - Temporal workflow adapter

**Low Priority**:
- `adapter/memcached` - Memcached cache adapter
- `adapter/dynamodb` - DynamoDB adapter
- `adapter/mongodb` - MongoDB adapter

---

## Development Epics (v2.0)

Unlike v1.0 which had monolithic epics, v2.0 development is organized around **adapters**:

### ✅ Epic 1: Core Foundation (Completed - October 2025)

**Goal**: Establish core interfaces with zero lock-in

**Delivered**:
- ✅ All core interfaces defined
- ✅ NoOp implementations
- ✅ Module system (CoreModule)
- ✅ Viper adapter (Config)
- ✅ Complete documentation

**Duration**: 4 weeks
**Status**: ✅ **COMPLETED**

---

### 🔜 Epic 2: Essential Adapters (v2.1 - December 2025)

**Goal**: Production-ready logging and database access

**Scope**:
- 🔜 Zap adapter implementation (5 days)
- 🔜 GORM adapter implementation (7 days)
- 🔜 Context implementation (3 days)
- 🔜 Example CRUD application (3 days)
- 🔜 Integration tests (2 days)

**Duration**: 4 weeks
**Status**: 🔜 **PLANNED**

---

### 🔜 Epic 3: Observability Stack (v2.2 - February 2026)

**Goal**: Complete observability with tracing and caching

**Scope**:
- 🔜 OpenTelemetry adapter (5 days)
- 🔜 Ristretto adapter (3 days)
- 🔜 Redis adapter (4 days)
- 🔜 Prometheus metrics (3 days)
- 🔜 Observability example (3 days)

**Duration**: 3 weeks
**Status**: 🔜 **PLANNED**

---

### 🔜 Epic 4: Web Framework (v2.3 - April 2026)

**Goal**: Production-ready web application framework

**Scope**:
- 🔜 Gin adapter (7 days)
- 🔜 Middleware suite (3 days)
- 🔜 Optional utilities (2 days)
- 🔜 Full-stack example (3 days)

**Duration**: 3 weeks
**Status**: 🔜 **PLANNED**

---

### 🔜 Epic 5: Microservices (v2.4 - June 2026)

**Goal**: Microservices architecture support

**Scope**:
- 🔜 gRPC adapter (5 days)
- 🔜 HTTP client adapter (3 days)
- 🔜 Service discovery (4 days)
- 🔜 Microservices example (3 days)

**Duration**: 3 weeks
**Status**: 🔜 **PLANNED**

---

## Quality Standards

All adapters must meet these standards before release:

### Code Quality
- [ ] Test coverage >= 80%
- [ ] golangci-lint passes with no errors
- [ ] All public APIs documented (godoc)
- [ ] README with usage examples
- [ ] Integration tests (where applicable)

### Performance
- [ ] Benchmarks for critical paths
- [ ] Performance overhead < 5% vs native library
- [ ] Memory profiling completed
- [ ] No memory leaks

### Documentation
- [ ] API documentation complete
- [ ] Code examples working
- [ ] Migration guide (if replacing NoOp)
- [ ] Troubleshooting guide

### Release
- [ ] CHANGELOG updated
- [ ] Semantic versioning followed
- [ ] GitHub release created
- [ ] Announced to community

---

## Success Metrics

### v2.0-alpha (Baseline)
- ✅ Core library: 0 dependencies (except fx)
- ✅ Documentation: 6,600+ lines
- ✅ Adapter example: Viper (fully functional)

### v2.1 Target (December 2025)
- Production users: 10+
- GitHub stars: 200+
- Community adapters: 1+
- Tutorial completions: 50+

### v2.4 Target (June 2026)
- Production users: 50+
- GitHub stars: 1,000+
- Community adapters: 10+
- Tutorial completions: 500+

### v3.0 Vision (2027+)
- Production users: 500+
- GitHub stars: 5,000+
- Community adapters: 50+
- Ecosystem packages: 100+

---

## Competitive Differentiation

### vs. Traditional Frameworks (Echo, Gin, Fiber)

**Traditional**:
- Forced to use framework's logger, router, etc.
- Hard to swap components
- Tight coupling

**Hyperion v2.0**:
- ✅ Use ANY logger (Zap, Zerolog, logrus, etc.)
- ✅ Use ANY database library (GORM, sqlx, ent, etc.)
- ✅ Use ANY router (Gin, Chi, Fiber, etc.)
- ✅ Zero coupling - pure interfaces

### vs. Go-Kit

**Go-Kit**:
- Microservices-focused
- Steep learning curve
- Limited documentation

**Hyperion v2.0**:
- ✅ Simple monolith OR microservices
- ✅ Gentle learning curve
- ✅ Comprehensive documentation (6,600+ lines)

### vs. Kratos (Bilibili)

**Kratos**:
- Opinionated (gRPC-first)
- Chinese-primary documentation
- Complex setup

**Hyperion v2.0**:
- ✅ HTTP or gRPC (your choice)
- ✅ English-first documentation
- ✅ Simple setup (15-minute quick start)

---

## User Personas

### 1. Startup Developer (Target: v2.1+)
**Needs**: Fast prototyping, flexibility to change later
**Solution**: Start with NoOp defaults, swap adapters as needs grow

### 2. Enterprise Developer (Target: v2.2+)
**Needs**: Production stability, observability, compliance
**Solution**: Use official adapters (Zap, GORM, OTel) with proven reliability

### 3. Open Source Maintainer (Target: v2.0+)
**Needs**: Library that doesn't force dependencies on users
**Solution**: Depend on Hyperion core (zero deps), let users choose adapters

### 4. Library Author (Target: All versions)
**Needs**: Create adapter for niche library
**Solution**: Follow adapter development guide, join ecosystem

---

## User Journey

### Phase 1: Discovery (v2.0-alpha)
1. Developer finds Hyperion on GitHub
2. Reads about "zero lock-in" promise
3. Tries 15-minute quick start
4. Builds sample app with NoOp defaults
5. **Impression**: "This is easy!"

### Phase 2: Adoption (v2.1+)
1. Developer adds Zap adapter for logging
2. App now has real logs
3. Adds GORM adapter for database
4. App now persists data
5. **Impression**: "This is powerful!"

### Phase 3: Production (v2.2+)
1. Developer adds OTel adapter for tracing
2. Adds Redis adapter for caching
3. App is fully observable
4. Deploys to production
5. **Impression**: "This is production-ready!"

### Phase 4: Growth (v2.3+)
1. Team grows, needs microservices
2. Adds gRPC adapter
3. Splits monolith into services
4. All using Hyperion
5. **Impression**: "This scales!"

---

## Risk Management

### Technical Risks

**Risk**: Adapter ecosystem doesn't grow
**Mitigation**:
- Provide 5+ official adapters (Zap, GORM, OTel, etc.)
- Make adapter development extremely easy
- Promote community adapters

**Risk**: Core interfaces too rigid
**Mitigation**:
- Design interfaces for 80% use cases
- Allow custom adapters for edge cases
- Accept breaking changes before v3.0

**Risk**: Performance overhead
**Mitigation**:
- Benchmark all adapters
- Target < 5% overhead vs native
- Optimize hot paths

### Community Risks

**Risk**: Developers don't understand Core-Adapter pattern
**Mitigation**:
- Crystal-clear documentation (✅ 6,600+ lines)
- Visual diagrams (✅ architecture diagrams)
- Step-by-step tutorials (✅ 15-minute quick start)

**Risk**: Fragmented ecosystem
**Mitigation**:
- Clear adapter naming conventions
- Curated list of quality adapters
- Adapter quality badges

---

## Go-to-Market Strategy

### Phase 1: Developer Awareness (v2.0-v2.1)
- Blog posts on "Zero Lock-in Architecture"
- Reddit posts on r/golang
- Hacker News launch
- Twitter/X promotion
- YouTube tutorial videos

### Phase 2: Community Building (v2.2-v2.3)
- Discord server for community
- Adapter development contests
- Conference talks (GopherCon)
- Podcast interviews

### Phase 3: Enterprise Adoption (v2.4+)
- Case studies from production users
- Enterprise support offerings
- Training programs
- Consulting services

---

## Related Documentation

### Architecture & Design
- [Architecture Overview](architecture.md) - Complete v2.0 architecture (2,531 lines)
- [Architecture Decisions](architecture-decisions.md) - ADRs explaining design choices (637 lines)
- [Source Tree Guide](architecture/source-tree.md) - Monorepo structure (601 lines)
- [Tech Stack](architecture/tech-stack.md) - Technology choices (479 lines)
- [Coding Standards](architecture/coding-standards.md) - Development guidelines (713 lines)

### Getting Started
- [Quick Start Guide](quick-start.md) - 15-minute tutorial (809 lines)
- [Implementation Plan](implementation-plan.md) - Detailed roadmap (843 lines)

### Legacy (v1.0)
- [Epic 1: Core Foundation](prd/epic-1-core-foundation.md) - ⚠️ v1.0 (deprecated)
- [Epic 2: Data Access Layer](prd/epic-2-data-access-layer.md) - ⚠️ v1.0 (deprecated)
- [Epic 3-7: Other Epics](prd/) - ⚠️ v1.0 (deprecated)

**Note**: v1.0 Epic documents are kept for historical reference but DO NOT reflect v2.0 architecture.

---

## Frequently Asked Questions

### Q: Why Core-Adapter pattern?
**A**: To achieve zero lock-in. Your application code depends ONLY on interfaces, never on concrete implementations. Swap any adapter without touching application code.

### Q: What if there's no adapter for my library?
**A**: Write your own! It's just implementing an interface and exporting an fx.Module. See [Adapter Development Guide](architecture/source-tree.md#adapter-development-guidelines).

### Q: Can I use multiple logger adapters?
**A**: No. One adapter per interface. But you can write a "composite adapter" that uses multiple libraries internally.

### Q: What's the performance overhead?
**A**: Minimal. Interfaces are compile-time, NoOp implementations are inline-able, adapters are thin wrappers (< 5% overhead).

### Q: Is this production-ready?
**A**: v2.0-alpha is for early adopters. v2.1+ will be production-ready with official adapters.

---

**Welcome to Hyperion v2.0 - Zero Lock-in, Maximum Flexibility! 🚀**

**Last Updated**: October 2025
**Version**: 2.0 (Core-Adapter Architecture)
