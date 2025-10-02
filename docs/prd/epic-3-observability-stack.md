# Epic 3: Observability Stack (v2.2)

**Version**: 2.2
**Status**: 🔜 **PLANNED** (February 2026)
**Duration**: 4 weeks
**Priority**: ⭐⭐⭐⭐

---

## Overview

Implement production-grade **observability adapters** for Tracing (OpenTelemetry), Caching (Ristretto/Redis), and Metrics (Prometheus), enabling comprehensive application monitoring and performance optimization while maintaining zero lock-in.

---

## Goals

### Primary Goals
1. OpenTelemetry tracing with distributed context propagation
2. High-performance in-memory caching (Ristretto)
3. Distributed caching with Redis
4. Prometheus metrics integration
5. Complete observability example application

### Success Criteria
- [ ] OpenTelemetry adapter passes all Tracer interface tests
- [ ] Ristretto adapter achieves >90% cache hit rate in benchmarks
- [ ] Redis adapter supports TTL, eviction, and clustering
- [ ] Prometheus metrics expose framework and business metrics
- [ ] Example application demonstrates full observability stack
- [ ] Performance overhead < 3% vs native libraries

---

## Deliverables

### 1. OpenTelemetry Tracer Adapter 🔜

**Package**: `adapter/otel/`

**Scope**:
- Implement `hyperion.Tracer` interface using OpenTelemetry
- Implement `hyperion.Span` interface (OTel-compatible)
- Support multiple exporters (Jaeger, Zipkin, OTLP)
- Distributed context propagation
- Automatic span attributes from Context
- Configuration integration via `hyperion.Config`

**Interface Implementation**:
```go
type otelTracer struct {
    tracer trace.Tracer
}

func (t *otelTracer) Start(ctx context.Context, spanName string, opts ...any) (context.Context, hyperion.Span) {
    spanCtx, otelSpan := t.tracer.Start(ctx, spanName, convertOpts(opts)...)
    return spanCtx, &otelSpan{Span: otelSpan}
}

type otelSpan struct {
    trace.Span
}

func (s *otelSpan) End(options ...any) {
    s.Span.End(convertEndOpts(options)...)
}

func (s *otelSpan) AddEvent(name string, options ...any) {
    s.Span.AddEvent(name, convertEventOpts(options)...)
}

func (s *otelSpan) RecordError(err error, options ...any) {
    s.Span.RecordError(err, convertErrorOpts(options)...)
}

func (s *otelSpan) SetAttributes(attributes ...any) {
    s.Span.SetAttributes(convertAttributes(attributes)...)
}
```

**Module Export**:
```go
var Module = fx.Module("hyperion.adapter.otel",
    fx.Provide(
        fx.Annotate(
            NewOtelTracer,
            fx.As(new(hyperion.Tracer)),
        ),
    ),
    fx.Invoke(RegisterShutdownHook), // Ensure trace flush on shutdown
)
```

**Configuration Example**:
```yaml
tracing:
  enabled: true
  service_name: my-service
  exporter: jaeger          # jaeger, zipkin, otlp
  endpoint: localhost:14268
  sample_rate: 1.0          # 0.0 - 1.0
  attributes:               # Global span attributes
    environment: production
    version: v1.0.0
```

**Context Integration**:
```go
func (s *UserService) GetUser(ctx hyperion.Context, id string) (*User, error) {
    // Automatic distributed tracing
    ctx, span := ctx.Tracer().Start(ctx, "UserService.GetUser")
    defer span.End()

    span.SetAttributes("user.id", id)

    user, err := s.userRepo.FindByID(ctx, id)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }

    span.SetAttributes("user.found", true)
    return user, nil
}
```

**Tasks**:
- [ ] Implement otelTracer struct (2 days)
- [ ] Implement otelSpan wrapper (1 day)
- [ ] Add exporter support (Jaeger, Zipkin, OTLP) (2 days)
- [ ] Integrate with Config (1 day)
- [ ] Add context propagation helpers (1 day)
- [ ] Write unit tests (>80% coverage) (1 day)
- [ ] Write integration tests (real exporters) (1 day)
- [ ] Documentation and examples (1 day)

**Timeline**: 7 working days

---

### 2. Ristretto Cache Adapter 🔜

**Package**: `adapter/ristretto/`

**Scope**:
- Implement `hyperion.Cache` interface using Ristretto
- High-performance in-memory caching
- Automatic cost-based eviction
- TTL support
- Metrics integration (hit rate, evictions)
- Type-safe generic wrappers

**Interface Implementation**:
```go
type ristrettoCache struct {
    cache *ristretto.Cache
}

func (c *ristrettoCache) Get(ctx context.Context, key string) ([]byte, error) {
    value, found := c.cache.Get(key)
    if !found {
        return nil, hyperion.ErrCacheKeyNotFound
    }
    return value.([]byte), nil
}

func (c *ristrettoCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    cost := int64(len(value))
    if !c.cache.SetWithTTL(key, value, cost, ttl) {
        return hyperion.ErrCacheSetFailed
    }
    return nil
}

func (c *ristrettoCache) Delete(ctx context.Context, key string) error {
    c.cache.Del(key)
    return nil
}

func (c *ristrettoCache) Exists(ctx context.Context, key string) (bool, error) {
    _, found := c.cache.Get(key)
    return found, nil
}
```

**Module Export**:
```go
var Module = fx.Module("hyperion.adapter.ristretto",
    fx.Provide(
        fx.Annotate(
            NewRistrettoCache,
            fx.As(new(hyperion.Cache)),
        ),
    ),
)
```

**Configuration Example**:
```yaml
cache:
  type: ristretto
  max_cost: 104857600      # 100 MB
  num_counters: 1000000    # 10x max items
  buffer_items: 64
  metrics: true            # Expose cache metrics
```

**Generic Helper (Optional)**:
```go
// Helper for type-safe caching
func Get[T any](ctx hyperion.Context, key string) (*T, error) {
    data, err := ctx.Cache().Get(ctx, key)
    if err != nil {
        return nil, err
    }

    var result T
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, err
    }
    return &result, nil
}

func Set[T any](ctx hyperion.Context, key string, value T, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return ctx.Cache().Set(ctx, key, data, ttl)
}
```

**Tasks**:
- [ ] Implement ristrettoCache struct (1 day)
- [ ] Add TTL and cost-based eviction (1 day)
- [ ] Integrate with Config (1 day)
- [ ] Add generic helpers (optional) (1 day)
- [ ] Expose metrics (hit rate, evictions) (1 day)
- [ ] Write unit tests (>80% coverage) (1 day)
- [ ] Write benchmark tests (1 day)
- [ ] Documentation and examples (1 day)

**Timeline**: 5 working days

---

### 3. Redis Cache Adapter 🔜

**Package**: `adapter/redis/`

**Scope**:
- Implement `hyperion.Cache` interface using Redis
- Distributed caching support
- TTL and eviction policies
- Connection pooling (go-redis)
- Cluster and Sentinel support
- Pipeline and transaction support (advanced)

**Interface Implementation**:
```go
type redisCache struct {
    client redis.UniversalClient
}

func (c *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
    data, err := c.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, hyperion.ErrCacheKeyNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("redis get failed: %w", err)
    }
    return data, nil
}

func (c *redisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
        return fmt.Errorf("redis set failed: %w", err)
    }
    return nil
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
    if err := c.client.Del(ctx, key).Err(); err != nil {
        return fmt.Errorf("redis delete failed: %w", err)
    }
    return nil
}

func (c *redisCache) Exists(ctx context.Context, key string) (bool, error) {
    count, err := c.client.Exists(ctx, key).Result()
    if err != nil {
        return false, fmt.Errorf("redis exists failed: %w", err)
    }
    return count > 0, nil
}
```

**Module Export**:
```go
var Module = fx.Module("hyperion.adapter.redis",
    fx.Provide(
        fx.Annotate(
            NewRedisCache,
            fx.As(new(hyperion.Cache)),
        ),
    ),
    fx.Invoke(RegisterHealthCheck), // Redis health check
)
```

**Configuration Example**:
```yaml
cache:
  type: redis
  mode: standalone         # standalone, cluster, sentinel
  addresses:
    - localhost:6379
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s

  # Cluster mode
  cluster:
    max_redirects: 3
    read_only: false

  # Sentinel mode
  sentinel:
    master_name: mymaster
    sentinel_addresses:
      - localhost:26379
```

**Advanced Features** (optional):
```go
// Pipeline support
type RedisCacheAdvanced interface {
    hyperion.Cache
    Pipeline() RedisPipeline
}

type RedisPipeline interface {
    Set(key string, value []byte, ttl time.Duration)
    Get(key string) *redis.StringCmd
    Exec(ctx context.Context) error
}
```

**Tasks**:
- [ ] Implement redisCache struct (1 day)
- [ ] Add standalone mode support (1 day)
- [ ] Add cluster mode support (1 day)
- [ ] Add sentinel mode support (1 day)
- [ ] Integrate with Config (1 day)
- [ ] Add health check integration (1 day)
- [ ] Write unit tests with miniredis (1 day)
- [ ] Write integration tests (real Redis) (1 day)
- [ ] Documentation and examples (1 day)

**Timeline**: 6 working days

---

### 4. Prometheus Metrics Integration 🔜

**Package**: `adapter/prometheus/`

**Scope**:
- Expose Prometheus metrics endpoint
- Framework metrics (request count, latency, error rate)
- Business metrics support
- Custom metric registration
- Integration with hyperweb (planned v2.3)

**Core Metrics Provider**:
```go
type MetricsProvider struct {
    registry *prometheus.Registry
}

// Framework metrics
var (
    RequestCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "hyperion_requests_total",
            Help: "Total number of requests",
        },
        []string{"method", "path", "status"},
    )

    RequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "hyperion_request_duration_seconds",
            Help:    "Request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )

    CacheHitRate = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "hyperion_cache_hit_rate",
            Help: "Cache hit rate",
        },
        []string{"cache_type"},
    )

    DatabaseConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "hyperion_db_connections_active",
            Help: "Active database connections",
        },
    )
)
```

**Module Export**:
```go
var Module = fx.Module("hyperion.adapter.prometheus",
    fx.Provide(NewMetricsProvider),
    fx.Invoke(RegisterDefaultMetrics),
    fx.Invoke(ExposeMetricsEndpoint), // HTTP /metrics endpoint
)
```

**Configuration Example**:
```yaml
metrics:
  enabled: true
  port: 9090              # Metrics endpoint port
  path: /metrics
  namespace: hyperion     # Metric prefix
  subsystem: api
  collect_runtime: true   # Go runtime metrics
```

**Custom Business Metrics**:
```go
func (s *UserService) RegisterUser(ctx hyperion.Context, req RegisterRequest) error {
    // Custom business metric
    userRegistrationCounter.Inc()

    start := time.Now()
    defer func() {
        registrationDuration.Observe(time.Since(start).Seconds())
    }()

    return s.uow.WithTransaction(ctx, func(txCtx hyperion.Context) error {
        // Business logic...
        return nil
    })
}
```

**Tasks**:
- [ ] Implement MetricsProvider (1 day)
- [ ] Define framework metrics (1 day)
- [ ] Add HTTP metrics endpoint (1 day)
- [ ] Add custom metric registration API (1 day)
- [ ] Integrate with Ristretto cache metrics (1 day)
- [ ] Write unit tests (1 day)
- [ ] Documentation and examples (1 day)

**Timeline**: 4 working days

---

### 5. Observability Example Application 🔜

**Package**: `examples/observability-demo/`

**Scope**:
- Microservice demonstrating full observability stack
- Distributed tracing across services
- Cache layer (Ristretto + Redis)
- Prometheus metrics dashboard
- Grafana/Jaeger integration guides

**Features**:
- Multi-service architecture (API Gateway + User Service + Order Service)
- Distributed tracing with context propagation
- Multi-level caching strategy
- Custom business metrics
- Docker Compose setup (Jaeger, Redis, Prometheus, Grafana)

**Architecture**:
```
API Gateway (Port 8080)
    ├── Traces: Jaeger
    ├── Cache: Ristretto (L1)
    └── Routes to:
        ├── User Service (Port 8081)
        │   ├── Cache: Redis (L2)
        │   └── Database: PostgreSQL
        └── Order Service (Port 8082)
            ├── Cache: Redis (L2)
            └── Database: PostgreSQL

Observability Stack:
    ├── Jaeger UI (Port 16686)
    ├── Prometheus (Port 9090)
    └── Grafana (Port 3000)
```

**Example Service Code**:
```go
// User Service with full observability
type UserService struct {
    userRepo hyperion.Repository
    uow      hyperion.UnitOfWork
}

func (s *UserService) GetUser(ctx hyperion.Context, id string) (*User, error) {
    // Distributed tracing
    ctx, span := ctx.Tracer().Start(ctx, "UserService.GetUser")
    defer span.End()

    span.SetAttributes("user.id", id)

    // L1 Cache (Ristretto)
    cacheKey := fmt.Sprintf("user:%s", id)
    if cached, err := cache.Get[User](ctx, cacheKey); err == nil {
        span.SetAttributes("cache.hit", true, "cache.layer", "ristretto")
        cacheHitCounter.WithLabelValues("ristretto", "user").Inc()
        return cached, nil
    }

    // L2 Cache (Redis)
    if cached, err := ctx.Cache().Get(ctx, cacheKey); err == nil {
        var user User
        json.Unmarshal(cached, &user)

        span.SetAttributes("cache.hit", true, "cache.layer", "redis")
        cacheHitCounter.WithLabelValues("redis", "user").Inc()

        // Warm L1 cache
        cache.Set(ctx, cacheKey, user, 5*time.Minute)
        return &user, nil
    }

    // Database query
    span.SetAttributes("cache.hit", false)
    cacheMissCounter.WithLabelValues("user").Inc()

    user, err := s.userRepo.FindByID(ctx, id)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }

    // Populate caches
    cache.Set(ctx, cacheKey, user, 5*time.Minute)  // L1
    ctx.Cache().Set(ctx, cacheKey, marshal(user), 15*time.Minute)  // L2

    return user, nil
}
```

**Docker Compose**:
```yaml
version: '3.8'

services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # UI
      - "14268:14268"  # Collector

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
```

**Tasks**:
- [ ] Set up multi-service architecture (2 days)
- [ ] Implement distributed tracing (1 day)
- [ ] Implement multi-level caching (1 day)
- [ ] Add custom business metrics (1 day)
- [ ] Create Docker Compose setup (1 day)
- [ ] Add Grafana dashboards (1 day)
- [ ] Write deployment guide (1 day)
- [ ] Add README with usage instructions (1 day)

**Timeline**: 6 working days

---

### 6. Integration Testing 🔜

**Scope**:
- Integration tests for OpenTelemetry adapter
- Integration tests for cache adapters (Ristretto, Redis)
- End-to-end observability tests

**Test Coverage**:
- [ ] OpenTelemetry with Jaeger exporter (Docker)
- [ ] Ristretto cache performance benchmarks
- [ ] Redis standalone mode (Docker)
- [ ] Redis cluster mode (Docker)
- [ ] Distributed tracing across services
- [ ] Cache hit/miss scenarios
- [ ] Metrics collection and exposure

**Tasks**:
- [ ] Set up test infrastructure (Docker Compose) (1 day)
- [ ] Write OpenTelemetry integration tests (1 day)
- [ ] Write Ristretto benchmark tests (1 day)
- [ ] Write Redis integration tests (1 day)
- [ ] Write E2E observability tests (1 day)

**Timeline**: 3 working days

---

## Implementation Timeline

### Week 1: Tracing Adapter (1 week)
- Days 1-2: Core OpenTelemetry implementation
- Days 3-4: Exporter support and configuration
- Days 5-7: Testing and documentation

### Week 2: Cache Adapters (1.5 weeks)
- Days 1-3: Ristretto adapter (implementation + tests)
- Days 4-6: Redis adapter (standalone)
- Days 7-8: Redis cluster/sentinel support

### Week 3: Metrics & Example App (1 week)
- Days 1-2: Prometheus metrics integration
- Days 3-6: Observability example application
- Day 7: Docker Compose and documentation

### Week 4: Integration & Release (0.5 week)
- Days 1-3: Integration testing
- Day 4: Final testing and bug fixes
- Day 5: Release v2.2

**Total**: 4 weeks

---

## Technical Challenges

### Challenge 1: Distributed Context Propagation
**Problem**: Ensure trace context propagates correctly across service boundaries
**Solution**: Use OpenTelemetry's built-in propagators (W3C Trace Context)
**Status**: Planned

### Challenge 2: Cache Stampede Prevention
**Problem**: Multiple goroutines requesting same cache key simultaneously
**Solution**: Implement singleflight pattern in cache adapters
**Status**: Planned

### Challenge 3: Redis Connection Pooling
**Problem**: Optimize Redis connection usage
**Solution**: Use go-redis with tuned pool settings, health checks
**Status**: Planned

### Challenge 4: Metric Cardinality Explosion
**Problem**: Too many label combinations causing high memory usage
**Solution**: Limit label values, use metric aggregation
**Status**: Planned

---

## Success Metrics

### Code Metrics
- Test coverage: >= 80% for all adapters
- Performance overhead: < 3% vs native libraries
- Lines of code: ~800 LOC (otel), ~600 LOC (ristretto), ~800 LOC (redis), ~400 LOC (prometheus)

### Quality Metrics
- golangci-lint: Zero errors
- Integration tests: All passing
- Example app: Fully functional observability stack

### Performance Benchmarks
- Ristretto cache: >90% hit rate, <100ns per operation
- Redis cache: <5ms p99 latency (local network)
- Tracing overhead: <1ms per span

### Community Metrics (Target)
- Production users: 20+ by end of v2.2
- GitHub stars: 500+
- Community adapters: 2+

---

## Next Epic

👉 **[Epic 4: Web Framework](epic-4-web-framework.md)** (v2.3 - Planned)

**Focus**: Gin web adapter, middleware suite, full-stack example application

**Timeline**: April 2026

---

## Related Documentation

- [Epic 1: Core Foundation](epic-1-core-foundation.md)
- [Epic 2: Essential Adapters](epic-2-essential-adapters.md)
- [Architecture Overview](../architecture.md)
- [Implementation Plan](../implementation-plan.md)

---

**Epic Status**: 🔜 **PLANNED** (February 2026)

**Last Updated**: October 2025
**Version**: 2.2 Planning
