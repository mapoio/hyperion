package hyperconfig

import (
	"go.uber.org/fx"
)

// Module provides the hyperconfig component to the fx dependency injection container.
// It registers the ViperProvider as the implementation for both Provider and Watcher interfaces.
//
// The configuration file path must be provided by the application through fx parameters
// or by creating the provider manually before passing it to fx.
//
// Example usage:
//
//	fx.New(
//	    hyperconfig.Module,
//	    fx.Provide(func() (*hyperconfig.ViperProvider, error) {
//	        return hyperconfig.NewViperProvider("configs/config.yaml")
//	    }),
//	    // ... other modules
//	)
var Module = fx.Module("hyperconfig",
	// Provide Provider interface
	fx.Provide(
		fx.Annotate(
			func(vp *ViperProvider) Provider {
				return vp
			},
			fx.ResultTags(`name:"hyperconfig.provider"`),
		),
	),
	// Provide Watcher interface
	fx.Provide(
		fx.Annotate(
			func(vp *ViperProvider) Watcher {
				return vp
			},
			fx.ResultTags(`name:"hyperconfig.watcher"`),
		),
	),
)
