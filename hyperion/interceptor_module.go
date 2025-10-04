package hyperion

import "go.uber.org/fx"

// InterceptorsModule is the base module that is included in CoreModule.
// It provides the infrastructure for interceptors but does NOT register
// any interceptors by default.
//
// This module is automatically included in CoreModule and should not be
// imported separately unless you are using CoreWithoutDefaultsModule.
var InterceptorsModule = fx.Module("hyperion.interceptors.base")

// TracingInterceptorModule provides OpenTelemetry tracing interceptor.
// This module is OPTIONAL and must be explicitly imported to enable tracing.
//
// To enable tracing interceptor in your application:
//
//	fx.New(
//	    hyperion.CoreModule,
//	    hyperion.TracingInterceptorModule, // Enable OpenTelemetry tracing
//	    // ... other modules
//	)
//
// The TracingInterceptor will:
//   - Create a span for each intercepted method
//   - Record errors to the span
//   - Execute with order 100 (outer-most)
//
// Implementation Note:
// Uses fx.Invoke to create and register TracingInterceptor AFTER all Provide/Decorate.
// This ensures Tracer is available when TracingInterceptor is constructed.
var TracingInterceptorModule = fx.Module("hyperion.interceptors.tracing",
	fx.Invoke(func(registry InterceptorRegistry, tracer Tracer) {
		interceptor := NewTracingInterceptor(tracer)
		registry.Register(interceptor)
	}),
)

// LoggingInterceptorModule provides structured logging interceptor.
// This module is OPTIONAL and must be explicitly imported to enable logging.
//
// To enable logging interceptor in your application:
//
//	fx.New(
//	    hyperion.CoreModule,
//	    hyperion.LoggingInterceptorModule, // Enable structured logging
//	    // ... other modules
//	)
//
// The LoggingInterceptor will:
//   - Log method start (debug level)
//   - Log method completion with duration (debug level)
//   - Log method failure with error (error level)
//   - Execute with order 200 (after tracing)
//
// Implementation Note:
// Uses fx.Invoke to create and register LoggingInterceptor AFTER all Provide/Decorate.
// This ensures Logger is available when LoggingInterceptor is constructed.
var LoggingInterceptorModule = fx.Module("hyperion.interceptors.logging",
	fx.Invoke(func(registry InterceptorRegistry, logger Logger) {
		interceptor := NewLoggingInterceptor(logger)
		registry.Register(interceptor)
	}),
)

// AllInterceptorsModule is a convenience module that enables both
// tracing and logging interceptors.
//
// This is equivalent to:
//
//	fx.Options(
//	    hyperion.TracingInterceptorModule,
//	    hyperion.LoggingInterceptorModule,
//	)
//
// Example usage:
//
//	fx.New(
//	    hyperion.CoreModule,
//	    hyperion.AllInterceptorsModule, // Enable all built-in interceptors
//	    // ... other modules
//	)
var AllInterceptorsModule = fx.Module("hyperion.interceptors.all",
	fx.Options(
		TracingInterceptorModule,
		LoggingInterceptorModule,
	),
)
