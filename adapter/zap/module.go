package zap

import (
	"go.uber.org/fx"

	"github.com/mapoio/hyperion"
)

// Module provides Zap logger as hyperion.Logger via fx dependency injection.
//
// Usage:
//
//	app := fx.New(
//	    hyperion.CoreModule,
//	    viper.Module,  // Provides Config
//	    zap.Module,    // Decorates Logger with Zap
//	    fx.Invoke(func(logger hyperion.Logger) {
//	        logger.Info("application started", "version", "1.0.0")
//	    }),
//	)
var Module = fx.Module("hyperion.adapter.zap",
	fx.Decorate(
		fx.Annotate(
			NewZapLogger,
			fx.As(new(hyperion.Logger)),
		),
	),
)
