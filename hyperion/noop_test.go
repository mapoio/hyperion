package hyperion_test

import (
	"context"
	"testing"
	"time"

	"github.com/mapoio/hyperion"
)

// TestNoOpLogger tests all NoOp Logger methods
func TestNoOpLogger(t *testing.T) {
	logger := hyperion.NewNoOpLogger()

	// Test all logging methods (should not panic)
	logger.Debug("debug message", "key", "value")
	logger.Info("info message", "key", "value")
	logger.Warn("warn message", "key", "value")
	logger.Error("error message", "key", "value")
	logger.Fatal("fatal message", "key", "value")

	// Test With method
	childLogger := logger.With("parent", "value")
	if childLogger == nil {
		t.Error("With should return a logger")
	}
	childLogger.Info("child log")

	// Test WithError method
	err := context.DeadlineExceeded
	errorLogger := logger.WithError(err)
	if errorLogger == nil {
		t.Error("WithError should return a logger")
	}
	errorLogger.Error("error occurred")

	// Test level methods
	logger.SetLevel(hyperion.InfoLevel)
	level := logger.GetLevel()
	if level != hyperion.InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", level)
	}

	// Test Sync
	if err := logger.Sync(); err != nil {
		t.Errorf("Sync should not error, got %v", err)
	}
}

// TestLogLevel tests LogLevel string conversion
func TestLogLevel(t *testing.T) {
	tests := []struct {
		expected string
		level    hyperion.LogLevel
	}{
		{"debug", hyperion.DebugLevel},
		{"info", hyperion.InfoLevel},
		{"warn", hyperion.WarnLevel},
		{"error", hyperion.ErrorLevel},
		{"fatal", hyperion.FatalLevel},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestNoOpConfig tests all NoOp Config methods
func TestNoOpConfig(t *testing.T) {
	cfg := hyperion.NewNoOpConfig()

	// Test Unmarshal
	var data map[string]interface{}
	if err := cfg.Unmarshal("", &data); err != nil {
		t.Errorf("Unmarshal should not error, got %v", err)
	}

	// Test Get
	if val := cfg.Get("key"); val != nil {
		t.Errorf("Get should return nil, got %v", val)
	}

	// Test GetString
	if val := cfg.GetString("key"); val != "" {
		t.Errorf("GetString should return empty string, got %v", val)
	}

	// Test GetInt
	if val := cfg.GetInt("key"); val != 0 {
		t.Errorf("GetInt should return 0, got %v", val)
	}

	// Test GetInt64
	if val := cfg.GetInt64("key"); val != 0 {
		t.Errorf("GetInt64 should return 0, got %v", val)
	}

	// Test GetBool
	if val := cfg.GetBool("key"); val != false {
		t.Errorf("GetBool should return false, got %v", val)
	}

	// Test GetFloat64
	if val := cfg.GetFloat64("key"); val != 0.0 {
		t.Errorf("GetFloat64 should return 0.0, got %v", val)
	}

	// Test GetStringSlice
	if val := cfg.GetStringSlice("key"); val != nil {
		t.Errorf("GetStringSlice should return nil, got %v", val)
	}

	// Test IsSet
	if val := cfg.IsSet("key"); val != false {
		t.Errorf("IsSet should return false, got %v", val)
	}

	// Test AllKeys
	if keys := cfg.AllKeys(); keys != nil {
		t.Errorf("AllKeys should return nil, got %v", keys)
	}
}

// TestNoOpTracer tests all NoOp Tracer methods
func TestNoOpTracer(t *testing.T) {
	tracer := hyperion.NewNoOpTracer()
	ctx := context.Background()

	// Test Start
	newCtx, span := tracer.Start(ctx, "test-span")
	if newCtx == nil {
		t.Error("Start should return a context")
	}
	if span == nil {
		t.Error("Start should return a span")
	}

	// Test Span methods
	span.SetAttributes(hyperion.String("key", "value"))
	span.RecordError(context.DeadlineExceeded)
	span.AddEvent("event", hyperion.WithEventAttributes(hyperion.String("event", "attr")))
	spanCtx := span.SpanContext()

	// Test SpanContext methods
	if spanCtx.TraceID() != "" {
		t.Error("NoOp SpanContext.TraceID should return empty string")
	}
	if spanCtx.SpanID() != "" {
		t.Error("NoOp SpanContext.SpanID should return empty string")
	}
	if spanCtx.IsValid() {
		t.Error("NoOp SpanContext.IsValid should return false")
	}

	// Test End (should not panic)
	span.End()
}

// TestSpanAttribute tests span attribute constructors
func TestSpanAttribute(t *testing.T) {
	// Just test that attributes can be created without panicking
	_ = hyperion.String("key", "value")
	_ = hyperion.Int("count", 42)
	_ = hyperion.Int64("large", 9999999999)
	_ = hyperion.Float64("pi", 3.14)
	_ = hyperion.Bool("flag", true)
}

// TestSpanOptions tests span option constructors
func TestSpanOptions(t *testing.T) {
	// Test WithAttributes (SpanOption)
	spanOpt := hyperion.WithAttributes(
		hyperion.String("key", "value"),
		hyperion.Int("count", 10),
	)
	if spanOpt == nil {
		t.Error("WithAttributes should return an option")
	}

	// Test WithSpanKind (SpanOption)
	spanOpt = hyperion.WithSpanKind(hyperion.SpanKindServer)
	if spanOpt == nil {
		t.Error("WithSpanKind should return an option")
	}

	// Test WithTimestamp (SpanOption)
	spanOpt = hyperion.WithTimestamp(time.Now())
	if spanOpt == nil {
		t.Error("WithTimestamp should return an option")
	}

	// Test WithEndTime (SpanEndOption)
	endOpt := hyperion.WithEndTime(time.Now())
	if endOpt == nil {
		t.Error("WithEndTime should return an option")
	}

	// Test WithEventAttributes (EventOption)
	eventOpt := hyperion.WithEventAttributes(hyperion.String("event", "test"))
	if eventOpt == nil {
		t.Error("WithEventAttributes should return an option")
	}

	// Test WithEventTimestamp (EventOption)
	eventOpt = hyperion.WithEventTimestamp(time.Now())
	if eventOpt == nil {
		t.Error("WithEventTimestamp should return an option")
	}
}

// TestNoOpDatabase tests all NoOp Database methods
func TestNoOpDatabase(t *testing.T) {
	db := hyperion.NewNoOpDatabase()
	ctx := context.Background()

	// Test Executor
	exec := db.Executor()
	if exec == nil {
		t.Error("Executor should return an executor")
	}

	// Test Health (NoOp returns error)
	if err := db.Health(ctx); err == nil {
		t.Error("Health should return error for NoOp database")
	}

	// Test Close
	if err := db.Close(); err != nil {
		t.Errorf("Close should not error, got %v", err)
	}

	// Test Executor methods
	exec.Exec(ctx, "SELECT 1")

	var users []map[string]any
	err := exec.Query(ctx, &users, "SELECT * FROM users")
	if err == nil {
		t.Error("Query should return error for NoOp executor")
	}

	tx, err := exec.Begin(ctx)
	if err != nil {
		t.Errorf("Begin should not error, got %v", err)
	}
	if tx == nil {
		t.Error("Begin should return an executor")
	}

	// Test transaction methods (should not panic)
	if exec.Commit() != nil {
		t.Error("Commit should return nil")
	}
	if exec.Rollback() != nil {
		t.Error("Rollback should return nil")
	}

	// Test Unwrap
	if underlying := exec.Unwrap(); underlying != nil {
		t.Error("Unwrap should return nil")
	}
}

// TestNoOpCache tests all NoOp Cache methods
func TestNoOpCache(t *testing.T) {
	cache := hyperion.NewNoOpCache()
	ctx := context.Background()

	// Test Get
	data, err := cache.Get(ctx, "key")
	if err == nil {
		t.Error("Get should return error for NoOp")
	}
	if data != nil {
		t.Error("Get should return nil data")
	}

	// Test Set
	if setErr := cache.Set(ctx, "key", []byte("value"), time.Minute); setErr != nil {
		t.Errorf("Set should not error, got %v", setErr)
	}

	// Test Delete
	if delErr := cache.Delete(ctx, "key"); delErr != nil {
		t.Errorf("Delete should not error, got %v", delErr)
	}

	// Test Exists
	exists, err := cache.Exists(ctx, "key")
	if err != nil {
		t.Errorf("Exists should not error, got %v", err)
	}
	if exists {
		t.Error("Exists should return false for NoOp")
	}

	// Test MGet (NoOp returns error)
	values, err := cache.MGet(ctx, "key1", "key2")
	if err == nil {
		t.Error("MGet should return error for NoOp cache")
	}
	if values != nil {
		t.Error("MGet should return nil values for NoOp")
	}

	// Test MSet
	if err := cache.MSet(ctx, map[string][]byte{"key1": []byte("value1")}, time.Minute); err != nil {
		t.Errorf("MSet should not error, got %v", err)
	}

	// Test Clear
	if err := cache.Clear(ctx); err != nil {
		t.Errorf("Clear should not error, got %v", err)
	}
}

// TestContext tests hyperion.Context implementation
func TestContext(t *testing.T) {
	logger := hyperion.NewNoOpLogger()
	tracer := hyperion.NewNoOpTracer()
	db := hyperion.NewNoOpDatabase()

	// Test New
	ctx := hyperion.New(context.Background(), logger, db.Executor(), tracer)
	if ctx == nil {
		t.Fatal("New should return a context")
	}

	// Test Logger
	if l := ctx.Logger(); l == nil {
		t.Error("Logger should not be nil")
	}

	// Test DB
	if d := ctx.DB(); d == nil {
		t.Error("DB should not be nil")
	}

	// Test Tracer
	if tr := ctx.Tracer(); tr == nil {
		t.Error("Tracer should not be nil")
	}

	// Test WithTimeout
	timeoutCtx, cancel := ctx.WithTimeout(time.Second)
	if timeoutCtx == nil {
		t.Error("WithTimeout should return a context")
	}
	if cancel == nil {
		t.Error("WithTimeout should return a cancel function")
	}
	cancel()

	// Test WithCancel
	cancelCtx, cancel := ctx.WithCancel()
	if cancelCtx == nil {
		t.Error("WithCancel should return a context")
	}
	if cancel == nil {
		t.Error("WithCancel should return a cancel function")
	}
	cancel()

	// Test WithDeadline
	deadline := time.Now().Add(time.Second)
	deadlineCtx, cancel := ctx.WithDeadline(deadline)
	if deadlineCtx == nil {
		t.Error("WithDeadline should return a context")
	}
	if cancel == nil {
		t.Error("WithDeadline should return a cancel function")
	}
	cancel()
}
