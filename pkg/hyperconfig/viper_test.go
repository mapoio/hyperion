package hyperconfig

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewViperProvider(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		content     string
		expectError bool
	}{
		{
			name:   "YAML format",
			format: "yaml",
			content: `
log:
  level: info
  format: json
database:
  host: localhost
  port: 5432
`,
			expectError: false,
		},
		{
			name:   "JSON format",
			format: "json",
			content: `{
  "log": {
    "level": "debug",
    "format": "text"
  },
  "database": {
    "host": "postgres",
    "port": 5432
  }
}`,
			expectError: false,
		},
		{
			name:   "TOML format",
			format: "toml",
			content: `
[log]
level = "warn"
format = "json"

[database]
host = "mysql"
port = 3306
`,
			expectError: false,
		},
		{
			name:        "invalid file",
			format:      "yaml",
			content:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError {
				// Test with non-existent file
				_, err := NewViperProvider("/nonexistent/config.yaml")
				assert.Error(t, err)
				return
			}

			// Create temp config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config."+tt.format)
			err := os.WriteFile(configPath, []byte(tt.content), 0o644)
			require.NoError(t, err)

			provider, err := NewViperProvider(configPath)
			assert.NoError(t, err)
			assert.NotNil(t, provider)
		})
	}
}

func TestViperProvider_GetMethods(t *testing.T) {
	// Create test config
	v := viper.New()
	v.Set("string_key", "test_value")
	v.Set("int_key", 42)
	v.Set("int64_key", int64(9223372036854775807))
	v.Set("bool_key", true)
	v.Set("float_key", 3.14)
	v.Set("slice_key", []string{"a", "b", "c"})
	v.Set("nested.key", "nested_value")

	provider := NewViperProviderWithViper(v)

	t.Run("GetString", func(t *testing.T) {
		assert.Equal(t, "test_value", provider.GetString("string_key"))
		assert.Equal(t, "", provider.GetString("nonexistent"))
	})

	t.Run("GetInt", func(t *testing.T) {
		assert.Equal(t, 42, provider.GetInt("int_key"))
		assert.Equal(t, 0, provider.GetInt("nonexistent"))
	})

	t.Run("GetInt64", func(t *testing.T) {
		assert.Equal(t, int64(9223372036854775807), provider.GetInt64("int64_key"))
		assert.Equal(t, int64(0), provider.GetInt64("nonexistent"))
	})

	t.Run("GetBool", func(t *testing.T) {
		assert.Equal(t, true, provider.GetBool("bool_key"))
		assert.Equal(t, false, provider.GetBool("nonexistent"))
	})

	t.Run("GetFloat64", func(t *testing.T) {
		assert.InDelta(t, 3.14, provider.GetFloat64("float_key"), 0.01)
		assert.Equal(t, 0.0, provider.GetFloat64("nonexistent"))
	})

	t.Run("GetStringSlice", func(t *testing.T) {
		assert.Equal(t, []string{"a", "b", "c"}, provider.GetStringSlice("slice_key"))
		assert.Nil(t, provider.GetStringSlice("nonexistent"))
	})

	t.Run("Get", func(t *testing.T) {
		assert.Equal(t, "test_value", provider.Get("string_key"))
		assert.Equal(t, 42, provider.Get("int_key"))
		assert.Nil(t, provider.Get("nonexistent"))
	})

	t.Run("nested keys", func(t *testing.T) {
		assert.Equal(t, "nested_value", provider.GetString("nested.key"))
	})
}

func TestViperProvider_IsSet(t *testing.T) {
	v := viper.New()
	v.Set("existing_key", "value")

	provider := NewViperProviderWithViper(v)

	assert.True(t, provider.IsSet("existing_key"))
	assert.False(t, provider.IsSet("nonexistent_key"))
}

func TestViperProvider_AllKeys(t *testing.T) {
	v := viper.New()
	v.Set("key1", "value1")
	v.Set("key2", "value2")
	v.Set("nested.key3", "value3")

	provider := NewViperProviderWithViper(v)

	keys := provider.AllKeys()
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
	assert.Contains(t, keys, "nested.key3")
}

func TestViperProvider_Unmarshal(t *testing.T) {
	type DatabaseConfig struct {
		Host     string `mapstructure:"host"`
		Username string `mapstructure:"username"`
		Port     int    `mapstructure:"port"`
	}

	type AppConfig struct {
		LogLevel string         `mapstructure:"log_level"`
		Database DatabaseConfig `mapstructure:"database"`
	}

	v := viper.New()
	v.Set("database.host", "localhost")
	v.Set("database.port", 5432)
	v.Set("database.username", "admin")
	v.Set("log_level", "debug")

	provider := NewViperProviderWithViper(v)

	t.Run("unmarshal specific key", func(t *testing.T) {
		var dbConfig DatabaseConfig
		err := provider.Unmarshal("database", &dbConfig)
		require.NoError(t, err)
		assert.Equal(t, "localhost", dbConfig.Host)
		assert.Equal(t, 5432, dbConfig.Port)
		assert.Equal(t, "admin", dbConfig.Username)
	})

	t.Run("unmarshal entire config", func(t *testing.T) {
		var appConfig AppConfig
		err := provider.Unmarshal("", &appConfig)
		require.NoError(t, err)
		assert.Equal(t, "localhost", appConfig.Database.Host)
		assert.Equal(t, "debug", appConfig.LogLevel)
	})

	t.Run("unmarshal nonexistent key", func(t *testing.T) {
		var config map[string]any
		err := provider.Unmarshal("nonexistent", &config)
		// Viper returns empty map for nonexistent keys, not an error
		assert.NoError(t, err)
	})
}

func TestViperProvider_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("APP_LOG_LEVEL", "debug")
	os.Setenv("APP_DATABASE_HOST", "envhost")
	defer func() {
		os.Unsetenv("APP_LOG_LEVEL")
		os.Unsetenv("APP_DATABASE_HOST")
	}()

	// Create config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	content := `
log:
  level: info
database:
  host: localhost
`
	err := os.WriteFile(configPath, []byte(content), 0o644)
	require.NoError(t, err)

	provider, err := NewViperProvider(configPath)
	require.NoError(t, err)

	// Environment variables should override file values
	assert.Equal(t, "debug", provider.GetString("log.level"))
	assert.Equal(t, "envhost", provider.GetString("database.host"))
}

func TestViperProvider_Watch(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	initialContent := `
log:
  level: info
`
	err := os.WriteFile(configPath, []byte(initialContent), 0o644)
	require.NoError(t, err)

	provider, err := NewViperProvider(configPath)
	require.NoError(t, err)

	// Register watch callback
	callbackCalled := make(chan bool, 1)
	stop, err := provider.Watch(func(event ChangeEvent) {
		callbackCalled <- true
	})
	require.NoError(t, err)
	defer stop()

	// Give viper time to set up watching
	time.Sleep(100 * time.Millisecond)

	// Modify config file
	updatedContent := `
log:
  level: debug
`
	err = os.WriteFile(configPath, []byte(updatedContent), 0o644)
	require.NoError(t, err)

	// Wait for callback (with timeout)
	select {
	case <-callbackCalled:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("callback not called after config change")
	}

	// Verify config was reloaded
	assert.Equal(t, "debug", provider.GetString("log.level"))
}

func TestViperProvider_Watch_MultipleCallbacks(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	initialContent := `test: value1`
	err := os.WriteFile(configPath, []byte(initialContent), 0o644)
	require.NoError(t, err)

	provider, err := NewViperProvider(configPath)
	require.NoError(t, err)

	// Register multiple callbacks
	callback1Called := make(chan bool, 1)
	callback2Called := make(chan bool, 1)

	stop1, err := provider.Watch(func(event ChangeEvent) {
		callback1Called <- true
	})
	require.NoError(t, err)
	defer stop1()

	stop2, err := provider.Watch(func(event ChangeEvent) {
		callback2Called <- true
	})
	require.NoError(t, err)
	defer stop2()

	// Give viper time to set up watching
	time.Sleep(100 * time.Millisecond)

	// Modify config file
	err = os.WriteFile(configPath, []byte("test: value2"), 0o644)
	require.NoError(t, err)

	// Both callbacks should be called
	timeout := time.After(5 * time.Second)
	callback1Received := false
	callback2Received := false

	for i := 0; i < 2; i++ {
		select {
		case <-callback1Called:
			callback1Received = true
		case <-callback2Called:
			callback2Received = true
		case <-timeout:
			t.Fatal("not all callbacks called")
		}
	}

	assert.True(t, callback1Received, "callback1 should be called")
	assert.True(t, callback2Received, "callback2 should be called")
}

func TestViperProvider_Watch_Stop(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	initialContent := `test: value1`
	err := os.WriteFile(configPath, []byte(initialContent), 0o644)
	require.NoError(t, err)

	provider, err := NewViperProvider(configPath)
	require.NoError(t, err)

	// Register callback
	callbackCalled := make(chan bool, 1)
	stop, err := provider.Watch(func(event ChangeEvent) {
		callbackCalled <- true
	})
	require.NoError(t, err)

	// Give viper time to set up watching
	time.Sleep(100 * time.Millisecond)

	// Stop watching before modifying file
	stop()

	// Modify config file
	err = os.WriteFile(configPath, []byte("test: value2"), 0o644)
	require.NoError(t, err)

	// Callback should NOT be called after stop
	select {
	case <-callbackCalled:
		t.Fatal("callback should not be called after stop")
	case <-time.After(1 * time.Second):
		// Success - callback not called
	}
}
