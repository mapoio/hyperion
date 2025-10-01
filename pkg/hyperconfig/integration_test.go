package hyperconfig

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_FileWatchingRealScenario tests the complete hot reload workflow
// with real file system operations, simulating a production scenario.
func TestIntegration_FileWatchingRealScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "app.yaml")

	initialConfig := `
app:
  name: test-app
  version: 1.0.0
log:
  level: info
  format: json
database:
  host: localhost
  port: 5432
  max_connections: 100
`
	err := os.WriteFile(configPath, []byte(initialConfig), 0o644)
	require.NoError(t, err)

	// Create provider
	provider, err := NewViperProvider(configPath)
	require.NoError(t, err)

	// Verify initial configuration
	assert.Equal(t, "test-app", provider.GetString("app.name"))
	assert.Equal(t, "info", provider.GetString("log.level"))
	assert.Equal(t, "localhost", provider.GetString("database.host"))
	assert.Equal(t, 100, provider.GetInt("database.max_connections"))

	// Set up watch callback with change tracking
	var changeCount atomic.Int32
	changesDetected := make(chan bool, 10)

	stop, err := provider.Watch(func(event ChangeEvent) {
		changeCount.Add(1)
		changesDetected <- true
	})
	require.NoError(t, err)
	defer stop()

	// Give the watcher time to set up
	time.Sleep(100 * time.Millisecond)

	// Scenario 1: Update log level (simulating runtime configuration change)
	updatedConfig1 := `
app:
  name: test-app
  version: 1.0.0
log:
  level: debug
  format: json
database:
  host: localhost
  port: 5432
  max_connections: 100
`
	err = os.WriteFile(configPath, []byte(updatedConfig1), 0o644)
	require.NoError(t, err)

	// Wait for change detection
	select {
	case <-changesDetected:
		// Change detected
	case <-time.After(5 * time.Second):
		t.Fatal("first config change not detected")
	}

	// Verify the change was applied
	time.Sleep(50 * time.Millisecond) // Give viper time to reload
	assert.Equal(t, "debug", provider.GetString("log.level"))

	// Scenario 2: Update database configuration
	updatedConfig2 := `
app:
  name: test-app
  version: 1.0.0
log:
  level: debug
  format: json
database:
  host: postgres-prod
  port: 5432
  max_connections: 200
`
	err = os.WriteFile(configPath, []byte(updatedConfig2), 0o644)
	require.NoError(t, err)

	// Wait for change detection
	select {
	case <-changesDetected:
		// Change detected
	case <-time.After(5 * time.Second):
		t.Fatal("second config change not detected")
	}

	// Verify the changes
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, "postgres-prod", provider.GetString("database.host"))
	assert.Equal(t, 200, provider.GetInt("database.max_connections"))

	// Verify multiple callbacks were triggered
	assert.GreaterOrEqual(t, int(changeCount.Load()), 2, "should have detected at least 2 config changes")
}

// TestIntegration_MultipleComponentsWatching tests the scenario where
// multiple application components are watching the same configuration.
func TestIntegration_MultipleComponentsWatching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	initialConfig := `
service_a:
  enabled: true
  timeout: 30
service_b:
  enabled: true
  timeout: 60
`
	err := os.WriteFile(configPath, []byte(initialConfig), 0o644)
	require.NoError(t, err)

	provider, err := NewViperProvider(configPath)
	require.NoError(t, err)

	// Simulate two different service components watching config
	serviceANotified := make(chan bool, 5)
	serviceBNotified := make(chan bool, 5)

	stopA, err := provider.Watch(func(event ChangeEvent) {
		// Service A reloads its configuration
		timeout := provider.GetInt("service_a.timeout")
		if timeout > 0 {
			serviceANotified <- true
		}
	})
	require.NoError(t, err)
	defer stopA()

	stopB, err := provider.Watch(func(event ChangeEvent) {
		// Service B reloads its configuration
		timeout := provider.GetInt("service_b.timeout")
		if timeout > 0 {
			serviceBNotified <- true
		}
	})
	require.NoError(t, err)
	defer stopB()

	time.Sleep(100 * time.Millisecond)

	// Update configuration
	updatedConfig := `
service_a:
  enabled: true
  timeout: 45
service_b:
  enabled: false
  timeout: 90
`
	err = os.WriteFile(configPath, []byte(updatedConfig), 0o644)
	require.NoError(t, err)

	// Both services should be notified
	timeout := time.After(5 * time.Second)
	aNotified := false
	bNotified := false

	for i := 0; i < 2; i++ {
		select {
		case <-serviceANotified:
			aNotified = true
		case <-serviceBNotified:
			bNotified = true
		case <-timeout:
			t.Fatal("not all services notified of config change")
		}
	}

	assert.True(t, aNotified, "Service A should be notified")
	assert.True(t, bNotified, "Service B should be notified")

	// Verify both services can read updated config
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 45, provider.GetInt("service_a.timeout"))
	assert.Equal(t, false, provider.GetBool("service_b.enabled"))
	assert.Equal(t, 90, provider.GetInt("service_b.timeout"))
}

// TestIntegration_StopWatchingDuringOperation tests that stopping
// a watch callback works correctly even during active watching.
func TestIntegration_StopWatchingDuringOperation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(configPath, []byte("key: value1"), 0o644)
	require.NoError(t, err)

	provider, err := NewViperProvider(configPath)
	require.NoError(t, err)

	// Start watching
	var notificationCount atomic.Int32
	notifications := make(chan bool, 10)

	stop, err := provider.Watch(func(event ChangeEvent) {
		notificationCount.Add(1)
		notifications <- true
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Trigger first change
	err = os.WriteFile(configPath, []byte("key: value2"), 0o644)
	require.NoError(t, err)

	// Wait for first notification
	select {
	case <-notifications:
		// Received
	case <-time.After(5 * time.Second):
		t.Fatal("first notification not received")
	}

	// Stop watching
	stop()

	// Trigger second change - should NOT notify
	err = os.WriteFile(configPath, []byte("key: value3"), 0o644)
	require.NoError(t, err)

	// Wait a bit to ensure no notification
	time.Sleep(1 * time.Second)

	// Should have received exactly 1 notification
	assert.Equal(t, int32(1), notificationCount.Load(), "should only receive notification before stop")
}

// TestIntegration_ComplexConfigStructure tests hot reload with nested
// configuration structures, simulating a real application config.
func TestIntegration_ComplexConfigStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	type ServerConfig struct {
		Host         string `mapstructure:"host"`
		Port         int    `mapstructure:"port"`
		ReadTimeout  int    `mapstructure:"read_timeout"`
		WriteTimeout int    `mapstructure:"write_timeout"`
	}

	type DatabaseConfig struct {
		Driver         string `mapstructure:"driver"`
		DSN            string `mapstructure:"dsn"`
		MaxConnections int    `mapstructure:"max_connections"`
		MaxIdleTime    int    `mapstructure:"max_idle_time"`
	}

	type AppConfig struct {
		LogLevel string         `mapstructure:"log_level"`
		Database DatabaseConfig `mapstructure:"database"`
		Server   ServerConfig   `mapstructure:"server"`
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	initialConfig := `
log_level: info
server:
  host: localhost
  port: 8080
  read_timeout: 30
  write_timeout: 30
database:
  driver: postgres
  dsn: postgres://localhost/testdb
  max_connections: 50
  max_idle_time: 300
`
	err := os.WriteFile(configPath, []byte(initialConfig), 0o644)
	require.NoError(t, err)

	provider, err := NewViperProvider(configPath)
	require.NoError(t, err)

	// Load initial config
	var appConfig AppConfig
	err = provider.Unmarshal("", &appConfig)
	require.NoError(t, err)

	assert.Equal(t, "localhost", appConfig.Server.Host)
	assert.Equal(t, 8080, appConfig.Server.Port)
	assert.Equal(t, 50, appConfig.Database.MaxConnections)

	// Watch for changes
	configReloaded := make(chan bool, 5)
	stop, err := provider.Watch(func(event ChangeEvent) {
		configReloaded <- true
	})
	require.NoError(t, err)
	defer stop()

	time.Sleep(100 * time.Millisecond)

	// Simulate production config update
	updatedConfig := `
log_level: warn
server:
  host: 0.0.0.0
  port: 8080
  read_timeout: 60
  write_timeout: 60
database:
  driver: postgres
  dsn: postgres://prod-db/appdb
  max_connections: 100
  max_idle_time: 600
`
	err = os.WriteFile(configPath, []byte(updatedConfig), 0o644)
	require.NoError(t, err)

	// Wait for reload
	select {
	case <-configReloaded:
		// Reloaded
	case <-time.After(5 * time.Second):
		t.Fatal("config not reloaded")
	}

	// Reload config into struct
	time.Sleep(50 * time.Millisecond)
	var newAppConfig AppConfig
	err = provider.Unmarshal("", &newAppConfig)
	require.NoError(t, err)

	// Verify all changes
	assert.Equal(t, "warn", newAppConfig.LogLevel)
	assert.Equal(t, "0.0.0.0", newAppConfig.Server.Host)
	assert.Equal(t, 60, newAppConfig.Server.ReadTimeout)
	assert.Equal(t, 60, newAppConfig.Server.WriteTimeout)
	assert.Equal(t, "postgres://prod-db/appdb", newAppConfig.Database.DSN)
	assert.Equal(t, 100, newAppConfig.Database.MaxConnections)
	assert.Equal(t, 600, newAppConfig.Database.MaxIdleTime)
}
