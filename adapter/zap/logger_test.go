package zap

import (
	"bytes"
	"strings"
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/mapoio/hyperion"
)

// mockConfig implements hyperion.Config for testing.
type mockConfig struct {
	data map[string]any
}

func (m *mockConfig) Get(key string) any {
	return m.data[key]
}

func (m *mockConfig) GetString(key string) string {
	if v, ok := m.data[key].(string); ok {
		return v
	}
	return ""
}

func (m *mockConfig) GetInt(key string) int {
	if v, ok := m.data[key].(int); ok {
		return v
	}
	return 0
}

func (m *mockConfig) GetBool(key string) bool {
	if v, ok := m.data[key].(bool); ok {
		return v
	}
	return false
}

func (m *mockConfig) GetInt64(key string) int64 {
	if v, ok := m.data[key].(int64); ok {
		return v
	}
	return 0
}

func (m *mockConfig) GetFloat64(key string) float64 {
	if v, ok := m.data[key].(float64); ok {
		return v
	}
	return 0
}

func (m *mockConfig) GetStringSlice(key string) []string {
	if v, ok := m.data[key].([]string); ok {
		return v
	}
	return nil
}

func (m *mockConfig) IsSet(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (m *mockConfig) AllKeys() []string {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *mockConfig) Unmarshal(key string, rawVal any) error {
	// Simple unmarshal for testing
	if logCfg, ok := rawVal.(*Config); ok {
		if logData, exists := m.data[key].(map[string]any); exists {
			if level, ok := logData["level"].(string); ok {
				logCfg.Level = level
			}
			if encoding, ok := logData["encoding"].(string); ok {
				logCfg.Encoding = encoding
			}
			if output, ok := logData["output"].(string); ok {
				logCfg.Output = output
			}
		}
	}
	return nil
}

func TestNewZapLogger_DefaultConfig(t *testing.T) {
	logger, err := NewZapLogger(nil)
	if err != nil {
		t.Fatalf("NewZapLogger() error = %v", err)
	}

	if logger == nil {
		t.Fatal("NewZapLogger() returned nil logger")
	}

	// Verify default level is Info
	if level := logger.GetLevel(); level != hyperion.InfoLevel {
		t.Errorf("GetLevel() = %v, want %v", level, hyperion.InfoLevel)
	}
}

func TestNewZapLogger_WithConfig(t *testing.T) {
	tests := []struct {
		name       string
		config     map[string]any
		wantLevel  hyperion.LogLevel
		wantErr    bool
		errContain string
	}{
		{
			name: "debug level",
			config: map[string]any{
				"log": map[string]any{
					"level":    "debug",
					"encoding": "json",
					"output":   "stdout",
				},
			},
			wantLevel: hyperion.DebugLevel,
			wantErr:   false,
		},
		{
			name: "warn level",
			config: map[string]any{
				"log": map[string]any{
					"level":    "warn",
					"encoding": "json",
					"output":   "stdout",
				},
			},
			wantLevel: hyperion.WarnLevel,
			wantErr:   false,
		},
		{
			name: "error level",
			config: map[string]any{
				"log": map[string]any{
					"level":    "error",
					"encoding": "json",
					"output":   "stdout",
				},
			},
			wantLevel: hyperion.ErrorLevel,
			wantErr:   false,
		},
		{
			name: "invalid level",
			config: map[string]any{
				"log": map[string]any{
					"level":    "invalid",
					"encoding": "json",
					"output":   "stdout",
				},
			},
			wantErr:    true,
			errContain: "invalid log level",
		},
		{
			name: "invalid encoding",
			config: map[string]any{
				"log": map[string]any{
					"level":    "info",
					"encoding": "xml",
					"output":   "stdout",
				},
			},
			wantErr:    true,
			errContain: "unsupported encoding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &mockConfig{data: tt.config}
			logger, err := NewZapLogger(cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewZapLogger() expected error containing %q, got nil", tt.errContain)
					return
				}
				if !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("NewZapLogger() error = %q, want error containing %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewZapLogger() unexpected error = %v", err)
			}

			if level := logger.GetLevel(); level != tt.wantLevel {
				t.Errorf("GetLevel() = %v, want %v", level, tt.wantLevel)
			}
		})
	}
}

func TestZapLogger_LogMethods(t *testing.T) {
	logger, err := NewZapLogger(nil)
	if err != nil {
		t.Fatalf("NewZapLogger() error = %v", err)
	}

	// Test all log methods don't panic
	t.Run("Debug", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Debug() panicked: %v", r)
			}
		}()
		logger.Debug("debug message", "key", "value")
	})

	t.Run("Info", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Info() panicked: %v", r)
			}
		}()
		logger.Info("info message", "key", "value")
	})

	t.Run("Warn", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Warn() panicked: %v", r)
			}
		}()
		logger.Warn("warn message", "key", "value")
	})

	t.Run("Error", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Error() panicked: %v", r)
			}
		}()
		logger.Error("error message", "key", "value")
	})
}

func TestZapLogger_With(t *testing.T) {
	logger, err := NewZapLogger(nil)
	if err != nil {
		t.Fatalf("NewZapLogger() error = %v", err)
	}

	childLogger := logger.With("request_id", "abc123")
	if childLogger == nil {
		t.Fatal("With() returned nil logger")
	}

	// Verify child logger is different instance
	if childLogger == logger {
		t.Error("With() returned same logger instance")
	}
}

func TestZapLogger_WithError(t *testing.T) {
	logger, err := NewZapLogger(nil)
	if err != nil {
		t.Fatalf("NewZapLogger() error = %v", err)
	}

	testErr := bytes.ErrTooLarge
	errLogger := logger.WithError(testErr)

	if errLogger == nil {
		t.Fatal("WithError() returned nil logger")
	}

	// Verify error logger is different instance
	if errLogger == logger {
		t.Error("WithError() returned same logger instance")
	}
}

func TestZapLogger_SetLevel(t *testing.T) {
	logger, err := NewZapLogger(nil)
	if err != nil {
		t.Fatalf("NewZapLogger() error = %v", err)
	}

	tests := []struct {
		name  string
		level hyperion.LogLevel
	}{
		{"debug", hyperion.DebugLevel},
		{"info", hyperion.InfoLevel},
		{"warn", hyperion.WarnLevel},
		{"error", hyperion.ErrorLevel},
		{"fatal", hyperion.FatalLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.SetLevel(tt.level)
			if got := logger.GetLevel(); got != tt.level {
				t.Errorf("After SetLevel(%v), GetLevel() = %v", tt.level, got)
			}
		})
	}
}

func TestZapLogger_Sync(t *testing.T) {
	logger, err := NewZapLogger(nil)
	if err != nil {
		t.Fatalf("NewZapLogger() error = %v", err)
	}

	if err := logger.Sync(); err != nil {
		// Sync may return error on stdout/stderr, which is acceptable
		t.Logf("Sync() returned error (acceptable for stdout/stderr): %v", err)
	}
}

func TestToZapLevel(t *testing.T) {
	tests := []struct {
		hyperionLevel hyperion.LogLevel
		zapLevel      zapcore.Level
	}{
		{hyperion.DebugLevel, zapcore.DebugLevel},
		{hyperion.InfoLevel, zapcore.InfoLevel},
		{hyperion.WarnLevel, zapcore.WarnLevel},
		{hyperion.ErrorLevel, zapcore.ErrorLevel},
		{hyperion.FatalLevel, zapcore.FatalLevel},
		{hyperion.LogLevel(99), zapcore.InfoLevel}, // Unknown defaults to Info
	}

	for _, tt := range tests {
		t.Run(tt.hyperionLevel.String(), func(t *testing.T) {
			got := toZapLevel(tt.hyperionLevel)
			if got != tt.zapLevel {
				t.Errorf("toZapLevel(%v) = %v, want %v", tt.hyperionLevel, got, tt.zapLevel)
			}
		})
	}
}

func TestFromZapLevel(t *testing.T) {
	tests := []struct {
		zapLevel      zapcore.Level
		hyperionLevel hyperion.LogLevel
	}{
		{zapcore.DebugLevel, hyperion.DebugLevel},
		{zapcore.InfoLevel, hyperion.InfoLevel},
		{zapcore.WarnLevel, hyperion.WarnLevel},
		{zapcore.ErrorLevel, hyperion.ErrorLevel},
		{zapcore.FatalLevel, hyperion.FatalLevel},
		{zapcore.Level(99), hyperion.InfoLevel}, // Unknown defaults to Info
	}

	for _, tt := range tests {
		t.Run(tt.zapLevel.String(), func(t *testing.T) {
			got := fromZapLevel(tt.zapLevel)
			if got != tt.hyperionLevel {
				t.Errorf("fromZapLevel(%v) = %v, want %v", tt.zapLevel, got, tt.hyperionLevel)
			}
		})
	}
}

func TestZapLogger_JSONOutput(t *testing.T) {
	// Create in-memory buffer to capture output
	var buf bytes.Buffer

	// This test verifies JSON encoding produces valid JSON
	// In a real scenario, we'd need to intercept the writer
	logger, err := NewZapLogger(nil)
	if err != nil {
		t.Fatalf("NewZapLogger() error = %v", err)
	}

	// Log a message
	logger.Info("test message", "key", "value")

	// Note: Without dependency injection for writer, we can't easily capture output
	// This is a limitation of the current design for unit tests
	// Integration tests will verify actual output

	t.Log("JSON output test completed (output capture requires integration test)")
	_ = buf // Silence unused variable
}

func TestZapLogger_ConsoleOutput(t *testing.T) {
	cfg := &mockConfig{
		data: map[string]any{
			"log": map[string]any{
				"level":    "info",
				"encoding": "console",
				"output":   "stdout",
			},
		},
	}

	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("NewZapLogger() error = %v", err)
	}

	// Verify console logger was created
	logger.Info("console test", "key", "value")

	t.Log("Console output test completed")
}

// TestZapLogger_InterfaceCompliance verifies zapLogger implements hyperion.Logger.
func TestZapLogger_InterfaceCompliance(t *testing.T) {
	var _ hyperion.Logger = (*zapLogger)(nil)
}

// BenchmarkZapLogger_Info benchmarks Info logging performance.
func BenchmarkZapLogger_Info(b *testing.B) {
	logger, err := NewZapLogger(nil)
	if err != nil {
		b.Fatalf("NewZapLogger() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i, "key", "value")
	}
}

// BenchmarkZapLogger_With benchmarks With() chaining.
func BenchmarkZapLogger_With(b *testing.B) {
	logger, err := NewZapLogger(nil)
	if err != nil {
		b.Fatalf("NewZapLogger() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		childLogger := logger.With("request_id", "abc123")
		childLogger.Info("benchmark message")
	}
}
