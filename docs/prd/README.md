# Hyperion Epic Documents - v1.0 (Deprecated)

**⚠️ DEPRECATED WARNING ⚠️**

The documents in this directory represent the **v1.0 (deprecated)** product requirements and implementation plan. They are kept for **historical reference only**.

---

## Why Deprecated?

Hyperion v2.0 adopted a **fundamentally different architecture**:

### v1.0 Architecture (Deprecated)
```
Single Package Structure (pkg/hyper*)
  ├── Bundled implementations (forced Zap, GORM, Viper)
  ├── Tight coupling
  └── Vendor lock-in
```

### v2.0 Architecture (Current)
```
Core-Adapter Pattern
  ├── Core Library: ZERO dependencies (except fx)
  ├── Adapters: Optional concrete implementations
  └── Zero lock-in: Choose ANY library
```

The v1.0 epic structure (monolithic component development) doesn't apply to v2.0's adapter-based development model.

---

## What Changed?

| Aspect | v1.0 | v2.0 |
|--------|------|------|
| **Structure** | Single package `pkg/hyper*` | Monorepo `hyperion/` + `adapter/*` |
| **Dependencies** | Bundled (Zap, GORM, Viper) | Core: zero deps, Adapters: specific deps |
| **Epics** | 7 monolithic epics | 5 adapter-focused epics |
| **Timeline** | 9.5 weeks | 16 weeks (phased) |
| **Philosophy** | Convention over configuration | Zero lock-in over convenience |

---

## v1.0 Epic Documents (Deprecated)

These documents describe implementation plans that **DO NOT** apply to v2.0:

- ⚠️ [Epic 1: Core Foundation](epic-1-core-foundation.md) - v1.0 bundled components
- ⚠️ [Epic 2: Data Access Layer](epic-2-data-access-layer.md) - v1.0 GORM integration
- ⚠️ [Epic 3: Error Handling & Validation](epic-3-error-validation.md) - v1.0 hypererror package
- ⚠️ [Epic 4: Web Service Layer](epic-4-web-service-layer.md) - v1.0 hyperweb/hypergrpc
- ⚠️ [Epic 5: Clients & Utilities](epic-5-clients-utilities.md) - v1.0 utility packages
- ⚠️ [Epic 6: Advanced Features](epic-6-advanced-features.md) - v1.0 advanced features
- ⚠️ [Epic 7: Documentation & Release](epic-7-documentation-release.md) - v1.0 release plan

**Do not use these as implementation guides for v2.0 development.**

---

## Current v2.0 Documentation

For v2.0 development, refer to these documents:

### Product & Planning
- **[PRD (v2.0)](../prd.md)** - Current product requirements
- **[Implementation Plan (v2.0)](../implementation-plan.md)** - Detailed v2.0 roadmap

### Architecture
- **[Architecture Overview](../architecture.md)** - Complete v2.0 architecture
- **[Architecture Decisions](../architecture-decisions.md)** - ADRs for v2.0 choices
- **[Source Tree Guide](../architecture/source-tree.md)** - Monorepo structure
- **[Tech Stack](../architecture/tech-stack.md)** - Technology decisions

### Getting Started
- **[Quick Start Guide](../quick-start.md)** - 15-minute v2.0 tutorial
- **[Coding Standards](../architecture/coding-standards.md)** - v2.0 development guidelines

---

## v2.0 Development Epics (Current)

v2.0 development is organized around **adapters**, not monolithic components:

### ✅ Epic 1: Core Foundation (Completed)
- Core interfaces with zero lock-in
- NoOp implementations
- Viper adapter
- Complete documentation

### 🔜 Epic 2: Essential Adapters (v2.1)
- Zap logger adapter
- GORM database adapter
- Context implementation
- Example CRUD app

### 🔜 Epic 3: Observability Stack (v2.2)
- OpenTelemetry tracer adapter
- Ristretto cache adapter
- Redis cache adapter
- Prometheus metrics

### 🔜 Epic 4: Web Framework (v2.3)
- Gin web adapter
- Middleware suite
- Full-stack example

### 🔜 Epic 5: Microservices (v2.4)
- gRPC adapter
- HTTP client adapter
- Service discovery
- Microservices example

See **[Implementation Plan](../implementation-plan.md)** for detailed timelines and tasks.

---

## Migration Notes

If you were planning to implement based on v1.0 epics:

1. **Stop**: Do not follow v1.0 epic documents
2. **Read**: Study the v2.0 architecture documents
3. **Understand**: Core-Adapter pattern is fundamentally different
4. **Start**: Follow v2.0 implementation plan for adapter development

### Key Differences to Understand

**v1.0 Approach** (Don't do this):
```go
// Forced to use bundled implementations
import "github.com/mapoio/hyperion/pkg/hyperlog"

logger := hyperlog.NewZapLogger(config)  // Forced Zap
```

**v2.0 Approach** (Do this):
```go
// Use interfaces, choose adapters
import "github.com/mapoio/hyperion"
import "github.com/mapoio/hyperion/adapter/zap"  // OR any other adapter

fx.New(
    hyperion.CoreModule,  // Interfaces + NoOp
    zap.Module,           // OR zerolog.Module, OR your own
)
```

---

## Questions?

If you have questions about v1.0 vs v2.0:

1. Read the **[Architecture Decisions](../architecture-decisions.md)** document
2. See **ADR-004: Core-Adapter Pattern for Zero Lock-in**
3. Review the **[PRD (v2.0)](../prd.md)** for product rationale
4. Ask in GitHub Discussions

---

**Remember**: These v1.0 documents are historical artifacts. All new development follows v2.0 architecture.

**Last Updated**: October 2025
