// Package otel provides OpenTelemetry implementations of the Hyperion observability interfaces.
//
// This adapter implements the hyperion.Tracer and hyperion.Meter interfaces using the
// OpenTelemetry SDK, enabling production-ready distributed tracing and metrics collection
// with automatic exemplar support for trace-to-metric correlation.
//
// # Features
//
//   - Full OpenTelemetry SDK integration for traces and metrics
//   - Automatic exemplar support linking metrics to traces
//   - W3C Trace Context propagation for distributed tracing
//   - Multiple exporter support: Jaeger, Prometheus, OTLP
//   - Configuration-driven setup via hyperion.Config
//   - Graceful shutdown with trace/metric flushing
//   - fx module integration with lifecycle management
//
// # Usage
//
// Basic setup with fx:
//
//	import (
//	    "github.com/mapoio/hyperion"
//	    "github.com/mapoio/hyperion/adapter/otel"
//	    "go.uber.org/fx"
//	)
//
//	fx.New(
//	    hyperion.CoreModule,
//	    otel.Module,  // Provides both Tracer and Meter
//	    // ... your app modules
//	)
//
// Configuration example (config.yaml):
//
//	tracing:
//	  enabled: true
//	  service_name: my-service
//	  exporter: jaeger
//	  endpoint: localhost:14268
//	  sample_rate: 1.0
//
//	metrics:
//	  enabled: true
//	  service_name: my-service
//	  exporter: prometheus
//	  interval: 10s
//
// # Automatic Observability with Interceptors
//
// The recommended usage pattern is the 3-line interceptor approach:
//
//	func (s *UserService) GetUser(ctx hyperion.Context, id string) (_ *User, err error) {
//	    ctx, end := ctx.UseIntercept("UserService", "GetUser")
//	    defer end(&err)
//
//	    // Automatic tracing, logging, and metrics
//	    return s.userRepo.FindByID(ctx, id)
//	}
//
// # Manual Tracing
//
// For fine-grained control:
//
//	func (s *UserService) GetUser(ctx hyperion.Context, id string) (*User, error) {
//	    ctx, span := ctx.Tracer().Start(ctx, "UserService.GetUser")
//	    defer span.End()
//
//	    span.SetAttributes(hyperion.String("user.id", id))
//
//	    user, err := s.userRepo.FindByID(ctx, id)
//	    if err != nil {
//	        span.RecordError(err)
//	        return nil, err
//	    }
//	    return user, nil
//	}
//
// # Metrics with Exemplars
//
// Metrics automatically include exemplars when called within an active span:
//
//	counter := ctx.Meter().Counter("user.lookups")
//	counter.Add(ctx, 1, hyperion.String("method", "GetUser"))
//	// Exemplar automatically links to current trace
//
// # Exporters
//
// Supported trace exporters:
//   - Jaeger: Direct export to Jaeger collector
//   - OTLP: Export via OpenTelemetry Protocol (gRPC)
//
// Supported metrics exporters:
//   - Prometheus: Pull-based metrics endpoint
//   - OTLP: Push-based metrics via OpenTelemetry Protocol (gRPC)
//
// # Performance
//
// The OpenTelemetry adapter is designed for production use with minimal overhead:
//   - Batched span export to reduce network calls
//   - Configurable sampling to control trace volume
//   - Async metric collection with periodic export
//   - Benchmarks verify <5% overhead vs NoOp implementations
package otel
