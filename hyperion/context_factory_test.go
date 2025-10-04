package hyperion

import (
	"context"
	"testing"
)

// TestNewContextFactory tests the factory constructor.
func TestNewContextFactory(t *testing.T) {
	logger := NewNoOpLogger()
	tracer := NewNoOpTracer()
	db := NewNoOpDatabase()
	meter := NewNoOpMeter()

	factory := NewContextFactory(logger, tracer, db, meter)
	if factory == nil {
		t.Fatal("NewContextFactory should return a factory")
	}

	// Verify factory creates valid contexts
	ctx := factory.New(context.Background())
	if ctx == nil {
		t.Error("Factory.New() should return a context")
	}
	if ctx.Logger() == nil {
		t.Error("Context should have a logger")
	}
	if ctx.Tracer() == nil {
		t.Error("Context should have a tracer")
	}
	if ctx.DB() == nil {
		t.Error("Context should have a database executor")
	}
	if ctx.Meter() == nil {
		t.Error("Context should have a meter")
	}
}

// TestContextSpan tests the Context.Span() method behavior.
func TestContextSpan(t *testing.T) {
	t.Run("returns no-op span when no span is active", func(t *testing.T) {
		logger := NewNoOpLogger()
		tracer := NewNoOpTracer()
		db := NewNoOpDatabase().Executor()
		meter := NewNoOpMeter()

		ctx := New(context.Background(), logger, db, tracer, meter)

		// Get span from context without starting any span
		span := ctx.Span()

		if span == nil {
			t.Fatal("Context.Span() should not return nil")
		}

		// Verify it's a no-op span
		spanCtx := span.SpanContext()
		if spanCtx.IsValid() {
			t.Error("Expected no-op span to be invalid")
		}
	})

	t.Run("returns active span after tracer.Start()", func(t *testing.T) {
		logger := NewNoOpLogger()
		tracer := NewNoOpTracer()
		db := NewNoOpDatabase().Executor()
		meter := NewNoOpMeter()

		ctx := New(context.Background(), logger, db, tracer, meter)

		// Start a span
		newCtx, span := ctx.Tracer().Start(ctx, "test-operation")

		// Verify the span is set in context
		contextSpan := newCtx.Span()
		if contextSpan == nil {
			t.Fatal("Context.Span() should not return nil after tracer.Start()")
		}

		// Should return the same span object
		if contextSpan != span {
			t.Error("Context.Span() should return the span created by tracer.Start()")
		}
	})

	t.Run("WithSpan sets span in context", func(t *testing.T) {
		logger := NewNoOpLogger()
		tracer := NewNoOpTracer()
		db := NewNoOpDatabase().Executor()
		meter := NewNoOpMeter()

		ctx := New(context.Background(), logger, db, tracer, meter)

		// Create a no-op span
		span := &noopSpan{}

		// Set span in context
		newCtx := WithSpan(ctx, span)

		// Verify span is set
		contextSpan := newCtx.Span()
		if contextSpan != span {
			t.Error("WithSpan should set the span in context")
		}

		// Original context should still have no span
		originalSpan := ctx.Span()
		if originalSpan == span {
			t.Error("WithSpan should create a new context without modifying original")
		}
	})

	t.Run("span is preserved across WithXxx operations", func(t *testing.T) {
		logger := NewNoOpLogger()
		tracer := NewNoOpTracer()
		db := NewNoOpDatabase().Executor()
		meter := NewNoOpMeter()

		ctx := New(context.Background(), logger, db, tracer, meter)

		// Start a span
		ctxWithSpan, span := ctx.Tracer().Start(ctx, "test-operation")

		// Use various WithXxx helpers
		newLogger := NewNoOpLogger()
		ctxWithLogger := WithLogger(ctxWithSpan, newLogger)

		// Verify span is preserved
		if ctxWithLogger.Span() != span {
			t.Error("WithLogger should preserve the span")
		}

		// Test WithDB
		newDB := NewNoOpDatabase().Executor()
		ctxWithDB := WithDB(ctxWithLogger, newDB)

		if ctxWithDB.Span() != span {
			t.Error("WithDB should preserve the span")
		}

		// Test WithTracer
		newTracer := NewNoOpTracer()
		ctxWithTracer := WithTracer(ctxWithDB, newTracer)

		if ctxWithTracer.Span() != span {
			t.Error("WithTracer should preserve the span")
		}
	})
}
