package hyperion

// Config is the configuration provider abstraction.
// Implementations should support multiple configuration formats and sources.
type Config interface {
	// Unmarshal unmarshals the configuration at the given key into rawVal.
	// If key is empty, unmarshals the entire configuration.
	Unmarshal(key string, rawVal any) error

	// Get returns the value for the given key.
	Get(key string) any

	// GetString returns the value for the given key as a string.
	GetString(key string) string

	// GetInt returns the value for the given key as an int.
	GetInt(key string) int

	// GetInt64 returns the value for the given key as an int64.
	GetInt64(key string) int64

	// GetBool returns the value for the given key as a bool.
	GetBool(key string) bool

	// GetFloat64 returns the value for the given key as a float64.
	GetFloat64(key string) float64

	// GetStringSlice returns the value for the given key as a string slice.
	GetStringSlice(key string) []string

	// IsSet checks if the key is set in the configuration.
	IsSet(key string) bool

	// AllKeys returns all keys in the configuration.
	AllKeys() []string
}

// ConfigWatcher supports configuration hot reload.
// Implementations should notify callbacks when configuration changes.
type ConfigWatcher interface {
	Config

	// Watch registers a callback to be invoked when configuration changes.
	// Returns a stop function to cancel watching.
	//
	// Multiple callbacks can be registered by calling Watch multiple times.
	// Callbacks are invoked in the order they were registered.
	Watch(callback func(event ChangeEvent)) (stop func(), err error)
}

// ChangeEvent represents a configuration change event.
type ChangeEvent struct {
	// Value is the new value, or nil for file-based events.
	Value any

	// Key is the configuration key that changed.
	// For file-based watchers, this may be the filename.
	Key string
}
