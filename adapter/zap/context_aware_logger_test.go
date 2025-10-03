package zap

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mapoio/hyperion"
)

// mockSpan implements trace.Span for testing
type mockSpan struct {
	trace.Span
	traceID trace.TraceID
	spanID  trace.SpanID
}

func (m *mockSpan) SpanContext() trace.SpanContext {
	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    m.traceID,
		SpanID:     m.spanID,
		TraceFlags: trace.FlagsSampled,
	})
}

func (m *mockSpan) IsRecording() bool {
	return true
}

// TestContextAwareLogger_AutomaticTraceInjection verifies that trace_id and span_id
// are automatically injected into logs when using a context with an active span.
func TestContextAwareLogger_AutomaticTraceInjection(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create zap logger that writes JSON to buffer
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		MessageKey:     "msg",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	})

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)

	// Wrap with OTel core
	otelCore := newOtelCore(core)
	zapCore := zap.New(otelCore)

	logger := &zapLogger{
		sugar:         zapCore.Sugar(),
		atom:          zap.NewAtomicLevelAt(zapcore.InfoLevel),
		core:          zapCore,
		contextLogger: newContextLogger(zapCore),
	}

	// Create mock trace context
	traceID := trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID := trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}

	mockSpan := &mockSpan{
		traceID: traceID,
		spanID:  spanID,
	}

	ctx := trace.ContextWithSpan(context.Background(), mockSpan)

	// Create context-aware logger
	ctxLogger := newContextAwareLogger(ctx, logger)

	// Log a message
	ctxLogger.Info("test message", "key", "value")

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v\nOutput: %s", err, buf.String())
	}

	// Verify trace_id and span_id are present
	expectedTraceID := traceID.String()
	expectedSpanID := spanID.String()

	if logEntry["trace_id"] != expectedTraceID {
		t.Errorf("Expected trace_id %q, got %q", expectedTraceID, logEntry["trace_id"])
	}

	if logEntry["span_id"] != expectedSpanID {
		t.Errorf("Expected span_id %q, got %q", expectedSpanID, logEntry["span_id"])
	}

	// Verify message and custom field
	if logEntry["msg"] != "test message" {
		t.Errorf("Expected msg %q, got %q", "test message", logEntry["msg"])
	}

	if logEntry["key"] != "value" {
		t.Errorf("Expected key=value, got key=%q", logEntry["key"])
	}
}

// TestContextAwareLogger_NoTraceContext verifies that logging works normally
// when there's no active span in the context.
func TestContextAwareLogger_NoTraceContext(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		MessageKey:     "msg",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	})

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)

	otelCore := newOtelCore(core)
	zapCore := zap.New(otelCore)

	logger := &zapLogger{
		sugar:         zapCore.Sugar(),
		atom:          zap.NewAtomicLevelAt(zapcore.InfoLevel),
		core:          zapCore,
		contextLogger: newContextLogger(zapCore),
	}

	// Create context WITHOUT span
	ctx := context.Background()

	// Create context-aware logger
	ctxLogger := newContextAwareLogger(ctx, logger)

	// Log a message
	ctxLogger.Info("test message without trace")

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Verify trace_id and span_id are NOT present
	if _, exists := logEntry["trace_id"]; exists {
		t.Error("trace_id should not be present when there's no active span")
	}

	if _, exists := logEntry["span_id"]; exists {
		t.Error("span_id should not be present when there's no active span")
	}

	// Verify message is still logged
	if logEntry["msg"] != "test message without trace" {
		t.Errorf("Expected msg %q, got %q", "test message without trace", logEntry["msg"])
	}
}

// TestContextAwareLogger_WithMethod verifies that With() creates a child logger
// that preserves the context binding.
func TestContextAwareLogger_WithMethod(t *testing.T) {
	var buf bytes.Buffer

	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		MessageKey:     "msg",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	})

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)

	otelCore := newOtelCore(core)
	zapCore := zap.New(otelCore)

	logger := &zapLogger{
		sugar:         zapCore.Sugar(),
		atom:          zap.NewAtomicLevelAt(zapcore.InfoLevel),
		core:          zapCore,
		contextLogger: newContextLogger(zapCore),
	}

	// Create mock trace context
	traceID := trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID := trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}

	mockSpan := &mockSpan{
		traceID: traceID,
		spanID:  spanID,
	}

	ctx := trace.ContextWithSpan(context.Background(), mockSpan)

	// Create context-aware logger with additional fields
	ctxLogger := newContextAwareLogger(ctx, logger)
	childLogger := ctxLogger.With("service", "user-service", "version", "1.0")

	// Log a message
	childLogger.Info("child logger test")

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v\nOutput: %s", err, buf.String())
	}

	// Debug: print actual log entry
	t.Logf("Log entry: %+v", logEntry)

	// Verify trace context is still present
	if logEntry["trace_id"] != traceID.String() {
		t.Error("trace_id should be preserved in child logger")
	}

	if logEntry["span_id"] != spanID.String() {
		t.Error("span_id should be preserved in child logger")
	}

	// Note: Additional fields from With() are tested in TestZapLogger_With.
	// This test focuses on verifying that trace context is preserved
	// when creating child loggers with With().
}

// TestZapLogger_WithContext verifies the ContextAwareLogger interface implementation.
func TestZapLogger_WithContext(t *testing.T) {
	// Create a test logger
	logger, err := NewZapLogger(nil)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Verify it implements ContextAwareLogger
	ctxAwareLogger, ok := logger.(hyperion.ContextAwareLogger)
	if !ok {
		t.Fatal("zapLogger should implement hyperion.ContextAwareLogger")
	}

	// Create a context with trace
	traceID := trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID := trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}

	mockSpan := &mockSpan{
		traceID: traceID,
		spanID:  spanID,
	}

	ctx := trace.ContextWithSpan(context.Background(), mockSpan)

	// Call WithContext
	ctxLogger := ctxAwareLogger.WithContext(ctx)

	// Verify it returns a valid logger
	if ctxLogger == nil {
		t.Fatal("WithContext should return a non-nil logger")
	}

	// Verify it's a contextAwareLogger
	if _, ok := ctxLogger.(*contextAwareLogger); !ok {
		t.Error("WithContext should return a contextAwareLogger instance")
	}
}

// TestContextAwareLogger_AllLogLevels verifies all log level methods work correctly.
func TestContextAwareLogger_AllLogLevels(t *testing.T) {
	var buf bytes.Buffer

	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	})

	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	otelCore := newOtelCore(core)
	zapCore := zap.New(otelCore)

	logger := &zapLogger{
		sugar:         zapCore.Sugar(),
		atom:          zap.NewAtomicLevelAt(zapcore.DebugLevel),
		core:          zapCore,
		contextLogger: newContextLogger(zapCore),
	}

	traceID := trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID := trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}
	mockSpan := &mockSpan{traceID: traceID, spanID: spanID}
	ctx := trace.ContextWithSpan(context.Background(), mockSpan)

	ctxLogger := newContextAwareLogger(ctx, logger)

	// Test Debug
	ctxLogger.Debug("debug message")
	// Test Info (already tested)
	ctxLogger.Info("info message")
	// Test Warn
	ctxLogger.Warn("warn message")
	// Test Error
	ctxLogger.Error("error message")

	// Parse output - should have 4 log entries
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	if len(lines) < 4 {
		t.Fatalf("Expected at least 4 log entries, got %d", len(lines))
	}

	// Verify each level
	levels := []string{"debug", "info", "warn", "error"}
	for i, expectedLevel := range levels {
		var logEntry map[string]any
		if err := json.Unmarshal(lines[i], &logEntry); err != nil {
			t.Fatalf("Failed to parse log entry %d: %v", i, err)
		}

		if logEntry["level"] != expectedLevel {
			t.Errorf("Expected level %q, got %q", expectedLevel, logEntry["level"])
		}

		// Verify trace context in all levels
		if logEntry["trace_id"] != traceID.String() {
			t.Errorf("trace_id missing in %s level", expectedLevel)
		}
		if logEntry["span_id"] != spanID.String() {
			t.Errorf("span_id missing in %s level", expectedLevel)
		}
	}
}

// TestContextAwareLogger_WithError verifies WithError method.
func TestContextAwareLogger_WithError(t *testing.T) {
	logger, _ := NewZapLogger(nil)
	ctx := context.Background()
	ctxLogger := newContextAwareLogger(ctx, logger.(*zapLogger))

	// Test WithError creates a new logger
	testErr := bytes.ErrTooLarge
	errLogger := ctxLogger.WithError(testErr)

	if errLogger == nil {
		t.Error("WithError should return a non-nil logger")
	}

	// Verify it's still a contextAwareLogger
	if _, ok := errLogger.(*contextAwareLogger); !ok {
		t.Error("WithError should return a contextAwareLogger instance")
	}
}

// TestContextAwareLogger_SetGetLevel verifies SetLevel and GetLevel methods.
func TestContextAwareLogger_SetGetLevel(t *testing.T) {
	logger, _ := NewZapLogger(nil)
	ctx := context.Background()
	ctxLogger := newContextAwareLogger(ctx, logger.(*zapLogger))

	// Test SetLevel
	ctxLogger.SetLevel(hyperion.WarnLevel)
	if level := ctxLogger.GetLevel(); level != hyperion.WarnLevel {
		t.Errorf("Expected level WarnLevel, got %v", level)
	}

	ctxLogger.SetLevel(hyperion.ErrorLevel)
	if level := ctxLogger.GetLevel(); level != hyperion.ErrorLevel {
		t.Errorf("Expected level ErrorLevel, got %v", level)
	}
}

// TestContextAwareLogger_Sync verifies Sync method.
func TestContextAwareLogger_Sync(t *testing.T) {
	logger, _ := NewZapLogger(nil)
	ctx := context.Background()
	ctxLogger := newContextAwareLogger(ctx, logger.(*zapLogger))

	// Sync should not return error (or acceptable error for stdout)
	err := ctxLogger.Sync()
	if err != nil {
		// Acceptable for stdout/stderr
		t.Logf("Sync returned error (acceptable for stdout/stderr): %v", err)
	}
}

// TestOtelCore_DisabledLevel verifies that disabled log levels are filtered.
func TestOtelCore_DisabledLevel(t *testing.T) {
	var buf bytes.Buffer

	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	})

	// Create core with WarnLevel - Debug and Info should be filtered
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.WarnLevel)
	otelCore := newOtelCore(core)
	zapCore := zap.New(otelCore)

	logger := &zapLogger{
		sugar:         zapCore.Sugar(),
		atom:          zap.NewAtomicLevelAt(zapcore.WarnLevel),
		core:          zapCore,
		contextLogger: newContextLogger(zapCore),
	}

	ctx := context.Background()
	ctxLogger := newContextAwareLogger(ctx, logger)

	// These should be filtered (not logged)
	ctxLogger.Debug("debug message")
	ctxLogger.Info("info message")

	// These should be logged
	ctxLogger.Warn("warn message")
	ctxLogger.Error("error message")

	// Verify only 2 log entries (warn and error)
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	logCount := 0
	for _, line := range lines {
		if len(line) > 0 {
			logCount++
		}
	}

	if logCount != 2 {
		t.Errorf("Expected 2 log entries (warn and error), got %d", logCount)
	}
}
