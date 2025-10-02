package hyperion

import "go.uber.org/fx"

// CoreModule is the default Hyperion module with all no-op implementations.
// This is the RECOMMENDED module for most applications.
//
// Adapters will automatically override no-op implementations when provided:
//   - adapter/zap.Module overrides Logger
//   - adapter/otel.Module overrides Tracer
//   - adapter/gorm.Module overrides Database
//   - adapter/viper.Module overrides Config
//   - adapter/redis.Module overrides Cache
//
// Example usage:
//
//	fx.New(
//	    hyperion.CoreModule,       // Provides all no-op defaults
//	    zap.Module,                // Override Logger
//	    otel.Module,               // Override Tracer
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

		// Infrastructure (will be implemented later)
		// fx.Provide(NewUnitOfWork),
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
	// Infrastructure only (will be implemented later)
	// fx.Provide(NewUnitOfWork),
	),
)
