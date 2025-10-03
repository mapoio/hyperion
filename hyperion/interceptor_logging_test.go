package hyperion

import (
	"context"
	"errors"
	"testing"
)

// captureLogger captures log calls for testing.
type captureLogger struct {
	noopLogger
	debugCalls []logCall
	errorCalls []logCall
}

type logCall struct {
	msg    string
	fields []any
}

func (c *captureLogger) Debug(msg string, fields ...any) {
	c.debugCalls = append(c.debugCalls, logCall{msg: msg, fields: fields})
}

func (c *captureLogger) Error(msg string, fields ...any) {
	c.errorCalls = append(c.errorCalls, logCall{msg: msg, fields: fields})
}

func TestNewLoggingInterceptor(t *testing.T) {
	logger := &captureLogger{}
	interceptor := NewLoggingInterceptor(logger)

	if interceptor == nil {
		t.Fatal("NewLoggingInterceptor() returned nil")
	}

	if interceptor.logger != logger {
		t.Error("logger not set correctly")
	}
}

func TestLoggingInterceptor_Name(t *testing.T) {
	logger := &captureLogger{}
	interceptor := NewLoggingInterceptor(logger)

	if interceptor.Name() != loggingInterceptorName {
		t.Errorf("Name() = %q, want %q", interceptor.Name(), loggingInterceptorName)
	}
}

func TestLoggingInterceptor_Order(t *testing.T) {
	logger := &captureLogger{}
	interceptor := NewLoggingInterceptor(logger)

	if interceptor.Order() != 200 {
		t.Errorf("Order() = %d, want 200", interceptor.Order())
	}
}

func TestLoggingInterceptor_Intercept_Success(t *testing.T) {
	logger := &captureLogger{}
	interceptor := NewLoggingInterceptor(logger)

	ctx := &hyperionContext{
		Context: context.Background(),
		logger:  logger,
	}

	fullPath := "UserService.GetUser"

	// Call Intercept
	newCtx, endFunc, err := interceptor.Intercept(ctx, fullPath)

	if err != nil {
		t.Errorf("Intercept() returned error: %v", err)
	}

	if newCtx != ctx {
		t.Error("Intercept() modified context (should not for logging)")
	}

	if endFunc == nil {
		t.Fatal("Intercept() returned nil end function")
	}

	// Verify method start was logged
	if len(logger.debugCalls) != 1 {
		t.Errorf("Expected 1 debug call for method start, got %d", len(logger.debugCalls))
	}

	if logger.debugCalls[0].msg != "Method started" {
		t.Errorf("Expected 'Method started', got %q", logger.debugCalls[0].msg)
	}

	// Call end function with success (no error)
	endFunc(nil)

	// Verify method completion was logged
	if len(logger.debugCalls) != 2 {
		t.Errorf("Expected 2 debug calls (start + completion), got %d", len(logger.debugCalls))
	}

	if logger.debugCalls[1].msg != "Method completed" {
		t.Errorf("Expected 'Method completed', got %q", logger.debugCalls[1].msg)
	}

	// Verify no error logs
	if len(logger.errorCalls) != 0 {
		t.Errorf("Expected no error calls, got %d", len(logger.errorCalls))
	}
}

func TestLoggingInterceptor_Intercept_WithError(t *testing.T) {
	logger := &captureLogger{}
	interceptor := NewLoggingInterceptor(logger)

	ctx := &hyperionContext{
		Context: context.Background(),
		logger:  logger,
	}

	fullPath := "UserService.GetUser"

	// Call Intercept
	_, endFunc, err := interceptor.Intercept(ctx, fullPath)

	if err != nil {
		t.Errorf("Intercept() returned error: %v", err)
	}

	// Verify method start was logged
	if len(logger.debugCalls) != 1 {
		t.Errorf("Expected 1 debug call for method start, got %d", len(logger.debugCalls))
	}

	// Call end function with error
	testErr := errors.New("database connection failed")
	endFunc(&testErr)

	// Verify method failure was logged as error
	if len(logger.errorCalls) != 1 {
		t.Errorf("Expected 1 error call, got %d", len(logger.errorCalls))
	}

	if logger.errorCalls[0].msg != "Method failed" {
		t.Errorf("Expected 'Method failed', got %q", logger.errorCalls[0].msg)
	}

	// Verify error details are in fields
	fields := logger.errorCalls[0].fields
	foundPath := false
	foundDuration := false
	foundError := false

	for i := 0; i < len(fields); i += 2 {
		if i+1 >= len(fields) {
			break
		}
		key := fields[i].(string)
		if key == "path" {
			foundPath = true
			if fields[i+1] != fullPath {
				t.Errorf("Expected path=%q, got %q", fullPath, fields[i+1])
			}
		}
		if key == "duration" {
			foundDuration = true
		}
		if key == "error" {
			foundError = true
			if err, ok := fields[i+1].(error); !ok || !errors.Is(err, testErr) {
				t.Errorf("Expected error=%v, got %v", testErr, fields[i+1])
			}
		}
	}

	if !foundPath {
		t.Error("path not found in error log fields")
	}
	if !foundDuration {
		t.Error("duration not found in error log fields")
	}
	if !foundError {
		t.Error("error not found in error log fields")
	}

	// Should only have start debug log, no completion debug log
	if len(logger.debugCalls) != 1 {
		t.Errorf("Expected only 1 debug call (start), got %d", len(logger.debugCalls))
	}
}

func TestLoggingInterceptor_Intercept_NilErrorPointer(t *testing.T) {
	logger := &captureLogger{}
	interceptor := NewLoggingInterceptor(logger)

	ctx := &hyperionContext{
		Context: context.Background(),
		logger:  logger,
	}

	_, endFunc, _ := interceptor.Intercept(ctx, "Test.Method")

	// Call end function with nil error pointer (success case)
	endFunc(nil)

	// Should log completion, not failure
	if len(logger.debugCalls) != 2 {
		t.Errorf("Expected 2 debug calls, got %d", len(logger.debugCalls))
	}

	if len(logger.errorCalls) != 0 {
		t.Errorf("Expected no error calls, got %d", len(logger.errorCalls))
	}
}
