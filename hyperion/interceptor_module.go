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
// Registers TracingInterceptor to the InterceptorRegistry via fx.Invoke.
// This ensures the interceptor is eagerly instantiated and registered.
var TracingInterceptorModule = fx.Module("hyperion.interceptors.tracing",
	fx.Provide(NewTracingInterceptor),
	fx.Invoke(func(registry InterceptorRegistry, interceptor *TracingInterceptor) {
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
// Registers LoggingInterceptor to the InterceptorRegistry via fx.Invoke.
// This ensures the interceptor is eagerly instantiated and registered.
var LoggingInterceptorModule = fx.Module("hyperion.interceptors.logging",
	fx.Provide(NewLoggingInterceptor),
	fx.Invoke(func(registry InterceptorRegistry, interceptor *LoggingInterceptor) {
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
