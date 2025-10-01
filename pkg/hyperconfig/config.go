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
// It contains information about which configuration key changed and its new value.
type ChangeEvent struct {
	// Key is the configuration key that changed.
	// For nested keys, this uses dot notation (e.g., "database.host").
	Key string

	// Value is the new value after the change.
	// The actual type depends on the configuration value.
	Value any
}
