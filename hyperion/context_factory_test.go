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

	factory := NewContextFactory(logger, tracer, db)
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
}

// TestContextFactory_WithDecorators tests decorator application.
func TestContextFactory_WithDecorators(t *testing.T) {
	logger := NewNoOpLogger()
	tracer := NewNoOpTracer()
	db := NewNoOpDatabase()

	// Create a simple logger decorator for testing
	decoratorApplied := false
	loggerDecorator := func(l Logger) Logger {
		decoratorApplied = true
		return l // Return same logger for test
	}

	factory := NewContextFactory(logger, tracer, db,
		WithLoggerDecorator(loggerDecorator),
	)

	// Create context - should apply decorator
	ctx := factory.New(context.Background())
	if ctx == nil {
		t.Fatal("Factory.New() should return a context")
	}

	if !decoratorApplied {
		t.Error("Logger decorator was not applied")
	}
}

// TestContextFactory_MultipleDecorators tests multiple decorator composition.
func TestContextFactory_MultipleDecorators(t *testing.T) {
	logger := NewNoOpLogger()
	tracer := NewNoOpTracer()
	db := NewNoOpDatabase()

	executionOrder := []string{}

	decorator1 := func(l Logger) Logger {
		executionOrder = append(executionOrder, "decorator1")
		return l
	}
	decorator2 := func(l Logger) Logger {
		executionOrder = append(executionOrder, "decorator2")
		return l
	}

	factory := NewContextFactory(logger, tracer, db,
		WithLoggerDecorator(decorator1, decorator2),
	)

	factory.New(context.Background())

	if len(executionOrder) != 2 {
		t.Fatalf("Expected 2 decorators to run, got %d", len(executionOrder))
	}
	if executionOrder[0] != "decorator1" {
		t.Errorf("Expected decorator1 first, got %s", executionOrder[0])
	}
	if executionOrder[1] != "decorator2" {
		t.Errorf("Expected decorator2 second, got %s", executionOrder[1])
	}
}

// TestContextFactory_ExecutorDecorator tests executor decorator.
func TestContextFactory_ExecutorDecorator(t *testing.T) {
	logger := NewNoOpLogger()
	tracer := NewNoOpTracer()
	db := NewNoOpDatabase()

	decoratorCalled := false
	executorDecorator := func(exec Executor) Executor {
		decoratorCalled = true
		return exec
	}

	factory := NewContextFactory(logger, tracer, db,
		WithExecutorDecorator(executorDecorator),
	)

	ctx := factory.New(context.Background())
	if ctx == nil {
		t.Fatal("Factory.New() should return a context")
	}

	if !decoratorCalled {
		t.Error("Executor decorator was not applied")
	}

	// Verify DB() returns decorated executor
	if ctx.DB() == nil {
		t.Error("Context should have a database executor")
	}
}

// TestContextFactory_AllDecoratorTypes tests all decorator types together.
func TestContextFactory_AllDecoratorTypes(t *testing.T) {
	logger := NewNoOpLogger()
	tracer := NewNoOpTracer()
	db := NewNoOpDatabase()

	loggerDecorated := false
	tracerDecorated := false
	executorDecorated := false

	factory := NewContextFactory(logger, tracer, db,
		WithLoggerDecorator(func(l Logger) Logger {
			loggerDecorated = true
			return l
		}),
		WithTracerDecorator(func(t Tracer) Tracer {
			tracerDecorated = true
			return t
		}),
		WithExecutorDecorator(func(e Executor) Executor {
			executorDecorated = true
			return e
		}),
	)

	ctx := factory.New(context.Background())
	if ctx == nil {
		t.Fatal("Factory.New() should return a context")
	}

	if !loggerDecorated {
		t.Error("Logger decorator was not applied")
	}
	if !tracerDecorated {
		t.Error("Tracer decorator was not applied")
	}
	if !executorDecorated {
		t.Error("Executor decorator was not applied")
	}
}
