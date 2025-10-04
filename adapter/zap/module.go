package zap

import (
	"go.uber.org/fx"

	"github.com/mapoio/hyperion"
)

// Module provides Zap-based Logger implementation.
//
// Usage:
//
//	fx.New(
//	    hyperion.CoreModule,
//	    viper.Module,  // Provides Config (optional for Zap)
//	    zap.Module,    // Provides Logger
//	    myapp.Module,
//	).Run()
var Module = fx.Module("hyperion.adapter.zap",
	fx.Provide(
		fx.Annotate(
			NewZapProvider,
			fx.As(new(hyperion.Logger)),
		),
	),
)

// NewZapProvider creates a Zap logger.
func NewZapProvider(cfg hyperion.Config) (hyperion.Logger, error) {
	return NewZapLogger(cfg)
}
