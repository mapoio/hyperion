# Epic 3: OpenTelemetry Integration (Planned)

**Version**: Post-v2.2
**Status**: 🔜 **PLANNED** (Q1 2026)
**Duration**: 3 weeks
**Priority**: ⭐⭐⭐⭐

---

## Overview

Implement **real OpenTelemetry adapter** providing actual distributed tracing and metrics collection with automatic correlation. This builds on the Interceptor pattern and Meter interface already completed in v2.0-2.2.

**Current State (v2.0-2.2 - ✅ Completed)**:
- ✅ Interceptor pattern (`ctx.UseIntercept()`) for automatic observability
- ✅ Core interfaces: `hyperion.Tracer`, `hyperion.Meter` (OTel-compatible)
- ✅ Built-in interceptors: `TracingInterceptor`, `LoggingInterceptor`
- ✅ NoOp implementations (zero overhead by default)
- ✅ Zap adapter with OTel Logs Bridge for trace correlation

**What's Missing**:
- Real OpenTelemetry SDK integration for actual span export
- Real metrics collection with exemplar support
- Exporters: Jaeger, Prometheus, OTLP
- Sampling strategies and configuration

---

## Goals

### Primary Goals
1. **OpenTelemetry Tracer Adapter**: Real distributed tracing with span export
2. **OpenTelemetry Meter Adapter**: Real metrics collection with exemplar support
3. **Automatic Trace Correlation**: Logs → Traces → Metrics via OTel Logs Bridge
4. **Exporters**: Jaeger, Prometheus, OTLP integration
5. **Complete Observability Example**: Multi-service application with full OTel stack

### Success Criteria
- [ ] OpenTelemetry adapter passes all Tracer interface tests
- [ ] OpenTelemetry adapter passes all Meter interface tests
- [ ] Automatic exemplar support linking metrics to traces
- [ ] OTel Logs Bridge automatically extracts TraceID from context
- [ ] Exporters work with Jaeger, Prometheus, and OTLP
- [ ] Example application demonstrates full observability correlation
- [ ] Performance overhead < 5% vs NoOp implementations

---

## Deliverables

### 1. OpenTelemetry Tracer + Meter Adapter 🔜

**Package**: `adapter/otel/`

**Scope**:
- Implement `hyperion.Tracer` interface using OpenTelemetry SDK
- Implement `hyperion.Meter` interface using OpenTelemetry SDK
- Implement `hyperion.Span`, `hyperion.Counter`, `hyperion.Histogram`, etc.
- Support multiple exporters (Jaeger for traces, Prometheus for metrics, OTLP for both)
- Distributed context propagation (W3C Trace Context)
- **Automatic exemplar support**: Metrics include trace samples
- **OTel Logs Bridge integration**: Logs automatically include TraceID/SpanID
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

**Meter Implementation** (New):
```go
type otelMeter struct {
    meter metric.Meter
}

func (m *otelMeter) Counter(name string, opts ...hyperion.MetricOption) hyperion.Counter {
    config := buildMetricConfig(opts...)
    counter, _ := m.meter.Int64Counter(name,
        metric.WithDescription(config.Description),
        metric.WithUnit(config.Unit),
    )
    return &otelCounter{counter: counter}
}

type otelCounter struct {
    counter metric.Int64Counter
}

func (c *otelCounter) Add(ctx context.Context, value int64, attrs ...hyperion.Attribute) {
    // Automatic exemplar support - links metric to current trace
    c.counter.Add(ctx, value, convertAttributes(attrs)...)
}
```

**Context Integration with Interceptor** (Recommended):
```go
func (s *UserService) GetUser(ctx hyperion.Context, id string) (_ *User, err error) {
    // 3-Line Interceptor Pattern - Automatic tracing, logging, and metrics
    ctx, end := ctx.UseIntercept("UserService", "GetUser")
    defer end(&err)

    // Business logic with manual metrics
    counter := ctx.Meter().Counter("user.lookups")
    counter.Add(ctx, 1, hyperion.String("method", "GetUser"))

    user, err := s.userRepo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    return user, nil
}
```

**Alternative: Manual Tracing** (Fine-grained control):
```go
func (s *UserService) GetUserManual(ctx hyperion.Context, id string) (*User, error) {
    ctx, span := ctx.Tracer().Start(ctx, "UserService.GetUser")
    defer span.End()

    span.SetAttributes(hyperion.String("user.id", id))

    user, err := s.userRepo.FindByID(ctx, id)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }

    span.SetAttributes(hyperion.Bool("user.found", true))
    return user, nil
}
```

**Tasks**:
- [ ] Implement otelTracer struct and otelSpan wrapper (2 days)
- [ ] Implement otelMeter struct (Counter, Histogram, Gauge, UpDownCounter) (2 days)
- [ ] Add trace exporters (Jaeger, Zipkin, OTLP) (1 day)
- [ ] Add metrics exporters (Prometheus, OTLP) (1 day)
- [ ] Implement automatic exemplar support (link metrics to traces) (1 day)
- [ ] Integrate with Config for tracer and meter (1 day)
- [ ] Add W3C Trace Context propagation (1 day)
- [ ] Add OTel Logs Bridge for automatic log correlation (1 day)
- [ ] Write unit tests (>80% coverage) (1 day)
- [ ] Write integration tests (real exporters) (1 day)
- [ ] Documentation and examples (1 day)

**Timeline**: 10 working days (~2 weeks)

---

### 2. Observability Example Application 🔜

**Package**: `examples/observability-demo/`

**Scope**:
- Multi-service application demonstrating full OpenTelemetry integration
- Distributed tracing across services
- Metrics with exemplar support
- Logs with automatic trace correlation
- Docker Compose setup (Jaeger, Prometheus, Grafana)

**Features**:
- Multi-service architecture (API Gateway + User Service + Order Service)
- 3-Line Interceptor Pattern for automatic observability
- Distributed tracing with W3C Trace Context propagation
- Custom business metrics with automatic trace exemplars
- Logs automatically include TraceID and SpanID
- Complete observability stack setup

**Architecture**:
```
API Gateway (Port 8080)
    ├── Traces: Jaeger (via OTLP)
    ├── Metrics: Prometheus
    ├── Logs: stdout with trace context
    └── Routes to:
        ├── User Service (Port 8081)
        │   └── Database: PostgreSQL
        └── Order Service (Port 8082)
            └── Database: PostgreSQL

Observability Stack:
    ├── Jaeger UI (Port 16686)       # Distributed traces
    ├── Prometheus (Port 9090)       # Metrics with exemplars
    └── Grafana (Port 3000)          # Unified dashboards
```

**Example Service Code**:
```go
// User Service with full OpenTelemetry integration
type UserService struct {
    userRepo hyperion.Repository
    uow      hyperion.UnitOfWork
}

func (s *UserService) GetUser(ctx hyperion.Context, id string) (_ *User, err error) {
    // 3-Line Interceptor Pattern
    ctx, end := ctx.UseIntercept("UserService", "GetUser")
    defer end(&err)

    // Log automatically includes TraceID and SpanID
    ctx.Logger().Info("fetching user", "user_id", id)

    // Metric automatically includes exemplar linking to current trace
    lookupCounter := ctx.Meter().Counter("user.lookups")
    lookupCounter.Add(ctx, 1, hyperion.String("method", "GetUser"))

    // Database query
    user, err := s.userRepo.FindByID(ctx, id)
    if err != nil {
        ctx.Logger().Error("user not found", "user_id", id, "error", err)
        return nil, err
    }

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
      - "4318:4318"    # OTLP HTTP receiver
      - "4317:4317"    # OTLP gRPC receiver

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--enable-feature=exemplar-storage'  # Enable exemplar support

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"

  api-gateway:
    build: ./services/api-gateway
    ports:
      - "8080:8080"
    environment:
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318
      - PROMETHEUS_ENDPOINT=:9090

  user-service:
    build: ./services/user-service
    ports:
      - "8081:8081"
    environment:
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318
      - POSTGRES_DSN=postgres://postgres:postgres@postgres:5432/users

  order-service:
    build: ./services/order-service
    ports:
      - "8082:8082"
    environment:
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318
      - POSTGRES_DSN=postgres://postgres:postgres@postgres:5432/orders
```

**Observability Correlation Flow**:
```
1. User request → API Gateway
   ├── TracingInterceptor creates root span with TraceID
   └── Context contains TraceID + SpanID

2. API Gateway → User Service (HTTP call)
   ├── W3C Trace Context propagated via HTTP headers
   └── User Service continues same trace

3. User Service business logic
   ├── Logger.Info() → includes TraceID + SpanID (via OTel Logs Bridge)
   ├── Meter.Counter.Add() → includes exemplar with TraceID + SpanID
   └── Tracer creates child span

4. Observability Backends
   ├── Jaeger: Shows distributed trace across services
   ├── Prometheus: Metrics with exemplars linking to Jaeger traces
   └── Logs: Include TraceID for correlation with traces
```

**Tasks**:
- [ ] Set up multi-service architecture (2 days)
- [ ] Implement distributed tracing with W3C propagation (1 day)
- [ ] Add business metrics with exemplar support (1 day)
- [ ] Configure OTel Logs Bridge for log correlation (1 day)
- [ ] Create Docker Compose setup (1 day)
- [ ] Add Grafana dashboards showing trace-metric correlation (1 day)
- [ ] Write deployment guide (1 day)
- [ ] Add README with usage instructions (1 day)

**Timeline**: 6 working days

---

### 3. Caching Adapters 🔜 (Moved to Future Epic)

**Note**: Cache adapters (Ristretto, Redis) have been moved to a future Epic focused on performance optimization. Epic 3 focuses exclusively on OpenTelemetry integration for observability.

**Deferred to**: Epic 6 or later (TBD)

---

### OLD SECTION (Cache Adapters) - REMOVED

The following sections have been removed from Epic 3 and deferred to a future Epic:
- ~~Ristretto Cache Adapter~~
- ~~Redis Cache Adapter~~
- ~~Prometheus Metrics Integration~~ (Replaced by OTel Meter with Prometheus exporter)

---

### 2. (REMOVED - See Section 3) Ristretto Cache Adapter 🔜

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

### Week 1-2: OpenTelemetry Adapter (2 weeks)
- Days 1-2: Implement otelTracer and otelSpan
- Days 3-4: Implement otelMeter (Counter, Histogram, Gauge, UpDownCounter)
- Days 5-6: Add exporters (Jaeger, Prometheus, OTLP)
- Days 7-8: Implement automatic exemplar support
- Day 9: Add W3C Trace Context propagation
- Day 10: Testing and documentation

### Week 3: Example Application & Integration (1 week)
- Days 1-3: Build multi-service observability demo
- Days 4-5: Integration testing with real exporters
- Days 6-7: Docker Compose setup and Grafana dashboards

**Total**: 3 weeks (~15 working days)

---

## Technical Challenges

### Challenge 1: Distributed Context Propagation
**Problem**: Ensure trace context propagates correctly across service boundaries
**Solution**: Use OpenTelemetry's built-in W3C Trace Context propagators
**Status**: Planned

### Challenge 2: Automatic Exemplar Support
**Problem**: Link metrics to traces automatically without manual TraceID extraction
**Solution**: Use OpenTelemetry SDK's automatic exemplar support via context
**Status**: Planned

### Challenge 3: OTel Logs Bridge Integration
**Problem**: Ensure logs automatically include TraceID and SpanID
**Solution**: Use OTel Logs Bridge API to extract trace context from `context.Context`
**Status**: Planned (Zap adapter already has bridge code, needs OTel SDK integration)

### Challenge 4: Metric Cardinality Explosion
**Problem**: Too many label combinations causing high memory usage in Prometheus
**Solution**:
- Limit attribute cardinality in metric definitions
- Use aggregation strategies
- Document best practices for metric labels
**Status**: Planned

---

## Success Metrics

### Code Metrics
- Test coverage: >= 80% for OTel adapter
- Performance overhead: < 5% vs NoOp implementations
- Lines of code: ~1000 LOC (otel tracer + meter + exporters)

### Quality Metrics
- golangci-lint: Zero errors
- Integration tests: All passing with real exporters (Jaeger, Prometheus)
- Example app: Demonstrates full trace-metric-log correlation

### Performance Benchmarks
- Tracing overhead: <1ms per span creation
- Metrics overhead: <100ns per metric operation (with exemplars)
- Context propagation: <50ns overhead

### Observability Metrics
- **Trace Correlation**: 100% of logs include TraceID when using Interceptor pattern
- **Metric Exemplars**: 100% of metric samples link to traces automatically
- **Cross-Service Traces**: Successfully propagate across 3+ microservices

### Community Metrics (Target)
- Production users: 15+ by Q2 2026
- GitHub stars: 600+
- Documentation examples: 5+ complete scenarios

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

**Epic Status**: 🔜 **PLANNED** (Q1 2026)

**Last Updated**: October 2025 (Updated)
**Version**: Post-v2.2 Planning
