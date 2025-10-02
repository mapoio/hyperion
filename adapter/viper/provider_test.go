package viper_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mapoio/hyperion"
	viperadapter "github.com/mapoio/hyperion/adapter/viper"
)

// TestNewProvider tests creating a new viper provider
func TestNewProvider(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
app:
  name: test-app
  version: 1.0.0
database:
  host: localhost
  port: 5432
  name: testdb
server:
  port: 8080
  timeout: 30
features:
  enabled: true
  rate: 0.95
  tags:
    - alpha
    - beta
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Test successful provider creation
	provider, err := viperadapter.NewProvider(configPath)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}
	if provider == nil {
		t.Fatal("Provider should not be nil")
	}
}

// TestNewProviderInvalidPath tests error handling for invalid config path
func TestNewProviderInvalidPath(t *testing.T) {
	_, err := viperadapter.NewProvider("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Expected error for invalid config path")
	}
}

// TestProviderGet tests all getter methods
func TestProviderGet(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
app:
  name: test-app
  version: 1.0.0
database:
  host: localhost
  port: 5432
  name: testdb
server:
  port: 8080
  timeout: 30
features:
  enabled: true
  rate: 0.95
  tags:
    - alpha
    - beta
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	provider, err := viperadapter.NewProvider(configPath)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	// Test Get
	if val := provider.Get("app.name"); val != "test-app" {
		t.Errorf("Get(app.name) = %v, want test-app", val)
	}

	// Test GetString
	if val := provider.GetString("database.host"); val != "localhost" {
		t.Errorf("GetString(database.host) = %v, want localhost", val)
	}

	// Test GetInt
	if val := provider.GetInt("server.port"); val != 8080 {
		t.Errorf("GetInt(server.port) = %v, want 8080", val)
	}

	// Test GetInt64
	if val := provider.GetInt64("database.port"); val != 5432 {
		t.Errorf("GetInt64(database.port) = %v, want 5432", val)
	}

	// Test GetBool
	if val := provider.GetBool("features.enabled"); !val {
		t.Error("GetBool(features.enabled) = false, want true")
	}

	// Test GetFloat64
	if val := provider.GetFloat64("features.rate"); val != 0.95 {
		t.Errorf("GetFloat64(features.rate) = %v, want 0.95", val)
	}

	// Test GetStringSlice
	tags := provider.GetStringSlice("features.tags")
	if len(tags) != 2 || tags[0] != "alpha" || tags[1] != "beta" {
		t.Errorf("GetStringSlice(features.tags) = %v, want [alpha beta]", tags)
	}

	// Test IsSet
	if !provider.IsSet("app.name") {
		t.Error("IsSet(app.name) = false, want true")
	}
	if provider.IsSet("nonexistent") {
		t.Error("IsSet(nonexistent) = true, want false")
	}

	// Test AllKeys
	keys := provider.AllKeys()
	if len(keys) == 0 {
		t.Error("AllKeys should return non-empty slice")
	}
}

// TestProviderUnmarshal tests unmarshaling configuration
func TestProviderUnmarshal(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
app:
  name: test-app
  version: 1.0.0
database:
  host: localhost
  port: 5432
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	provider, err := viperadapter.NewProvider(configPath)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	// Test unmarshal entire config
	var config map[string]any
	if err := provider.Unmarshal("", &config); err != nil {
		t.Errorf("Unmarshal entire config failed: %v", err)
	}
	if config == nil {
		t.Fatal("Unmarshaled config should not be nil")
	}

	// Test unmarshal specific key
	type DatabaseConfig struct {
		Host string
		Port int
	}
	var dbConfig DatabaseConfig
	if err := provider.Unmarshal("database", &dbConfig); err != nil {
		t.Errorf("Unmarshal database config failed: %v", err)
	}
	if dbConfig.Host != "localhost" || dbConfig.Port != 5432 {
		t.Errorf("Database config = %+v, want {Host:localhost Port:5432}", dbConfig)
	}
}

// TestProviderWatch tests configuration file watching
func TestProviderWatch(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	initialConfig := `
app:
  name: initial
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	provider, err := viperadapter.NewProvider(configPath)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	// Register watch callback
	called := make(chan bool, 1)
	var receivedEvent hyperion.ChangeEvent
	stop, err := provider.Watch(func(event hyperion.ChangeEvent) {
		receivedEvent = event
		select {
		case called <- true:
		default:
		}
	})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}
	defer stop()

	// Modify config file
	updatedConfig := `
app:
  name: updated
`
	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	if err := os.WriteFile(configPath, []byte(updatedConfig), 0644); err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Wait for callback
	select {
	case <-called:
		// Callback was invoked
		if receivedEvent.Key != "config.yaml" {
			t.Errorf("Event key = %v, want config.yaml", receivedEvent.Key)
		}
	case <-time.After(2 * time.Second):
		t.Error("Watch callback was not invoked")
	}

	// Verify config was reloaded
	time.Sleep(100 * time.Millisecond)
	if name := provider.GetString("app.name"); name != "updated" {
		t.Errorf("Config not reloaded: app.name = %v, want updated", name)
	}
}

// TestProviderWatchMultipleCallbacks tests multiple watch callbacks
func TestProviderWatchMultipleCallbacks(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	initialConfig := `value: 1`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	provider, err := viperadapter.NewProvider(configPath)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	// Register multiple callbacks
	called1 := make(chan bool, 1)
	called2 := make(chan bool, 1)

	stop1, err := provider.Watch(func(event hyperion.ChangeEvent) {
		select {
		case called1 <- true:
		default:
		}
	})
	if err != nil {
		t.Fatalf("First Watch failed: %v", err)
	}
	defer stop1()

	stop2, err := provider.Watch(func(event hyperion.ChangeEvent) {
		select {
		case called2 <- true:
		default:
		}
	})
	if err != nil {
		t.Fatalf("Second Watch failed: %v", err)
	}
	defer stop2()

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Modify config
	if err := os.WriteFile(configPath, []byte("value: 2"), 0644); err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Both callbacks should be invoked
	timeout := time.After(2 * time.Second)
	gotCallback1 := false
	gotCallback2 := false

	for !gotCallback1 || !gotCallback2 {
		select {
		case <-called1:
			gotCallback1 = true
		case <-called2:
			gotCallback2 = true
		case <-timeout:
			if !gotCallback1 {
				t.Error("First callback was not invoked")
			}
			if !gotCallback2 {
				t.Error("Second callback was not invoked")
			}
			return
		}
	}
}

// TestNewProviderFromEnv tests environment-based provider creation
func TestNewProviderFromEnv(t *testing.T) {
	// Create a temporary config file in the default location
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "configs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `app:
  name: env-test`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Change to temp directory so relative path works
	origDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("Failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test NewProviderFromEnv (will use default path "configs/config.yaml")
	provider, err := viperadapter.NewProviderFromEnv()
	if err != nil {
		t.Fatalf("NewProviderFromEnv failed: %v", err)
	}
	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	// Verify it loaded the config
	if name := provider.GetString("app.name"); name != "env-test" {
		t.Errorf("Config not loaded: app.name = %v, want env-test", name)
	}
}

// TestProviderWatchAtomicWrite tests handling of atomic file writes (rename/remove)
func TestProviderWatchAtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	initialConfig := `value: 1`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	provider, err := viperadapter.NewProvider(configPath)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	called := make(chan bool, 1)
	stop, err := provider.Watch(func(event hyperion.ChangeEvent) {
		select {
		case called <- true:
		default:
		}
	})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}
	defer stop()

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Simulate atomic write: write to temp file then rename
	tmpFile := filepath.Join(tmpDir, "config.yaml.tmp")
	updatedConfig := `value: 2`
	if err := os.WriteFile(tmpFile, []byte(updatedConfig), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Rename (simulates atomic write used by editors and k8s)
	if err := os.Rename(tmpFile, configPath); err != nil {
		t.Fatalf("Failed to rename: %v", err)
	}

	// Wait for callback
	select {
	case <-called:
		// Success: callback invoked
	case <-time.After(2 * time.Second):
		t.Error("Watch callback was not invoked for atomic write")
	}

	// Verify config was reloaded
	time.Sleep(100 * time.Millisecond)
	if val := provider.GetInt("value"); val != 2 {
		t.Errorf("Config not reloaded: value = %v, want 2", val)
	}
}

// TestProviderWatchStop tests stopping watch
func TestProviderWatchStop(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("value: 1"), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	provider, err := viperadapter.NewProvider(configPath)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	called := make(chan bool, 1)
	stop, err := provider.Watch(func(event hyperion.ChangeEvent) {
		select {
		case called <- true:
		default:
		}
	})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Stop watching
	stop()

	// Modify config after stopping
	time.Sleep(100 * time.Millisecond)
	if err := os.WriteFile(configPath, []byte("value: 2"), 0644); err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Callback should NOT be invoked
	select {
	case <-called:
		t.Error("Callback was invoked after stop")
	case <-time.After(500 * time.Millisecond):
		// Expected: callback not invoked
	}
}
