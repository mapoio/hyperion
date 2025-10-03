package hyperion

import "testing"

// TestChainLoggerDecorators tests logger decorator chaining.
func TestChainLoggerDecorators(t *testing.T) {
	logger := NewNoOpLogger()
	executionOrder := []string{}

	decorator1 := func(l Logger) Logger {
		executionOrder = append(executionOrder, "logger1")
		return l
	}
	decorator2 := func(l Logger) Logger {
		executionOrder = append(executionOrder, "logger2")
		return l
	}

	chained := ChainLoggerDecorators(decorator1, decorator2)
	result := chained(logger)

	if result == nil {
		t.Fatal("Chained decorator should return a logger")
	}

	if len(executionOrder) != 2 {
		t.Fatalf("Expected 2 decorators to execute, got %d", len(executionOrder))
	}
	if executionOrder[0] != "logger1" || executionOrder[1] != "logger2" {
		t.Errorf("Decorator execution order incorrect: %v", executionOrder)
	}
}

// TestChainTracerDecorators tests tracer decorator chaining.
func TestChainTracerDecorators(t *testing.T) {
	tracer := NewNoOpTracer()
	executionOrder := []string{}

	decorator1 := func(t Tracer) Tracer {
		executionOrder = append(executionOrder, "tracer1")
		return t
	}
	decorator2 := func(t Tracer) Tracer {
		executionOrder = append(executionOrder, "tracer2")
		return t
	}

	chained := ChainTracerDecorators(decorator1, decorator2)
	result := chained(tracer)

	if result == nil {
		t.Fatal("Chained decorator should return a tracer")
	}

	if len(executionOrder) != 2 {
		t.Fatalf("Expected 2 decorators to execute, got %d", len(executionOrder))
	}
	if executionOrder[0] != "tracer1" || executionOrder[1] != "tracer2" {
		t.Errorf("Decorator execution order incorrect: %v", executionOrder)
	}
}

// TestChainExecutorDecorators tests executor decorator chaining.
func TestChainExecutorDecorators(t *testing.T) {
	executor := NewNoOpDatabase().Executor()
	executionOrder := []string{}

	decorator1 := func(e Executor) Executor {
		executionOrder = append(executionOrder, "executor1")
		return e
	}
	decorator2 := func(e Executor) Executor {
		executionOrder = append(executionOrder, "executor2")
		return e
	}

	chained := ChainExecutorDecorators(decorator1, decorator2)
	result := chained(executor)

	if result == nil {
		t.Fatal("Chained decorator should return an executor")
	}

	if len(executionOrder) != 2 {
		t.Fatalf("Expected 2 decorators to execute, got %d", len(executionOrder))
	}
	if executionOrder[0] != "executor1" || executionOrder[1] != "executor2" {
		t.Errorf("Decorator execution order incorrect: %v", executionOrder)
	}
}

// TestEmptyDecoratorChain tests chaining with no decorators.
func TestEmptyDecoratorChain(t *testing.T) {
	logger := NewNoOpLogger()

	chained := ChainLoggerDecorators()
	result := chained(logger)

	if result != logger {
		t.Error("Empty chain should return original logger")
	}
}

// TestSingleDecoratorChain tests chaining with single decorator.
func TestSingleDecoratorChain(t *testing.T) {
	logger := NewNoOpLogger()
	called := false

	decorator := func(l Logger) Logger {
		called = true
		return l
	}

	chained := ChainLoggerDecorators(decorator)
	result := chained(logger)

	if result == nil {
		t.Fatal("Chained decorator should return a logger")
	}
	if !called {
		t.Error("Decorator was not called")
	}
}
