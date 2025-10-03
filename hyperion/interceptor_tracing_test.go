package hyperion

import (
	"context"
	"errors"
	"testing"
	"time"
)

// captureTracer captures tracer calls for testing.
type captureTracer struct {
	startCalls []traceStartCall
	spans      []*captureSpan
}

type traceStartCall struct {
	ctx      Context
	spanName string
}

func (c *captureTracer) Start(ctx Context, spanName string, opts ...SpanOption) (Context, Span) {
	c.startCalls = append(c.startCalls, traceStartCall{ctx: ctx, spanName: spanName})
	span := &captureSpan{}
	c.spans = append(c.spans, span)
	return ctx, span
}

// captureSpan captures span operations for testing.
type captureSpan struct {
	noopSpan
	recordedErrors []error
	ended          bool
}

func (c *captureSpan) RecordError(err error, opts ...EventOption) {
	c.recordedErrors = append(c.recordedErrors, err)
}

func (c *captureSpan) End(opts ...SpanEndOption) {
	c.ended = true
}

func TestNewTracingInterceptor(t *testing.T) {
	tracer := &captureTracer{}
	interceptor := NewTracingInterceptor(tracer)

	if interceptor == nil {
		t.Fatal("NewTracingInterceptor() returned nil")
	}

	if interceptor.tracer != tracer {
		t.Error("tracer not set correctly")
	}
}

func TestTracingInterceptor_Name(t *testing.T) {
	tracer := &captureTracer{}
	interceptor := NewTracingInterceptor(tracer)

	if interceptor.Name() != tracingInterceptorName {
		t.Errorf("Name() = %q, want %q", interceptor.Name(), tracingInterceptorName)
	}
}

func TestTracingInterceptor_Order(t *testing.T) {
	tracer := &captureTracer{}
	interceptor := NewTracingInterceptor(tracer)

	if interceptor.Order() != 100 {
		t.Errorf("Order() = %d, want 100", interceptor.Order())
	}
}

func TestTracingInterceptor_Intercept_Success(t *testing.T) {
	tracer := &captureTracer{}
	interceptor := NewTracingInterceptor(tracer)

	ctx := &hyperionContext{
		Context: context.Background(),
		logger:  &noopLogger{},
		tracer:  tracer,
		db:      &noopExecutor{},
		meter:   &noOpMeter{},
	}

	fullPath := "UserService.GetUser"

	// Call Intercept
	newCtx, endFunc, err := interceptor.Intercept(ctx, fullPath)

	if err != nil {
		t.Errorf("Intercept() returned error: %v", err)
	}

	if newCtx == nil {
		t.Fatal("Intercept() returned nil context")
	}

	if endFunc == nil {
		t.Fatal("Intercept() returned nil end function")
	}

	// Verify span was started
	if len(tracer.startCalls) != 1 {
		t.Errorf("Expected 1 Start call, got %d", len(tracer.startCalls))
	}

	if tracer.startCalls[0].spanName != fullPath {
		t.Errorf("Expected span name %q, got %q", fullPath, tracer.startCalls[0].spanName)
	}

	// Verify span was created
	if len(tracer.spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(tracer.spans))
	}

	span := tracer.spans[0]

	// Call end function with success (no error)
	endFunc(nil)

	// Verify span was ended
	if !span.ended {
		t.Error("Span was not ended")
	}

	// Verify no errors were recorded
	if len(span.recordedErrors) != 0 {
		t.Errorf("Expected no recorded errors, got %d", len(span.recordedErrors))
	}
}

func TestTracingInterceptor_Intercept_WithError(t *testing.T) {
	tracer := &captureTracer{}
	interceptor := NewTracingInterceptor(tracer)

	ctx := &hyperionContext{
		Context: context.Background(),
		logger:  &noopLogger{},
		tracer:  tracer,
		db:      &noopExecutor{},
		meter:   &noOpMeter{},
	}

	fullPath := "UserService.CreateUser"

	// Call Intercept
	_, endFunc, err := interceptor.Intercept(ctx, fullPath)

	if err != nil {
		t.Errorf("Intercept() returned error: %v", err)
	}

	// Verify span was created
	if len(tracer.spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(tracer.spans))
	}

	span := tracer.spans[0]

	// Call end function with error
	testErr := errors.New("validation failed")
	endFunc(&testErr)

	// Verify error was recorded in span
	if len(span.recordedErrors) != 1 {
		t.Errorf("Expected 1 recorded error, got %d", len(span.recordedErrors))
	}

	if !errors.Is(span.recordedErrors[0], testErr) {
		t.Errorf("Expected error %v, got %v", testErr, span.recordedErrors[0])
	}

	// Verify span was ended
	if !span.ended {
		t.Error("Span was not ended")
	}
}

func TestTracingInterceptor_Intercept_NilErrorPointer(t *testing.T) {
	tracer := &captureTracer{}
	interceptor := NewTracingInterceptor(tracer)

	ctx := &hyperionContext{
		Context: context.Background(),
		logger:  &noopLogger{},
		tracer:  tracer,
		db:      &noopExecutor{},
		meter:   &noOpMeter{},
	}

	_, endFunc, _ := interceptor.Intercept(ctx, "Test.Method")

	span := tracer.spans[0]

	// Call end function with nil error pointer (success case)
	endFunc(nil)

	// Should end span without recording error
	if !span.ended {
		t.Error("Span was not ended")
	}

	if len(span.recordedErrors) != 0 {
		t.Errorf("Expected no recorded errors, got %d", len(span.recordedErrors))
	}
}

func TestTracingInterceptor_Intercept_ContextPreservation(t *testing.T) {
	tracer := &captureTracer{}
	interceptor := NewTracingInterceptor(tracer)

	originalLogger := &noopLogger{}
	originalDB := &noopExecutor{}
	originalMeter := &noOpMeter{}
	originalInterceptors := []Interceptor{&mockInterceptor{name: "test"}}

	ctx := &hyperionContext{
		Context:      context.Background(),
		logger:       originalLogger,
		tracer:       tracer,
		db:           originalDB,
		meter:        originalMeter,
		interceptors: originalInterceptors,
	}

	// Call Intercept
	newCtx, endFunc, _ := interceptor.Intercept(ctx, "Test.Method")

	// Verify all fields are preserved
	hctx, ok := newCtx.(*hyperionContext)
	if !ok {
		t.Fatal("Returned context is not *hyperionContext")
	}

	if hctx.logger != originalLogger {
		t.Error("Logger was not preserved")
	}

	if hctx.tracer != tracer {
		t.Error("Tracer was not preserved")
	}

	if hctx.db != originalDB {
		t.Error("DB was not preserved")
	}

	if hctx.meter != originalMeter {
		t.Error("Meter was not preserved")
	}

	if len(hctx.interceptors) != len(originalInterceptors) {
		t.Error("Interceptors were not preserved")
	}

	// Clean up
	endFunc(nil)
}

func TestTracingInterceptor_Intercept_FallbackPath(t *testing.T) {
	tracer := &captureTracer{}
	interceptor := NewTracingInterceptor(tracer)

	// Use a non-hyperionContext implementation
	ctx := &customContext{
		Context: context.Background(),
		logger:  &noopLogger{},
		db:      &noopExecutor{},
		tracer:  tracer,
		meter:   &noOpMeter{},
	}

	// Call Intercept
	newCtx, endFunc, err := interceptor.Intercept(ctx, "Test.Method")

	if err != nil {
		t.Errorf("Intercept() returned error: %v", err)
	}

	if newCtx == nil {
		t.Fatal("Intercept() returned nil context")
	}

	// Verify span was created and can be ended
	if len(tracer.spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(tracer.spans))
	}

	span := tracer.spans[0]

	endFunc(nil)

	if !span.ended {
		t.Error("Span was not ended in fallback path")
	}
}

// customContext is a non-hyperionContext implementation for testing fallback.
type customContext struct {
	context.Context
	logger Logger
	db     Executor
	tracer Tracer
	meter  Meter
}

func (c *customContext) Logger() Logger { return c.logger }
func (c *customContext) DB() Executor   { return c.db }
func (c *customContext) Tracer() Tracer { return c.tracer }
func (c *customContext) Meter() Meter   { return c.meter }
func (c *customContext) UseIntercept(parts ...any) (ctx Context, endFunc func(*error)) {
	return c, func(*error) {}
}
func (c *customContext) WithTimeout(timeout time.Duration) (Context, context.CancelFunc) {
	return c, func() {}
}
func (c *customContext) WithCancel() (Context, context.CancelFunc) {
	return c, func() {}
}
func (c *customContext) WithDeadline(deadline time.Time) (Context, context.CancelFunc) {
	return c, func() {}
}
