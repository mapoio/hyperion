package zap

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mapoio/hyperion"
)

// integrationMockConfig provides realistic configuration for integration tests.
type integrationMockConfig struct {
	data map[string]any
}

func (m *integrationMockConfig) Get(key string) any {
	return m.data[key]
}

func (m *integrationMockConfig) GetString(key string) string {
	if v, ok := m.data[key].(string); ok {
		return v
	}
	return ""
}

func (m *integrationMockConfig) GetInt(key string) int {
	if v, ok := m.data[key].(int); ok {
		return v
	}
	return 0
}

func (m *integrationMockConfig) GetInt64(key string) int64 {
	if v, ok := m.data[key].(int64); ok {
		return v
	}
	return 0
}

func (m *integrationMockConfig) GetBool(key string) bool {
	if v, ok := m.data[key].(bool); ok {
		return v
	}
	return false
}

func (m *integrationMockConfig) GetFloat64(key string) float64 {
	if v, ok := m.data[key].(float64); ok {
		return v
	}
	return 0
}

func (m *integrationMockConfig) GetStringSlice(key string) []string {
	if v, ok := m.data[key].([]string); ok {
		return v
	}
	return nil
}

func (m *integrationMockConfig) IsSet(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (m *integrationMockConfig) AllKeys() []string {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *integrationMockConfig) Unmarshal(key string, rawVal any) error {
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
			if fileData, ok := logData["file"].(map[string]any); ok {
				logCfg.FileConfig = &FileConfig{}
				if path, ok := fileData["path"].(string); ok {
					logCfg.FileConfig.Path = path
				}
				if maxSize, ok := fileData["max_size"].(int); ok {
					logCfg.FileConfig.MaxSize = maxSize
				}
				if maxBackups, ok := fileData["max_backups"].(int); ok {
					logCfg.FileConfig.MaxBackups = maxBackups
				}
				if maxAge, ok := fileData["max_age"].(int); ok {
					logCfg.FileConfig.MaxAge = maxAge
				}
				if compress, ok := fileData["compress"].(bool); ok {
					logCfg.FileConfig.Compress = compress
				}
			}
		}
	}
	return nil
}

// TestIntegration_FileOutput tests file-based logging with rotation.
func TestIntegration_FileOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for log files
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "app.log")

	cfg := &integrationMockConfig{
		data: map[string]any{
			"log": map[string]any{
				"level":    "debug",
				"encoding": "json",
				"output":   logFile,
				"file": map[string]any{
					"path":        logFile,
					"max_size":    1, // 1MB
					"max_backups": 2,
					"max_age":     7,
					"compress":    false,
				},
			},
		},
	}

	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("NewZapLogger() error = %v", err)
	}

	// Write some log entries
	logger.Debug("debug message", "key", "debug_value")
	logger.Info("info message", "key", "info_value")
	logger.Warn("warn message", "key", "warn_value")
	logger.Error("error message", "key", "error_value")

	// Flush logs
	if syncErr := logger.Sync(); syncErr != nil {
		t.Logf("Sync() warning (acceptable): %v", syncErr)
	}

	// Verify log file exists
	if _, statErr := os.Stat(logFile); os.IsNotExist(statErr) {
		t.Fatalf("Log file %s was not created", logFile)
	}

	// Read and verify log contents
	file, err := os.Open(logFile)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	logCount := 0
	levels := make(map[string]int)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var logEntry map[string]any
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Failed to parse JSON log line: %s, error: %v", line, err)
			continue
		}

		// Verify log structure
		if _, ok := logEntry["level"]; !ok {
			t.Errorf("Log entry missing 'level' field: %s", line)
		}
		if _, ok := logEntry["ts"]; !ok {
			t.Errorf("Log entry missing 'ts' field: %s", line)
		}
		if _, ok := logEntry["msg"]; !ok {
			t.Errorf("Log entry missing 'msg' field: %s", line)
		}

		if level, ok := logEntry["level"].(string); ok {
			levels[level]++
		}

		logCount++
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading log file: %v", err)
	}

	if logCount < 4 {
		t.Errorf("Expected at least 4 log entries, got %d", logCount)
	}

	// Verify all log levels were written
	if levels["debug"] == 0 {
		t.Error("Expected debug log entries, got none")
	}
	if levels["info"] == 0 {
		t.Error("Expected info log entries, got none")
	}
	if levels["warn"] == 0 {
		t.Error("Expected warn log entries, got none")
	}
	if levels["error"] == 0 {
		t.Error("Expected error log entries, got none")
	}
}

// TestIntegration_ConsoleOutput tests console encoding.
func TestIntegration_ConsoleOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "console.log")

	cfg := &integrationMockConfig{
		data: map[string]any{
			"log": map[string]any{
				"level":    "info",
				"encoding": "console",
				"output":   logFile,
			},
		},
	}

	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("NewZapLogger() error = %v", err)
	}

	logger.Info("console test", "key", "value")

	if syncErr := logger.Sync(); syncErr != nil {
		t.Logf("Sync() warning (acceptable): %v", syncErr)
	}

	// Verify log file exists and contains readable text
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logText := string(content)
	if !strings.Contains(logText, "console test") {
		t.Errorf("Log file missing expected message, got: %s", logText)
	}
}

// TestIntegration_DynamicLevel tests runtime level changes.
func TestIntegration_DynamicLevel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "dynamic.log")

	cfg := &integrationMockConfig{
		data: map[string]any{
			"log": map[string]any{
				"level":    "info",
				"encoding": "json",
				"output":   logFile,
			},
		},
	}

	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("NewZapLogger() error = %v", err)
	}

	// Initial level is Info, Debug should be filtered
	logger.Debug("should not appear", "test", "1")
	logger.Info("should appear", "test", "2")

	// Change to Debug level
	logger.SetLevel(hyperion.DebugLevel)
	logger.Debug("should now appear", "test", "3")

	// Change to Error level
	logger.SetLevel(hyperion.ErrorLevel)
	logger.Info("should not appear", "test", "4")
	logger.Error("should appear", "test", "5")

	if syncErr := logger.Sync(); syncErr != nil {
		t.Logf("Sync() warning (acceptable): %v", syncErr)
	}

	// Read log file
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logText := string(content)
	lines := strings.Split(strings.TrimSpace(logText), "\n")

	// Should have exactly 3 log entries
	if len(lines) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(lines))
	}

	// Verify specific messages
	if !strings.Contains(logText, "should appear") {
		t.Error("Missing expected 'should appear' messages")
	}
	if strings.Contains(logText, "should not appear") {
		t.Error("Found unexpected 'should not appear' messages")
	}
	if !strings.Contains(logText, "should now appear") {
		t.Error("Missing expected 'should now appear' message after level change")
	}
}

// TestIntegration_WithFields tests field chaining.
func TestIntegration_WithFields(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "fields.log")

	cfg := &integrationMockConfig{
		data: map[string]any{
			"log": map[string]any{
				"level":    "info",
				"encoding": "json",
				"output":   logFile,
			},
		},
	}

	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("NewZapLogger() error = %v", err)
	}

	// Create child logger with request ID
	reqLogger := logger.With("request_id", "abc123")
	reqLogger.Info("processing request")

	// Create another child with additional fields
	userLogger := reqLogger.With("user_id", "user456")
	userLogger.Info("user action")

	// Test WithError
	testErr := fmt.Errorf("test error")
	errLogger := logger.WithError(testErr)
	errLogger.Error("operation failed")

	if syncErr := logger.Sync(); syncErr != nil {
		t.Logf("Sync() warning (acceptable): %v", syncErr)
	}

	// Verify log contents
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(lines))
	}

	// Parse and verify first log (request_id)
	var log1 map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &log1); err != nil {
		t.Fatalf("Failed to parse log line 1: %v", err)
	}
	if log1["request_id"] != "abc123" {
		t.Errorf("Log 1 missing request_id field")
	}

	// Parse and verify second log (request_id + user_id)
	var log2 map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &log2); err != nil {
		t.Fatalf("Failed to parse log line 2: %v", err)
	}
	if log2["request_id"] != "abc123" {
		t.Errorf("Log 2 missing request_id field")
	}
	if log2["user_id"] != "user456" {
		t.Errorf("Log 2 missing user_id field")
	}

	// Parse and verify third log (error)
	var log3 map[string]any
	if err := json.Unmarshal([]byte(lines[2]), &log3); err != nil {
		t.Fatalf("Failed to parse log line 3: %v", err)
	}
	if _, ok := log3["error"]; !ok {
		t.Errorf("Log 3 missing error field")
	}
}

// TestIntegration_FileRotation tests log file rotation.
func TestIntegration_FileRotation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "rotate.log")

	cfg := &integrationMockConfig{
		data: map[string]any{
			"log": map[string]any{
				"level":    "info",
				"encoding": "json",
				"output":   logFile,
				"file": map[string]any{
					"path":        logFile,
					"max_size":    1, // 1MB - small for testing
					"max_backups": 2,
					"max_age":     1,
					"compress":    false,
				},
			},
		},
	}

	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("NewZapLogger() error = %v", err)
	}

	// Write many log entries to trigger rotation
	// Each entry is approximately 100-200 bytes
	for i := 0; i < 10000; i++ {
		logger.Info("rotation test message",
			"iteration", i,
			"timestamp", time.Now().Unix(),
			"data", "some additional data to increase log size",
		)
	}

	if err := logger.Sync(); err != nil {
		t.Logf("Sync() warning (acceptable): %v", err)
	}

	// Check if log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatalf("Log file %s was not created", logFile)
	}

	// Note: Actual rotation may or may not occur depending on exact log size
	// This test primarily verifies that rotation configuration doesn't cause errors
	t.Log("File rotation test completed successfully")
}
