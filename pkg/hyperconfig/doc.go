// Package hyperconfig provides configuration management for Hyperion applications.
//
// It wraps spf13/viper with support for multiple configuration sources and hot reload.
// Supports YAML, JSON, TOML formats and environment variable overrides.
//
// # Features
//
//   - Multi-format support (YAML, JSON, TOML)
//   - Automatic environment variable override
//   - Hot reload with file watching
//   - Type-safe configuration unmarshalling
//   - Nested configuration key access
//
// # Basic Usage
//
//	provider, err := hyperconfig.NewViperProvider("configs/config.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Read configuration values
//	logLevel := provider.GetString("log.level")
//	port := provider.GetInt("server.port")
//
//	// Unmarshal into struct
//	var dbConfig DatabaseConfig
//	if err := provider.Unmarshal("database", &dbConfig); err != nil {
//	    log.Fatal(err)
//	}
//
// # Hot Reload
//
//	stop, err := provider.Watch(func(event hyperconfig.ChangeEvent) {
//	    log.Printf("Configuration changed: %s", event.Key)
//	    // Reload configuration as needed
//	    newLevel := provider.GetString("log.level")
//	    logger.SetLevel(parseLevel(newLevel))
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer stop() // Stop watching when done
//
// # Environment Variables
//
// Environment variables automatically override configuration file values.
// Use underscores in place of dots for nested keys:
//
//	export APP_LOG_LEVEL=debug        # Overrides log.level
//	export APP_DATABASE_HOST=postgres # Overrides database.host
//
// # fx Integration
//
//	fx.New(
//	    fx.Provide(func() (*hyperconfig.ViperProvider, error) {
//	        return hyperconfig.NewViperProvider("configs/config.yaml")
//	    }),
//	    hyperconfig.Module,
//	    fx.Invoke(func(cfg hyperconfig.Provider) {
//	        // Use configuration
//	    }),
//	)
package hyperconfig
