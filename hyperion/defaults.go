package hyperion

import (
	"go.uber.org/fx"
)

// DefaultLoggerModule provides a default no-op Logger implementation.
// Adapters use fx.Decorate to replace this with real implementations.
var DefaultLoggerModule = fx.Module("hyperion.default_logger",
	fx.Provide(
		fx.Annotate(
			NewNoOpLogger,
			fx.As(new(Logger)),
		),
	),
)

// DefaultTracerModule provides a default no-op Tracer implementation.
// Adapters use fx.Decorate to replace this with real implementations.
var DefaultTracerModule = fx.Module("hyperion.default_tracer",
	fx.Provide(
		fx.Annotate(
			NewNoOpTracer,
			fx.As(new(Tracer)),
		),
	),
)

// DefaultDatabaseModule provides a default no-op Database implementation.
// Adapters use fx.Decorate to replace this with real implementations.
var DefaultDatabaseModule = fx.Module("hyperion.default_database",
	fx.Provide(
		fx.Annotate(
			NewNoOpDatabase,
			fx.As(new(Database)),
		),
	),
)

// DefaultConfigModule provides a default no-op Config implementation.
// Adapters use fx.Decorate to replace this with real implementations.
var DefaultConfigModule = fx.Module("hyperion.default_config",
	fx.Provide(
		fx.Annotate(
			NewNoOpConfig,
			fx.As(new(Config)),
		),
	),
)

// DefaultCacheModule provides a default no-op Cache implementation.
// Adapters use fx.Decorate to replace this with real implementations.
var DefaultCacheModule = fx.Module("hyperion.default_cache",
	fx.Provide(
		fx.Annotate(
			NewNoOpCache,
			fx.As(new(Cache)),
		),
	),
)

// DefaultMeterModule provides a default no-op Meter implementation.
// Adapters use fx.Decorate to replace this with real implementations.
var DefaultMeterModule = fx.Module("hyperion.default_meter",
	fx.Provide(
		fx.Annotate(
			NewNoOpMeter,
			fx.As(new(Meter)),
		),
	),
)
