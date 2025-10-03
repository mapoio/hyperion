package hyperion

import (
	"fmt"

	"go.uber.org/fx"
)

// DefaultLoggerModule provides a default no-op Logger implementation.
var DefaultLoggerModule = fx.Module("hyperion.default_logger",
	fx.Provide(func() Logger {
		fmt.Println("[Hyperion] Using no-op Logger")
		return NewNoOpLogger()
	}),
)

// DefaultTracerModule provides a default no-op Tracer implementation.
var DefaultTracerModule = fx.Module("hyperion.default_tracer",
	fx.Provide(func() Tracer {
		fmt.Println("[Hyperion] Using no-op Tracer")
		return NewNoOpTracer()
	}),
)

// DefaultDatabaseModule provides a default no-op Database implementation.
var DefaultDatabaseModule = fx.Module("hyperion.default_database",
	fx.Provide(func() Database {
		fmt.Println("[Hyperion] Using no-op Database")
		return NewNoOpDatabase()
	}),
)

// DefaultConfigModule provides a default no-op Config implementation.
var DefaultConfigModule = fx.Module("hyperion.default_config",
	fx.Provide(func() Config {
		fmt.Println("[Hyperion] Using no-op Config")
		return NewNoOpConfig()
	}),
)

// DefaultCacheModule provides a default no-op Cache implementation.
var DefaultCacheModule = fx.Module("hyperion.default_cache",
	fx.Provide(func() Cache {
		fmt.Println("[Hyperion] Using no-op Cache")
		return NewNoOpCache()
	}),
)

// DefaultMeterModule provides a default no-op Meter implementation.
var DefaultMeterModule = fx.Module("hyperion.default_meter",
	fx.Provide(NewNoOpMeter),
)
