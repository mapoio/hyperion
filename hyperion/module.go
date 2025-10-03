package hyperion

import "go.uber.org/fx"

// CoreModule is the default Hyperion module with all no-op implementations.
// This is the RECOMMENDED module for most applications.
//
// CoreModule includes:
//   - All no-op default implementations (Logger, Tracer, Database, Config, Cache, Meter)
//   - ContextFactory with interceptor infrastructure
//   - InterceptorsModule (base infrastructure, no interceptors registered)
//
// Adapters will automatically override no-op implementations when provided:
//   - adapter/zap.Module overrides Logger
//   - adapter/otel.Module overrides Tracer and Meter
//   - adapter/gorm.Module overrides Database
//   - adapter/viper.Module overrides Config
//   - adapter/redis.Module overrides Cache
//
// To enable built-in interceptors, add them separately:
//
//	fx.New(
//	    hyperion.CoreModule,                  // Core infrastructure
//	    hyperion.TracingInterceptorModule,    // Optional: enable tracing
//	    hyperion.LoggingInterceptorModule,    // Optional: enable logging
//	    zap.Module,                           // Override Logger
//	    otel.Module,                          // Override Tracer and Meter
//	    myapp.Module,
//	).Run()
var CoreModule = fx.Module("hyperion.core",
	fx.Options(
		// Default implementations (nil + Decorate pattern)
		DefaultLoggerModule,
		DefaultTracerModule,
		DefaultDatabaseModule,
		DefaultConfigModule,
		DefaultCacheModule,
		DefaultMeterModule,

		// Context infrastructure with interceptor support
		ContextModule,
		InterceptorsModule, // Base infrastructure (no interceptors registered)
	),
)

// CoreWithoutDefaultsModule is the minimal Hyperion module without any default implementations.
// Use this ONLY when you want to provide ALL adapters explicitly.
//
// If any adapter is missing, fx will fail with a dependency error at startup.
// This is useful for production environments where you want to enforce
// that all dependencies are explicitly configured.
//
// Example usage:
//
//	fx.New(
//	    hyperion.CoreWithoutDefaultsModule,  // No defaults
//	    zap.Module,                          // MUST provide
//	    otel.Module,                         // MUST provide
//	    gorm.Module,                         // MUST provide
//	    viper.Module,                        // MUST provide
//	    redis.Module,                        // MUST provide
//	    myapp.Module,
//	).Run()
var CoreWithoutDefaultsModule = fx.Module("hyperion.core.minimal",
	fx.Options(
		// Context infrastructure with interceptor support
		ContextModule,
		InterceptorsModule, // Base infrastructure (no interceptors registered)
	),
)

// ContextModule provides ContextFactory for dependency injection.
// This module is automatically included in CoreModule.
//
// The ContextFactory will automatically inject interceptors from the
// "hyperion.interceptors" fx group if any are registered.
//
// Example usage (standalone):
//
//	fx.New(
//	    hyperion.ContextModule,
//	    zap.Module,
//	    gorm.Module,
//	    otel.Module,
//	    myapp.Module,
//	).Run()
var ContextModule = fx.Module("hyperion.context",
	fx.Provide(
		// Provide ContextFactory with interceptors from group
		// If no interceptors are registered, the slice will be empty
		func(params struct {
			fx.In
			Logger       Logger
			Tracer       Tracer
			DB           Database
			Meter        Meter
			Interceptors []Interceptor `group:"hyperion.interceptors"`
		}) ContextFactory {
			return NewContextFactory(
				params.Logger,
				params.Tracer,
				params.DB,
				params.Meter,
				WithInterceptors(params.Interceptors...),
			)
		},
	),
)
