package viper

import (
	"go.uber.org/fx"

	"github.com/mapoio/hyperion"
)

// Module provides a Viper-based Config implementation.
// It overrides the default no-op Config when imported.
//
// Example usage:
//
//	fx.New(
//	    hyperion.CoreModule,
//	    viper.Module,  // Decorates hyperion.Config with Viper
//	    myapp.Module,
//	).Run()
var Module = fx.Module("hyperion.adapter.viper",
	fx.Decorate(
		fx.Annotate(
			NewProviderFromEnv,
			fx.As(new(hyperion.Config)),
			fx.As(new(hyperion.ConfigWatcher)),
		),
	),
)

// NewProviderFromEnv creates a viper provider using the CONFIG_PATH environment variable.
// Falls back to "configs/config.yaml" if CONFIG_PATH is not set.
func NewProviderFromEnv() (hyperion.ConfigWatcher, error) {
	// TODO: Read from environment variable or fx params
	// For now, use a sensible default
	return NewProvider("configs/config.yaml")
}
