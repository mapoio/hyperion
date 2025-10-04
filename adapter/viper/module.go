package viper

import (
	"go.uber.org/fx"

	"github.com/mapoio/hyperion"
)

// Module provides Viper-based Config implementation.
//
// Usage:
//
//	fx.New(
//	    hyperion.CoreModule,
//	    viper.Module,  // Provides Config
//	    myapp.Module,
//	).Run()
var Module = fx.Module("hyperion.adapter.viper",
	fx.Provide(
		fx.Annotate(
			NewViperProvider,
			fx.As(new(hyperion.Config)),
			fx.As(new(hyperion.ConfigWatcher)),
		),
	),
)

// NewViperProvider creates a Viper config provider.
func NewViperProvider() (hyperion.ConfigWatcher, error) {
	configPath := "configs/config.yaml"
	provider, err := NewProvider(configPath)
	if err != nil {
		return nil, err
	}
	return provider, nil
}