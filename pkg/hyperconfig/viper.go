package hyperconfig

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// callbackEntry holds a callback function with its unique ID.
type callbackEntry struct {
	id       uint64
	callback func(ChangeEvent)
}

// ViperProvider is a Provider implementation based on spf13/viper.
// It supports multiple configuration formats (YAML, JSON, TOML) and
// automatic environment variable override.
type ViperProvider struct {
	v          *viper.Viper
	mu         sync.RWMutex
	callbacks  map[uint64]func(ChangeEvent)
	nextCallID uint64
}

// NewViperProvider creates a new ViperProvider from the given configuration file path.
// It automatically detects the file format based on the file extension.
//
// Supported formats: .yaml, .yml, .json, .toml
//
// Environment variables are automatically mapped to configuration keys.
// For example, APP_DATABASE_HOST maps to the "database.host" config key.
//
// Returns an error if the configuration file cannot be read or parsed.
func NewViperProvider(configPath string) (*ViperProvider, error) {
	v := viper.New()

	// Set config file path
	v.SetConfigFile(configPath)

	// Enable automatic environment variable override
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return &ViperProvider{
		v:          v,
		callbacks:  make(map[uint64]func(ChangeEvent)),
		nextCallID: 0,
	}, nil
}

// NewViperProviderWithViper creates a new ViperProvider from an existing viper instance.
// This is useful for testing or when you need more control over viper configuration.
func NewViperProviderWithViper(v *viper.Viper) *ViperProvider {
	return &ViperProvider{
		v:          v,
		callbacks:  make(map[uint64]func(ChangeEvent)),
		nextCallID: 0,
	}
}

// Unmarshal unmarshals the configuration at the given key into the provided struct.
func (p *ViperProvider) Unmarshal(key string, rawVal any) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if key == "" {
		// Unmarshal entire config
		return p.v.Unmarshal(rawVal)
	}

	// Unmarshal specific key
	return p.v.UnmarshalKey(key, rawVal)
}

// Get returns the value for the given key as an interface{}.
func (p *ViperProvider) Get(key string) any {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.Get(key)
}

// GetString returns the value for the given key as a string.
func (p *ViperProvider) GetString(key string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetString(key)
}

// GetInt returns the value for the given key as an int.
func (p *ViperProvider) GetInt(key string) int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetInt(key)
}

// GetInt64 returns the value for the given key as an int64.
func (p *ViperProvider) GetInt64(key string) int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetInt64(key)
}

// GetBool returns the value for the given key as a bool.
func (p *ViperProvider) GetBool(key string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetBool(key)
}

// GetFloat64 returns the value for the given key as a float64.
func (p *ViperProvider) GetFloat64(key string) float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetFloat64(key)
}

// GetStringSlice returns the value for the given key as a string slice.
func (p *ViperProvider) GetStringSlice(key string) []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetStringSlice(key)
}

// IsSet checks if the key is set in the configuration.
func (p *ViperProvider) IsSet(key string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.IsSet(key)
}

// AllKeys returns all keys in the configuration.
func (p *ViperProvider) AllKeys() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.AllKeys()
}

// Watch starts watching for configuration file changes and triggers callbacks.
// It uses fsnotify internally via viper's WatchConfig mechanism.
//
// The callback is invoked whenever the configuration file is modified.
// Note: The ChangeEvent.Key will contain the filename and Value will be nil for
// file-based watches. Use the Provider methods to read the updated configuration values.
//
// Multiple callbacks can be registered by calling Watch multiple times.
// Each callback is assigned a unique ID to ensure safe removal even in
// concurrent scenarios.
//
// Returns a stop function to cancel watching and an error if watching cannot be started.
func (p *ViperProvider) Watch(callback func(event ChangeEvent)) (stop func(), err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Generate unique ID for this callback using atomic increment
	callbackID := atomic.AddUint64(&p.nextCallID, 1)

	// Add callback to map with unique ID
	p.callbacks[callbackID] = callback

	// Only set up watching once (when first callback is registered)
	if len(p.callbacks) == 1 {
		p.v.WatchConfig()
		p.v.OnConfigChange(func(e fsnotify.Event) {
			p.mu.RLock()
			// Create a snapshot of callbacks to avoid holding lock during execution
			callbacks := make([]func(ChangeEvent), 0, len(p.callbacks))
			for _, cb := range p.callbacks {
				callbacks = append(callbacks, cb)
			}
			p.mu.RUnlock()

			// Trigger all registered callbacks
			event := ChangeEvent{
				Key:   filepath.Base(e.Name),
				Value: nil, // File-level change, not specific key
			}

			for _, cb := range callbacks {
				cb(event)
			}
		})
	}

	// Return stop function that removes this specific callback by ID
	return func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		delete(p.callbacks, callbackID)
	}, nil
}
