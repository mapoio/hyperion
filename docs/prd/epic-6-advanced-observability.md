# Epic 6: Advanced Observability Features

**Status**: 🔜 Planned
**Target**: Q2 2026
**Priority**: Medium
**Dependencies**: Epic 3 (OpenTelemetry Integration)

---

## Overview

Enhance Hyperion's observability capabilities with advanced features for production environments, including custom trace sampling, context propagation strategies, and enhanced instrumentation.

## Background

While Epic 3 delivered core OpenTelemetry integration (Tracer, Meter, and trace correlation), production environments often require fine-grained control over:
- Sampling strategies (reduce trace overhead while maintaining visibility)
- Cross-process context propagation (support for multiple propagation formats)
- Custom instrumentation hooks (framework-agnostic observability)
- Performance optimization (reduce observability overhead)

## Goals

### Primary Goals

1. **Flexible Sampling** - Support custom sampling strategies beyond AlwaysSample/NeverSample
2. **Multi-Format Propagation** - Support W3C TraceContext, B3, and custom propagators
3. **Extensible Instrumentation** - Plugin-based instrumentation for third-party libraries
4. **Performance Optimization** - Minimize observability overhead in hot paths

### Non-Goals

- Building a complete APM solution (use existing solutions like Jaeger, Grafana, HyperDX)
- Replacing OpenTelemetry SDK (we're an adapter, not a replacement)
- Custom trace storage/query (that's the backend's job)

## Success Metrics

- [ ] Support at least 3 different sampling strategies
- [ ] Support W3C TraceContext and B3 propagation formats
- [ ] Observability overhead < 5% in production workloads
- [ ] 90%+ test coverage for all new features
- [ ] Complete documentation with production examples

---

## Stories

### Story 6.1: Custom Sampling Strategies

**Priority**: P0
**Effort**: 5 points
**Dependencies**: Epic 3 complete

#### User Story

> As a **production engineer**,
> I want to **configure custom sampling strategies**,
> so that **I can control trace volume and reduce costs while maintaining visibility into critical transactions**.

#### Acceptance Criteria

1. Support built-in samplers:
   - Probability sampler (sample X% of traces)
   - Rate-limiting sampler (max N traces per second)
   - Parent-based sampler (follow parent's sampling decision)
   - Rule-based sampler (sample based on attributes, URLs, errors)

2. Allow custom sampler implementation:
   ```go
   type Sampler interface {
       ShouldSample(ctx context.Context, traceID string, spanName string, attrs ...Attribute) SamplingResult
   }
   ```

3. Configuration via adapter options:
   ```go
   tracer := hyperotel.NewOtelTracer(
       hyperotel.WithSampler(
           hyperotel.ProbabilitySampler(0.1), // Sample 10% of traces
       ),
   )
   ```

4. Support sampler composition:
   ```go
   sampler := hyperotel.CompositeSampler(
       hyperotel.AlwaysSampleIfError(),
       hyperotel.AlwaysSampleIfSlowRequest(500*time.Millisecond),
       hyperotel.ProbabilitySampler(0.01), // Sample 1% of normal requests
   )
   ```

#### Technical Design

**Core Interface** (`hyperion/tracer.go`):
```go
// Sampler determines whether a span should be sampled
type Sampler interface {
    // ShouldSample returns a sampling decision for the given trace
    ShouldSample(params SamplingParameters) SamplingResult
}

type SamplingParameters struct {
    ParentContext  Context
    TraceID        string
    Name           string
    Kind           SpanKind
    Attributes     []Attribute
    Links          []Link
}

type SamplingResult struct {
    Decision   SamplingDecision
    Attributes []Attribute // Additional attributes to add to span
}

type SamplingDecision int

const (
    Drop SamplingDecision = iota  // Do not record or sample
    RecordOnly                    // Record but do not sample
    RecordAndSample               // Record and sample
)
```

**OTel Adapter Implementation** (`adapter/otel/sampler.go`):
```go
// Built-in samplers
func ProbabilitySampler(probability float64) Sampler
func RateLimitingSampler(maxPerSecond int) Sampler
func ParentBasedSampler(root Sampler) Sampler
func RuleBasedSampler(rules ...SamplingRule) Sampler

// Sampler composition
func CompositeSampler(samplers ...Sampler) Sampler

// Smart samplers
func AlwaysSampleIfError() Sampler
func AlwaysSampleIfSlowRequest(threshold time.Duration) Sampler
func AlwaysSampleIfAttribute(key string, value any) Sampler
```

#### Testing Strategy

- Unit tests for each sampler type
- Integration tests with OTel SDK
- Performance benchmarks (overhead measurement)
- Property-based testing for probability sampler (actual rate ≈ configured rate)

---

### Story 6.2: Multi-Format Context Propagation

**Priority**: P1
**Effort**: 3 points
**Dependencies**: Story 6.1

#### User Story

> As a **microservices architect**,
> I want to **support multiple trace propagation formats**,
> so that **Hyperion services can interoperate with legacy systems using different tracing protocols**.

#### Acceptance Criteria

1. Support propagation formats:
   - W3C TraceContext (default, standard)
   - B3 (Zipkin compatibility)
   - Jaeger (legacy compatibility)
   - Custom propagators

2. Automatic propagator selection based on incoming headers

3. Configuration via adapter options:
   ```go
   tracer := hyperotel.NewOtelTracer(
       hyperotel.WithPropagators(
           hyperotel.W3CTraceContext,
           hyperotel.B3Propagator,
       ),
   )
   ```

4. Convenience methods for HTTP client/server:
   ```go
   // HTTP Server - extract context
   hctx := factory.New(req.Context())
   hctx = propagator.Extract(hctx, req.Header)

   // HTTP Client - inject context
   propagator.Inject(hctx, req.Header)
   ```

#### Technical Design

**Core Interface** (`hyperion/propagator.go`):
```go
// Propagator handles trace context propagation across process boundaries
type Propagator interface {
    // Inject injects trace context into carrier (e.g., HTTP headers)
    Inject(ctx Context, carrier Carrier) error

    // Extract extracts trace context from carrier
    Extract(ctx Context, carrier Carrier) (Context, error)

    // Fields returns the fields this propagator uses
    Fields() []string
}

// Carrier is a key-value abstraction for propagation
type Carrier interface {
    Get(key string) string
    Set(key string, value string)
    Keys() []string
}

// HTTPHeaderCarrier wraps http.Header
type HTTPHeaderCarrier http.Header

func (c HTTPHeaderCarrier) Get(key string) string
func (c HTTPHeaderCarrier) Set(key, value string)
func (c HTTPHeaderCarrier) Keys() []string
```

**OTel Adapter Implementation** (`adapter/otel/propagator.go`):
```go
var (
    W3CTraceContext Propagator = &w3cPropagator{}
    B3Propagator    Propagator = &b3Propagator{}
    JaegerPropagator Propagator = &jaegerPropagator{}
)

func CompositePropagator(propagators ...Propagator) Propagator
```

#### Testing Strategy

- Unit tests for each propagator format
- Integration tests with microservices (cross-service trace continuity)
- Interoperability tests with Zipkin, Jaeger
- HTTP header injection/extraction tests

---

### Story 6.3: Plugin-Based Instrumentation

**Priority**: P2
**Effort**: 8 points
**Dependencies**: Story 6.1, 6.2

#### User Story

> As a **framework user**,
> I want to **easily add instrumentation to third-party libraries**,
> so that **I get automatic tracing for database drivers, HTTP clients, message queues, etc. without manual span creation**.

#### Acceptance Criteria

1. Plugin interface for automatic instrumentation:
   ```go
   type Instrumentation interface {
       Name() string
       Install(ctx context.Context) error
       Uninstall(ctx context.Context) error
   }
   ```

2. Built-in instrumentations:
   - `instrumentation/http` - HTTP client/server
   - `instrumentation/sql` - database/sql
   - `instrumentation/redis` - Redis client
   - `instrumentation/grpc` - gRPC client/server

3. Registration via fx module:
   ```go
   fx.New(
       hyperion.CoreModule,
       otel.Module,
       instrumentation.HTTPModule,    // Auto-instrument HTTP
       instrumentation.SQLModule,     // Auto-instrument DB
       instrumentation.RedisModule,   // Auto-instrument Redis
       // Your app...
   )
   ```

4. Opt-out mechanism for specific libraries

#### Technical Design

**Core Package** (`instrumentation/`):
```
instrumentation/
├── instrumentation.go    # Core interface
├── http/                 # HTTP instrumentation
│   ├── client.go
│   ├── server.go
│   └── module.go
├── sql/                  # SQL instrumentation
│   ├── driver.go
│   └── module.go
└── registry.go           # Plugin registry
```

**Example: HTTP Client Instrumentation**:
```go
package http

import (
    "net/http"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewInstrumentedHTTPClient(tracer hyperion.Tracer) *http.Client {
    otelTracer := tracer.(*hyperotel.OtelTracer)
    return &http.Client{
        Transport: otelhttp.NewTransport(
            http.DefaultTransport,
            otelhttp.WithTracerProvider(otelTracer.TracerProvider()),
        ),
    }
}

var HTTPModule = fx.Module("instrumentation.http",
    fx.Provide(NewInstrumentedHTTPClient),
)
```

#### Testing Strategy

- Integration tests for each instrumentation plugin
- E2E tests with real third-party libraries
- Performance benchmarks (overhead measurement)
- Compatibility matrix (library versions vs. instrumentation)

---

## Implementation Plan

### Phase 1: Foundation (Weeks 1-2)

- [ ] Design and review core interfaces (`Sampler`, `Propagator`)
- [ ] Implement core types and tests
- [ ] Create ADR (Architecture Decision Record)

### Phase 2: Sampling (Weeks 3-4)

- [ ] Implement built-in samplers (Story 6.1)
- [ ] Add configuration support
- [ ] Write comprehensive tests
- [ ] Performance benchmarks

### Phase 3: Propagation (Weeks 5-6)

- [ ] Implement W3C TraceContext propagator (Story 6.2)
- [ ] Implement B3 propagator
- [ ] Add HTTP helper utilities
- [ ] Interoperability tests

### Phase 4: Instrumentation (Weeks 7-10)

- [ ] Design plugin architecture (Story 6.3)
- [ ] Implement HTTP instrumentation
- [ ] Implement SQL instrumentation
- [ ] Implement Redis instrumentation
- [ ] Integration tests

### Phase 5: Documentation & Release (Weeks 11-12)

- [ ] Write comprehensive documentation
- [ ] Create production examples
- [ ] Performance tuning guide
- [ ] Release v3.0.0

---

## Technical Considerations

### Performance Impact

- **Sampling overhead**: < 1µs per sampling decision
- **Propagation overhead**: < 5µs per HTTP request
- **Instrumentation overhead**: < 2% for instrumented operations

### Backward Compatibility

- All new features are **opt-in**
- Existing applications continue to work without changes
- Default behavior unchanged (AlwaysSample, W3C TraceContext)

### Dependencies

- `go.opentelemetry.io/otel/sdk/trace` - for sampler integration
- `go.opentelemetry.io/contrib/propagators` - for B3, Jaeger propagators
- `go.opentelemetry.io/contrib/instrumentation/*` - for library instrumentation

---

## Documentation Updates

### New Documentation

- [ ] Sampling Guide (`docs/guides/sampling.md`)
- [ ] Propagation Guide (`docs/guides/propagation.md`)
- [ ] Instrumentation Guide (`docs/guides/instrumentation.md`)
- [ ] Performance Tuning Guide (`docs/guides/performance-tuning.md`)

### Updated Documentation

- [ ] Update `docs/architecture.md` with new interfaces
- [ ] Update `docs/observability.md` with advanced features
- [ ] Update `adapter/otel/README.md` with new options
- [ ] Update `CLAUDE.md` with new patterns

---

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| OTel SDK API changes | High | Medium | Pin to stable OTel version, abstract via interfaces |
| Performance regression | High | Low | Comprehensive benchmarks, performance gates in CI |
| Instrumentation compatibility | Medium | Medium | Maintain compatibility matrix, version testing |
| Complex configuration | Medium | High | Provide sensible defaults, clear examples |

---

## Open Questions

1. **Q**: Should we support OpenCensus propagators for legacy compatibility?
   **A**: TBD - Need to assess user demand

2. **Q**: Should instrumentation plugins be in main repo or separate?
   **A**: TBD - Lean towards main repo for better integration testing

3. **Q**: How to handle instrumentation version conflicts?
   **A**: TBD - May need plugin versioning strategy

---

## References

- [OpenTelemetry Sampling Spec](https://opentelemetry.io/docs/specs/otel/trace/sdk/#sampling)
- [W3C TraceContext Specification](https://www.w3.org/TR/trace-context/)
- [B3 Propagation Specification](https://github.com/openzipkin/b3-propagation)
- [OpenTelemetry Contrib Instrumentation](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation)

---

**Last Updated**: 2025-10-04
**Version**: 1.0
**Status**: Draft
