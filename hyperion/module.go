package hyperion

import "go.uber.org/fx"

// CoreModule provides the minimal core infrastructure.
// It ONLY includes ContextFactory and Interceptor infrastructure.
//
// You MUST provide implementations for all interfaces via adapters:
//   - Config: viper.Module, etc.
//   - Logger: zap.Module, etc.
//   - Tracer: hyperotel.Module, etc.
//   - Meter: hyperotel.Module, etc.
//   - Database: gorm.Module, etc.
//   - Cache: redis.Module, etc.
//   - UnitOfWork: gorm.Module, etc.
//
// Usage:
//
//	fx.New(
//	    hyperion.CoreModule,   // Core infrastructure only
//	    viper.Module,          // Provide Config
//	    zap.Module,            // Provide Logger
//	    hyperotel.Module,      // Provide Tracer and Meter
//	    gorm.Module,           // Provide Database and UnitOfWork
//	    redis.Module,          // Provide Cache
//	    myapp.Module,
//	).Run()
var CoreModule = fx.Module("hyperion.core",
	fx.Options(
		ContextModule,
		InterceptorsModule,
	),
)

// ContextModule provides ContextFactory and InterceptorRegistry for dependency injection.
// This module is automatically included in CoreModule.
//
// The ContextFactory uses the InterceptorRegistry to dynamically fetch interceptors
// at context creation time, avoiding timing and lazy loading issues.
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
		// Provide InterceptorRegistry singleton
		fx.Annotate(
			NewInterceptorRegistry,
			fx.As(new(InterceptorRegistry)),
		),
		// Provide ContextFactory with registry
		func(params struct {
			fx.In
			Logger   Logger
			Tracer   Tracer
			DB       Database
			Meter    Meter
			Registry InterceptorRegistry
		}) ContextFactory {
			return NewContextFactory(
				params.Logger,
				params.Tracer,
				params.DB,
				params.Meter,
				WithRegistry(params.Registry),
			)
		},
	),
)
