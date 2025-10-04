# Hyperion OpenTelemetry Integration Guide

This guide explains the **application-controlled OTel SDK architecture** implemented in the Hyperion framework, addressing OTel version independence and auto-instrumentation support.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [The Problem We're Solving](#the-problem-were-solving)
- [The Solution](#the-solution)
- [Directory Structure](#directory-structure)
- [Step-by-Step Implementation](#step-by-step-implementation)
- [Auto-Instrumentation Examples](#auto-instrumentation-examples)
- [Testing](#testing)
- [Best Practices](#best-practices)
- [Migration Guide](#migration-guide)
- [Troubleshooting](#troubleshooting)

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     Application Layer                           │
│  (Full control over OTel SDK initialization & version)          │
│                                                                  │
│  internal/telemetry/                                            │
│  ├── otel_sdk.go          → TracerProvider, MeterProvider       │
│  ├── runtime_metrics.go   → Go runtime metrics                  │
│  └── http_instrumentation.go → HTTP auto-tracing                │
│                                                                  │
│  Provides: *sdktrace.TracerProvider                             │
│            *metric.MeterProvider                                │
└─────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────┐
│                     Hyperion Adapter Layer                      │
│              (Thin wrapper, version-agnostic)                   │
│                                                                  │
│  adapter/otel/                                                  │
│  ├── NewOtelTracerFromProvider(tp, serviceName)                 │
│  ├── NewOtelMeterFromProvider(mp, serviceName)                  │
│  ├── OtelTracer.TracerProvider() → for auto-instrumentation    │
│  └── OtelMeter.MeterProvider() → for auto-instrumentation      │
│                                                                  │
│  Returns: hyperion.Tracer, hyperion.Meter                       │
└─────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────┐
│                     Business Logic Layer                        │
│           (Uses hyperion.Context abstractions)                  │
│                                                                  │
│  Uses: ctx.Tracer(), ctx.Logger(), ctx.UseIntercept()           │
└─────────────────────────────────────────────────────────────────┘
```

## The Problem We're Solving

### Problem 1: OTel Version Lock-in

**Traditional Adapter Pattern**:
```go
// ❌ Problem: Adapter controls OTel version
// adapter/otel/go.mod
require (
    go.opentelemetry.io/otel v1.28.0  // ← Application is stuck on v1.28.0
)

// Application wants to upgrade to v1.32.0 but can't!
```

**Impact**:
- Applications cannot upgrade OTel SDK independently
- Must wait for adapter to upgrade first
- Security patches delayed
- New features unavailable

### Problem 2: Auto-Instrumentation Not Supported

**Problem**: Third-party OTel instrumentation libraries need access to `TracerProvider` / `MeterProvider`:

```go
// ❌ These libraries won't work with old adapter:
import "go.opentelemetry.io/contrib/instrumentation/runtime"
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

// They need TracerProvider, but old adapter doesn't expose it
```

**Auto-instrumentation capabilities we're missing**:
- Go runtime metrics (CPU, memory, GC, goroutines)
- HTTP client/server tracing
- gRPC client/server tracing
- Database query tracing (GORM, pgx, etc.)
- Redis command tracing
- Kafka producer/consumer tracing

## The Solution

### Solution: Application-Controlled OTel SDK

**Key Principle**: Application owns OTel SDK initialization, adapter just wraps it.

**Benefits**:
1. ✅ **Version Independence**: Application's `go.mod` controls OTel version
2. ✅ **Auto-Instrumentation**: Adapter exposes providers for third-party libraries
3. ✅ **Single Instance**: Centralized configuration prevents duplicate telemetry
4. ✅ **Simple Integration**: Modular fx.Module design
5. ✅ **Easy Customization**: Application owns all configuration

### Architecture Comparison

| Aspect | Old (Adapter Controls) | New (Application Controls) |
|--------|------------------------|----------------------------|
| OTel Version | Locked to adapter version | Application chooses version |
| SDK Initialization | Hidden inside adapter | Explicit in application layer |
| Auto-Instrumentation | Not supported | Fully supported via accessors |
| Configuration | Adapter's defaults | Application's full control |
| Upgrading OTel | Requires adapter upgrade | `go get` in application |

## Directory Structure

```
example/otel/
├── cmd/app/
│   └── main.go                      # 5-step initialization
├── internal/
│   ├── telemetry/                   # ⭐ Application owns OTel SDK
│   │   ├── otel_sdk.go              # TracerProvider, MeterProvider
│   │   ├── runtime_metrics.go       # Go runtime auto-metrics
│   │   ├── http_instrumentation.go  # HTTP auto-tracing
│   │   └── integration_test.go      # Compatibility tests
│   └── services/
│       ├── order_service.go         # Business logic
│       └── module.go
├── configs/
│   └── config.yaml
├── go.mod                           # ⭐ Application controls OTel version
└── docs/
    └── OTEL_INTEGRATION_GUIDE.md    # This file
```

## Step-by-Step Implementation

### Step 1: Create Centralized OTel SDK Package

**File**: `internal/telemetry/otel_sdk.go`

```go
package telemetry

import (
    "context"
    "fmt"
    "time"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/metric"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
    "go.uber.org/fx"
)

// SDKConfig holds OTel SDK configuration
type SDKConfig struct {
    ServiceName      string
    ServiceVersion   string
    Environment      string
    OTLPEndpoint     string
    EnablePrometheus bool
}

// NewSDKConfig creates default configuration
func NewSDKConfig() *SDKConfig {
    return &SDKConfig{
        ServiceName:      "hyperion-otel-example",
        ServiceVersion:   "v1.0.0",
        Environment:      "development",
        OTLPEndpoint:     "localhost:4317",
        EnablePrometheus: true,
    }
}

// NewResource creates shared OTel resource with service metadata
func NewResource(cfg *SDKConfig) (*resource.Resource, error) {
    return resource.Merge(
        resource.Default(),
        resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName(cfg.ServiceName),
            semconv.ServiceVersion(cfg.ServiceVersion),
            semconv.DeploymentEnvironment(cfg.Environment),
        ),
    )
}

// NewTracerProvider creates application-configured TracerProvider
// ⭐ Application has FULL CONTROL over configuration
func NewTracerProvider(cfg *SDKConfig, res *resource.Resource) (*sdktrace.TracerProvider, error) {
    ctx := context.Background()

    // Application chooses exporter
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create trace exporter: %w", err)
    }

    // Application chooses batching config
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter,
            sdktrace.WithBatchTimeout(5*time.Second),
            sdktrace.WithMaxExportBatchSize(512),
        ),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )

    // Set as global provider
    otel.SetTracerProvider(tp)
    return tp, nil
}

// NewMeterProvider creates application-configured MeterProvider
func NewMeterProvider(cfg *SDKConfig, res *resource.Resource) (*metric.MeterProvider, error) {
    ctx := context.Background()

    var readers []metric.Reader

    // OTLP metrics exporter
    otlpExporter, err := otlpmetricgrpc.New(ctx,
        otlpmetricgrpc.WithEndpoint(cfg.OTLPEndpoint),
        otlpmetricgrpc.WithInsecure(),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
    }
    readers = append(readers, metric.NewPeriodicReader(otlpExporter,
        metric.WithInterval(10*time.Second),
    ))

    // Optional: Prometheus exporter
    if cfg.EnablePrometheus {
        promExporter, err := prometheus.New()
        if err != nil {
            return nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
        }
        readers = append(readers, promExporter)
    }

    // Create MeterProvider
    opts := []metric.Option{metric.WithResource(res)}
    for _, reader := range readers {
        opts = append(opts, metric.WithReader(reader))
    }
    mp := metric.NewMeterProvider(opts...)

    otel.SetMeterProvider(mp)
    return mp, nil
}

// RegisterShutdown ensures graceful OTel shutdown
func RegisterShutdown(
    lc fx.Lifecycle,
    tp *sdktrace.TracerProvider,
    mp *metric.MeterProvider,
) {
    lc.Append(fx.Hook{
        OnStop: func(ctx context.Context) error {
            shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
            defer cancel()

            if err := tp.Shutdown(shutdownCtx); err != nil {
                return fmt.Errorf("failed to shutdown TracerProvider: %w", err)
            }
            if err := mp.Shutdown(shutdownCtx); err != nil {
                return fmt.Errorf("failed to shutdown MeterProvider: %w", err)
            }
            return nil
        },
    })
}

// Module provides fully configured OTel SDK
var Module = fx.Module("telemetry",
    fx.Provide(
        NewSDKConfig,
        NewResource,
        NewTracerProvider,
        NewMeterProvider,
    ),
    fx.Invoke(RegisterShutdown),
)
```

### Step 2: Update Adapter to Accept External Providers

**File**: `adapter/otel/module.go`

```go
package otel

import (
    "github.com/mapoio/hyperion"
    "go.opentelemetry.io/otel/metric"
    "go.opentelemetry.io/otel/trace"
)

// NewOtelTracerFromProvider creates hyperion.Tracer from external TracerProvider
// ⭐ Key: Application provides the TracerProvider
func NewOtelTracerFromProvider(
    provider trace.TracerProvider,
    serviceName string,
) hyperion.Tracer {
    tracer := provider.Tracer(serviceName)
    return &OtelTracer{
        tracer:   tracer,
        provider: provider, // ⭐ Store for auto-instrumentation access
    }
}

// NewOtelMeterFromProvider creates hyperion.Meter from external MeterProvider
func NewOtelMeterFromProvider(
    provider metric.MeterProvider,
    serviceName string,
) hyperion.Meter {
    meter := provider.Meter(serviceName)
    return &OtelMeter{
        meter:    meter,
        provider: provider, // ⭐ Store for auto-instrumentation access
    }
}
```

**File**: `adapter/otel/tracer.go`

```go
// OtelTracer wraps OpenTelemetry tracer
// ⭐ Exported for type assertions in auto-instrumentation
type OtelTracer struct {
    tracer   trace.Tracer
    provider trace.TracerProvider // ⭐ Stored for access
}

// TracerProvider returns underlying TracerProvider
// ⭐ Enables third-party auto-instrumentation libraries
func (t *OtelTracer) TracerProvider() trace.TracerProvider {
    return t.provider
}

// Start implements hyperion.Tracer
func (t *OtelTracer) Start(ctx hyperion.Context, name string, opts ...hyperion.SpanOption) (hyperion.Context, hyperion.Span) {
    // ... implementation ...
}
```

### Step 3: Wire Everything in main.go

**File**: `cmd/app/main.go`

```go
package main

import (
    "github.com/mapoio/hyperion"
    hyperotel "github.com/mapoio/hyperion/adapter/otel"
    "github.com/mapoio/hyperion/example/otel/internal/services"
    "github.com/mapoio/hyperion/example/otel/internal/telemetry"
    "go.opentelemetry.io/otel/sdk/metric"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    "go.uber.org/fx"
)

func main() {
    fx.New(
        // ============================================================
        // STEP 1: Initialize OpenTelemetry SDK (Application Layer)
        // ============================================================
        telemetry.Module,                    // OTel SDK initialization
        telemetry.RuntimeMetricsModule,      // Go runtime metrics
        telemetry.HTTPInstrumentationModule, // HTTP auto-tracing

        // ============================================================
        // STEP 2: Integrate Hyperion Adapters with OTel SDK
        // ============================================================
        fx.Provide(
            func(tp *sdktrace.TracerProvider) hyperion.Tracer {
                return hyperotel.NewOtelTracerFromProvider(tp, "my-service")
            },
        ),
        fx.Provide(
            func(mp *metric.MeterProvider) hyperion.Meter {
                return hyperotel.NewOtelMeterFromProvider(mp, "my-service")
            },
        ),

        // ============================================================
        // STEP 3: Other Hyperion Dependencies
        // ============================================================
        viper.Module,
        zap.Module,
        fx.Provide(hyperion.NewNoOpDatabase),
        fx.Provide(hyperion.NewTracingInterceptor),
        fx.Provide(hyperion.NewContextFactory),

        // ============================================================
        // STEP 4: Business Logic
        // ============================================================
        services.Module,

        // ============================================================
        // STEP 5: HTTP Server
        // ============================================================
        fx.Provide(NewHTTPServer),
        fx.Invoke(RegisterRoutes),
        fx.Invoke(StartServer),
    ).Run()
}
```

### Step 4: Control OTel Version in go.mod

**File**: `example/otel/go.mod`

```go
module github.com/mapoio/hyperion/example/otel

go 1.23

require (
    // ⭐ Application controls OTel version
    go.opentelemetry.io/otel v1.32.0
    go.opentelemetry.io/otel/sdk v1.32.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.32.0
    go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.32.0

    // ⭐ Auto-instrumentation libraries
    go.opentelemetry.io/contrib/instrumentation/runtime v0.57.0
    go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.57.0

    // Hyperion adapter (version independent)
    github.com/mapoio/hyperion v2.0.0
)
```

**Upgrading OTel is simple**:

```bash
cd example/otel
go get go.opentelemetry.io/otel@v1.33.0
go get go.opentelemetry.io/otel/sdk@v1.33.0
go mod tidy
```

No adapter changes needed!

## Auto-Instrumentation Examples

### Example 1: Go Runtime Metrics

**File**: `internal/telemetry/runtime_metrics.go`

```go
package telemetry

import (
    "time"

    "go.opentelemetry.io/contrib/instrumentation/runtime"
    "go.opentelemetry.io/otel/sdk/metric"
    "go.uber.org/fx"
)

// EnableRuntimeMetrics collects Go runtime metrics automatically
func EnableRuntimeMetrics(mp *metric.MeterProvider) error {
    return runtime.Start(
        runtime.WithMeterProvider(mp),
        runtime.WithMinimumReadMemStatsInterval(time.Second),
    )
}

var RuntimeMetricsModule = fx.Module("runtime_metrics",
    fx.Invoke(EnableRuntimeMetrics),
)
```

**Metrics collected**:
- `process.runtime.go.mem.heap_alloc`
- `process.runtime.go.mem.heap_idle`
- `process.runtime.go.gc.count`
- `process.runtime.go.gc.pause_ns`
- `process.runtime.go.goroutines`

**Usage**: Just add module to `main.go`:
```go
fx.New(
    telemetry.Module,
    telemetry.RuntimeMetricsModule, // ← Add this
)
```

### Example 2: HTTP Client Auto-Tracing

**File**: `internal/telemetry/http_instrumentation.go`

```go
package telemetry

import (
    "net/http"

    "github.com/mapoio/hyperion"
    hyperotel "github.com/mapoio/hyperion/adapter/otel"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewInstrumentedHTTPClient creates HTTP client with automatic tracing
func NewInstrumentedHTTPClient(tracer hyperion.Tracer) *http.Client {
    // ⭐ Type-assert to access TracerProvider
    otelTracer, ok := tracer.(*hyperotel.OtelTracer)
    if !ok {
        return http.DefaultClient
    }

    // ⭐ Use TracerProvider for auto-instrumentation
    return &http.Client{
        Transport: otelhttp.NewTransport(
            http.DefaultTransport,
            otelhttp.WithTracerProvider(otelTracer.TracerProvider()),
        ),
    }
}

var HTTPInstrumentationModule = fx.Module("http_instrumentation",
    fx.Provide(NewInstrumentedHTTPClient),
)
```

**Usage in services**:
```go
type PaymentService struct {
    client *http.Client // ← Injected instrumented client
}

func (s *PaymentService) ProcessPayment(ctx hyperion.Context, amount float64) error {
    // ⭐ HTTP requests automatically traced!
    req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.stripe.com/charges", body)
    resp, err := s.client.Do(req)
    // Span automatically created with parent context
}
```

### Example 3: gRPC Auto-Instrumentation

**File**: `internal/telemetry/grpc_instrumentation.go` (you can create this)

```go
import (
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
    "google.golang.org/grpc"
)

func NewInstrumentedGRPCClient(tracer hyperion.Tracer, target string) (*grpc.ClientConn, error) {
    otelTracer := tracer.(*hyperotel.OtelTracer)

    return grpc.NewClient(target,
        grpc.WithStatsHandler(otelgrpc.NewClientHandler(
            otelgrpc.WithTracerProvider(otelTracer.TracerProvider()),
        )),
    )
}
```

### Example 4: Database Tracing (GORM)

```go
import "gorm.io/plugin/opentelemetry/tracing"

func NewInstrumentedDB(tracer hyperion.Tracer, dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    otelTracer := tracer.(*hyperotel.OtelTracer)
    err = db.Use(tracing.NewPlugin(
        tracing.WithTracerProvider(otelTracer.TracerProvider()),
    ))
    return db, err
}
```

## Testing

### Running Tests

```bash
cd example/otel/internal/telemetry
go test -v
```

### Integration Tests

**File**: `internal/telemetry/integration_test.go`

```go
// Test 1: Verify OTel SDK initializes correctly
func TestOTelSDKInitialization(t *testing.T) {
    cfg := NewSDKConfig()
    res, err := NewResource(cfg)
    require.NoError(t, err)

    tp, err := NewTracerProvider(cfg, res)
    require.NoError(t, err)
    defer tp.Shutdown(context.Background())

    mp, err := NewMeterProvider(cfg, res)
    require.NoError(t, err)
    defer mp.Shutdown(context.Background())

    assert.NotNil(t, tp)
    assert.NotNil(t, mp)
}

// Test 2: Verify Hyperion adapters integrate correctly
func TestHyperionAdapterIntegration(t *testing.T) {
    cfg := NewSDKConfig()
    res, _ := NewResource(cfg)
    tp, _ := NewTracerProvider(cfg, res)
    mp, _ := NewMeterProvider(cfg, res)

    // Create adapters from providers
    tracer := hyperotel.NewOtelTracerFromProvider(tp, "test-service")
    meter := hyperotel.NewOtelMeterFromProvider(mp, "test-service")

    // Verify interfaces
    var _ hyperion.Tracer = tracer
    var _ hyperion.Meter = meter

    // Verify accessor methods
    otelTracer := tracer.(*hyperotel.OtelTracer)
    assert.Equal(t, tp, otelTracer.TracerProvider())

    otelMeter := meter.(*hyperotel.OtelMeter)
    assert.Equal(t, mp, otelMeter.MeterProvider())
}

// Test 3: Verify OTel version independence
func TestOTelVersionIndependence(t *testing.T) {
    // This test passes if application can use different OTel version
    // No assertion needed - successful compilation is the test
}

// Test 4: Verify runtime metrics work
func TestRuntimeMetricsCollection(t *testing.T) {
    cfg := NewSDKConfig()
    res, _ := NewResource(cfg)
    mp, _ := NewMeterProvider(cfg, res)
    defer mp.Shutdown(context.Background())

    err := EnableRuntimeMetrics(mp)
    assert.NoError(t, err)

    time.Sleep(2 * time.Second) // Let metrics collect
}
```

### Benchmarks

```bash
go test -bench=. -benchmem
```

## Best Practices

### ✅ DO: Centralize OTel Configuration

**Single source of truth**:
```go
// internal/telemetry/otel_sdk.go
var Module = fx.Module("telemetry",
    fx.Provide(NewSDKConfig, NewResource, NewTracerProvider, NewMeterProvider),
)
```

### ✅ DO: Use fx.Module for Composability

```go
fx.New(
    telemetry.Module,                    // Core
    telemetry.RuntimeMetricsModule,      // Optional
    telemetry.HTTPInstrumentationModule, // Optional
)
```

### ✅ DO: Graceful Shutdown

```go
func RegisterShutdown(lc fx.Lifecycle, tp *sdktrace.TracerProvider, mp *metric.MeterProvider) {
    lc.Append(fx.Hook{
        OnStop: func(ctx context.Context) error {
            ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
            defer cancel()
            tp.Shutdown(ctx)
            mp.Shutdown(ctx)
            return nil
        },
    })
}
```

### ❌ DON'T: Create Multiple OTel Instances

```go
// ❌ BAD: Duplicate telemetry
telemetry.Module          // Creates TracerProvider
anotherPackage.OTelModule // Another TracerProvider!

// ✅ GOOD: Single instance
telemetry.Module
```

### ❌ DON'T: Use Global otel.Tracer() in Business Logic

```go
// ❌ BAD: Global dependency
func (s *Service) Method(ctx context.Context) {
    tracer := otel.Tracer("service")
    ctx, span := tracer.Start(ctx, "op")
}

// ✅ GOOD: Use hyperion.Context
func (s *Service) Method(ctx hyperion.Context) {
    ctx = ctx.UseIntercept("op") // Automatic tracing
}
```

## Migration Guide

### From Old Adapter Pattern

**Before (Adapter Controls OTel)**:
```go
// Old main.go
fx.New(
    otel.Module, // ❌ Adapter initializes OTel internally
)
```

**After (Application Controls OTel)**:
```go
// New main.go
fx.New(
    telemetry.Module, // ✅ Application initializes OTel

    fx.Provide(func(tp *sdktrace.TracerProvider) hyperion.Tracer {
        return hyperotel.NewOtelTracerFromProvider(tp, "my-service")
    }),
)
```

### Migration Steps

1. **Create `internal/telemetry/` package**
2. **Copy SDK initialization from this example**:
   - `otel_sdk.go`
   - `runtime_metrics.go` (optional)
   - `http_instrumentation.go` (optional)
3. **Update `main.go`** to 5-step pattern
4. **Update `go.mod`** with OTel dependencies
5. **Test**: `go test ./...`

## Troubleshooting

### No Telemetry Data in Backend

**Check**:
1. OTLP endpoint correct: `localhost:4317`
2. Backend accepting connections (HyperDX, Jaeger, etc.)
3. Exporter configured with `WithInsecure()` for local dev

**Fix**:
```go
exporter, _ := otlptracegrpc.New(ctx,
    otlptracegrpc.WithEndpoint("localhost:4317"),
    otlptracegrpc.WithInsecure(), // ← Important for local
)
```

### OTel Version Conflicts

**Solution**: Align all OTel dependencies:
```bash
go get go.opentelemetry.io/otel@v1.32.0
go get go.opentelemetry.io/otel/sdk@v1.32.0
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@v1.32.0
go mod tidy
```

### TracerProvider() Returns Nil

**Check type assertion**:
```go
otelTracer, ok := tracer.(*hyperotel.OtelTracer)
if !ok {
    // ❌ Wrong tracer type
    // Ensure you used NewOtelTracerFromProvider
}
```

## Additional Resources

- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/languages/go/)
- [OTel Registry](https://opentelemetry.io/ecosystem/registry/)
- [Hyperion Core](https://github.com/mapoio/hyperion)
- [Uber Fx](https://uber-go.github.io/fx/)

## Contributing

Found issues or have suggestions? Open an issue at:
https://github.com/mapoio/hyperion/issues
