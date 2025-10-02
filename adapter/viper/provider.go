package viper

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/mapoio/hyperion"
)

// Provider is a hyperion.ConfigWatcher implementation based on spf13/viper.
// It supports multiple configuration formats (YAML, JSON, TOML) and
// automatic environment variable override with hot reload capabilities.
type Provider struct {
	v          *viper.Viper                          // Viper instance
	watcher    *fsnotify.Watcher                     // File system watcher
	callbacks  map[uint64]func(hyperion.ChangeEvent) // Registered callbacks
	watchDone  chan struct{}                         // Signal to stop watching
	configPath string                                // Original config file path
	mu         sync.RWMutex                          // Protects callbacks and viper access
	nextCallID uint64                                // Atomic counter for callback IDs
}

// NewProvider creates a new viper-based config provider from the given configuration file path.
// It automatically detects the file format based on the file extension.
//
// Supported formats: .yaml, .yml, .json, .toml
//
// Environment variables are automatically mapped to configuration keys.
// For example, APP_DATABASE_HOST maps to the "database.host" config key.
//
// Returns an error if the configuration file cannot be read or parsed.
func NewProvider(configPath string) (hyperion.ConfigWatcher, error) {
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

	return &Provider{
		v:          v,
		callbacks:  make(map[uint64]func(hyperion.ChangeEvent)),
		nextCallID: 0,
		configPath: configPath,
	}, nil
}

// Unmarshal unmarshals the configuration at the given key into the provided struct.
func (p *Provider) Unmarshal(key string, rawVal any) error {
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
func (p *Provider) Get(key string) any {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.Get(key)
}

// GetString returns the value for the given key as a string.
func (p *Provider) GetString(key string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetString(key)
}

// GetInt returns the value for the given key as an int.
func (p *Provider) GetInt(key string) int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetInt(key)
}

// GetInt64 returns the value for the given key as an int64.
func (p *Provider) GetInt64(key string) int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetInt64(key)
}

// GetBool returns the value for the given key as a bool.
func (p *Provider) GetBool(key string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetBool(key)
}

// GetFloat64 returns the value for the given key as a float64.
func (p *Provider) GetFloat64(key string) float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetFloat64(key)
}

// GetStringSlice returns the value for the given key as a string slice.
func (p *Provider) GetStringSlice(key string) []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.GetStringSlice(key)
}

// IsSet checks if the key is set in the configuration.
func (p *Provider) IsSet(key string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.IsSet(key)
}

// AllKeys returns all keys in the configuration.
func (p *Provider) AllKeys() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.v.AllKeys()
}

// Watch starts watching for configuration file changes and triggers callbacks.
// It uses fsnotify to watch the configuration file directly.
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
func (p *Provider) Watch(callback func(event hyperion.ChangeEvent)) (stop func(), err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Generate unique ID for this callback using atomic increment
	callbackID := atomic.AddUint64(&p.nextCallID, 1)

	// Add callback to map with unique ID
	p.callbacks[callbackID] = callback

	// Only set up watching once (when first callback is registered)
	if len(p.callbacks) == 1 {
		// Create a new file system watcher
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			delete(p.callbacks, callbackID)
			return nil, fmt.Errorf("failed to create watcher: %w", err)
		}

		// Add the config file to the watcher
		if err := watcher.Add(p.configPath); err != nil {
			if closeErr := watcher.Close(); closeErr != nil {
				// Log close error but prioritize returning the original error
				fmt.Printf("failed to close watcher: %v\n", closeErr)
			}
			delete(p.callbacks, callbackID)
			return nil, fmt.Errorf("failed to watch config file: %w", err)
		}

		p.watcher = watcher
		p.watchDone = make(chan struct{})

		// Start the watch loop in a separate goroutine
		go p.watchLoop()
	}

	// Return stop function that removes this specific callback by ID
	return func() {
		p.mu.Lock()
		defer p.mu.Unlock()

		delete(p.callbacks, callbackID)

		// If this was the last callback, stop watching
		if len(p.callbacks) == 0 && p.watcher != nil {
			close(p.watchDone)
			if err := p.watcher.Close(); err != nil {
				// Log error but continue cleanup (non-critical)
				fmt.Printf("failed to close watcher: %v\n", err)
			}
			p.watcher = nil
			p.watchDone = nil
		}
	}, nil
}

// watchLoop handles file system events and reloads configuration.
// It runs in a separate goroutine and terminates when watchDone is closed.
func (p *Provider) watchLoop() {
	// Get channels with lock to avoid races during shutdown
	p.mu.RLock()
	// Guard against nil watcher in case stop() is called before this goroutine starts
	if p.watcher == nil || p.watchDone == nil {
		p.mu.RUnlock()
		return
	}
	eventsC := p.watcher.Events
	errorsC := p.watcher.Errors
	doneC := p.watchDone
	p.mu.RUnlock()

	for {
		select {
		case event, ok := <-eventsC:
			if !ok {
				return
			}
			p.handleFileEvent(event)

		case err, ok := <-errorsC:
			if !ok {
				return
			}
			// Log error but continue watching (in production, use proper logger)
			fmt.Printf("watcher error: %v\n", err)

		case <-doneC:
			return
		}
	}
}

// handleFileEvent processes a file system event and triggers callbacks.
// Handles atomic writes (rename/remove) used by editors and k8s ConfigMaps.
func (p *Provider) handleFileEvent(event fsnotify.Event) {
	// Handle atomic writes: many editors and k8s ConfigMap reloads use
	// atomic writes (write to temp file, then rename), which triggers
	// Rename or Remove events instead of Write.
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove) == 0 {
		return
	}

	// For rename/remove events, re-add the watch to handle atomic writes
	// The watch stays on the old inode after rename, so we need to watch the new file
	if event.Op&(fsnotify.Rename|fsnotify.Remove) != 0 {
		p.mu.Lock()
		if p.watcher != nil {
			// Remove the old watch (if it exists) - ignore error as path may not exist
			if err := p.watcher.Remove(p.configPath); err != nil {
				// Log but continue - the path may have already been removed
				fmt.Printf("note: failed to remove old watch (expected after rename): %v\n", err)
			}
			// Re-add watch to the config path (which now points to the new file)
			if err := p.watcher.Add(p.configPath); err != nil {
				fmt.Printf("failed to re-add watch after rename: %v\n", err)
				p.mu.Unlock()
				return
			}
		}
		p.mu.Unlock()
	}

	// Reload configuration with write lock
	p.mu.Lock()
	if err := p.v.ReadInConfig(); err != nil {
		fmt.Printf("failed to reload config: %v\n", err)
		p.mu.Unlock()
		return
	}

	// Create change event
	changeEvent := hyperion.ChangeEvent{
		Key:   filepath.Base(event.Name),
		Value: nil,
	}

	// Snapshot callbacks while holding lock
	callbacks := make([]func(hyperion.ChangeEvent), 0, len(p.callbacks))
	for _, cb := range p.callbacks {
		callbacks = append(callbacks, cb)
	}
	p.mu.Unlock()

	// Execute callbacks outside lock
	for _, cb := range callbacks {
		cb(changeEvent)
	}
}
