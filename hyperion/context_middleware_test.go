package hyperion

import (
	"context"
	"errors"
	"testing"
)

// TestChainMiddleware tests middleware chaining.
func TestChainMiddleware(t *testing.T) {
	ctx := New(context.Background(), NewNoOpLogger(), NewNoOpDatabase().Executor(), NewNoOpTracer())
	executionOrder := []string{}

	middleware1 := func(c Context, next func(Context) error) error {
		executionOrder = append(executionOrder, "mw1-before")
		err := next(c)
		executionOrder = append(executionOrder, "mw1-after")
		return err
	}

	middleware2 := func(c Context, next func(Context) error) error {
		executionOrder = append(executionOrder, "mw2-before")
		err := next(c)
		executionOrder = append(executionOrder, "mw2-after")
		return err
	}

	handler := func(c Context) error {
		executionOrder = append(executionOrder, "handler")
		return nil
	}

	chain := ChainMiddleware(middleware1, middleware2)
	err := chain(ctx, handler)

	if err != nil {
		t.Errorf("ChainMiddleware should not return error, got %v", err)
	}

	expected := []string{
		"mw1-before",
		"mw2-before",
		"handler",
		"mw2-after",
		"mw1-after",
	}

	if len(executionOrder) != len(expected) {
		t.Fatalf("Expected %d execution steps, got %d", len(expected), len(executionOrder))
	}

	for i, step := range expected {
		if executionOrder[i] != step {
			t.Errorf("Step %d: expected %s, got %s", i, step, executionOrder[i])
		}
	}
}

// TestChainMiddleware_ErrorPropagation tests error handling in middleware chain.
func TestChainMiddleware_ErrorPropagation(t *testing.T) {
	ctx := New(context.Background(), NewNoOpLogger(), NewNoOpDatabase().Executor(), NewNoOpTracer())
	expectedError := errors.New("handler error")

	middleware1 := func(c Context, next func(Context) error) error {
		err := next(c)
		// Middleware can observe error
		if err == nil {
			t.Error("Middleware should receive error from handler")
		}
		return err
	}

	handler := func(c Context) error {
		return expectedError
	}

	chain := ChainMiddleware(middleware1)
	err := chain(ctx, handler)

	if !errors.Is(err, expectedError) {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

// TestChainMiddleware_ContextModification tests context modification in middleware.
func TestChainMiddleware_ContextModification(t *testing.T) {
	ctx := New(context.Background(), NewNoOpLogger(), NewNoOpDatabase().Executor(), NewNoOpTracer())

	// Middleware modifies context by adding a logger
	middleware := func(c Context, next func(Context) error) error {
		newLogger := NewNoOpLogger()
		modifiedCtx := WithLogger(c, newLogger)
		return next(modifiedCtx)
	}

	var capturedLogger Logger
	handler := func(c Context) error {
		capturedLogger = c.Logger()
		return nil
	}

	chain := ChainMiddleware(middleware)
	err := chain(ctx, handler)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Handler should receive modified logger
	if capturedLogger == nil {
		t.Error("Handler should receive modified logger")
	}

	// Original context should be unchanged
	if ctx.Logger() == capturedLogger {
		t.Error("Original context should not be modified")
	}
}

// TestApplyMiddleware tests single middleware application.
func TestApplyMiddleware(t *testing.T) {
	ctx := New(context.Background(), NewNoOpLogger(), NewNoOpDatabase().Executor(), NewNoOpTracer())
	middlewareCalled := false

	middleware := func(c Context, next func(Context) error) error {
		middlewareCalled = true
		return next(c)
	}

	handler := func(c Context) error {
		return nil
	}

	err := ApplyMiddleware(middleware, ctx, handler)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !middlewareCalled {
		t.Error("Middleware was not called")
	}
}

// TestEmptyMiddlewareChain tests chain with no middleware.
func TestEmptyMiddlewareChain(t *testing.T) {
	ctx := New(context.Background(), NewNoOpLogger(), NewNoOpDatabase().Executor(), NewNoOpTracer())
	handlerCalled := false

	handler := func(c Context) error {
		handlerCalled = true
		return nil
	}

	chain := ChainMiddleware()
	err := chain(ctx, handler)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Error("Handler should be called even with empty middleware chain")
	}
}

// TestSingleMiddleware tests chain with single middleware.
func TestSingleMiddleware(t *testing.T) {
	ctx := New(context.Background(), NewNoOpLogger(), NewNoOpDatabase().Executor(), NewNoOpTracer())
	middlewareCalled := false
	handlerCalled := false

	middleware := func(c Context, next func(Context) error) error {
		middlewareCalled = true
		return next(c)
	}

	handler := func(c Context) error {
		handlerCalled = true
		return nil
	}

	chain := ChainMiddleware(middleware)
	err := chain(ctx, handler)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !middlewareCalled {
		t.Error("Middleware was not called")
	}
	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

// TestMiddlewareFunc_Adapter tests MiddlewareFunc adapter.
func TestMiddlewareFunc_Adapter(t *testing.T) {
	ctx := New(context.Background(), NewNoOpLogger(), NewNoOpDatabase().Executor(), NewNoOpTracer())
	called := false

	fn := MiddlewareFunc(func(c Context, next func(Context) error) error {
		called = true
		return next(c)
	})

	middleware := fn.Middleware()
	err := middleware(ctx, func(c Context) error { return nil })

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !called {
		t.Error("MiddlewareFunc was not called")
	}
}
