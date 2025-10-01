package hyperconfig

// Provider defines the interface for configuration providers.
// It supports reading configuration from multiple sources (files, environment variables)
// and provides type-safe accessors for configuration values.
type Provider interface {
	// Unmarshal unmarshals the value at the given key into the provided struct.
	// Returns an error if the key doesn't exist or unmarshalling fails.
	Unmarshal(key string, rawVal any) error

	// Get returns the value for the given key as an interface{}.
	// Returns nil if the key doesn't exist.
	Get(key string) any

	// GetString returns the value for the given key as a string.
	// Returns empty string if the key doesn't exist.
	GetString(key string) string

	// GetInt returns the value for the given key as an int.
	// Returns 0 if the key doesn't exist.
	GetInt(key string) int

	// GetInt64 returns the value for the given key as an int64.
	// Returns 0 if the key doesn't exist.
	GetInt64(key string) int64

	// GetBool returns the value for the given key as a bool.
	// Returns false if the key doesn't exist.
	GetBool(key string) bool

	// GetFloat64 returns the value for the given key as a float64.
	// Returns 0.0 if the key doesn't exist.
	GetFloat64(key string) float64

	// GetStringSlice returns the value for the given key as a string slice.
	// Returns nil if the key doesn't exist.
	GetStringSlice(key string) []string

	// IsSet checks if the key is set in the configuration.
	IsSet(key string) bool

	// AllKeys returns all keys in the configuration.
	AllKeys() []string
}

// Watcher defines the interface for watching configuration changes.
// Implementations should support hot reload by monitoring configuration sources
// and triggering callbacks when changes are detected.
type Watcher interface {
	// Watch starts watching for configuration changes and calls the provided
	// callback function when changes are detected.
	//
	// The callback is invoked with a ChangeEvent containing the changed key
	// and new value. Multiple callbacks can be registered by calling Watch
	// multiple times.
	//
	// Returns a stop function that can be called to cancel watching, and an
	// error if watching cannot be started.
	//
	// Example:
	//   stop, err := watcher.Watch(func(event ChangeEvent) {
	//       log.Printf("Config changed: %s = %v", event.Key, event.Value)
	//   })
	//   if err != nil {
	//       return err
	//   }
	//   defer stop() // Stop watching when done
	Watch(callback func(event ChangeEvent)) (stop func(), err error)
}

// ChangeEvent represents a configuration change event.
//
// Important: The semantics of Key and Value depend on the Provider implementation:
//
// For file-based providers (like ViperProvider):
//   - Key contains the configuration filename (e.g., "config.yaml")
//   - Value is always nil
//   - Applications should re-read configuration using Provider methods
//
// For future key-level providers:
//   - Key would contain the specific config key (e.g., "database.host")
//   - Value would contain the new value
//
// Example usage with file-based provider:
//
//	provider.Watch(func(event ChangeEvent) {
//	    log.Printf("Config file changed: %s", event.Key)
//	    // Re-read the specific config you care about
//	    newLogLevel := provider.GetString("log.level")
//	    logger.SetLevel(parseLevel(newLogLevel))
//	})
type ChangeEvent struct {
	// Value is the new value after the change. For file-based watching, this is nil.
	// For key-level watching (future enhancement), this would be the actual new value.
	Value any

	// Key identifies what changed. For file-based watching, this is the filename.
	// For key-level watching (future enhancement), this would be the config key path.
	Key string
}
