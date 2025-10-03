package hyperion

import "testing"

// TestChain_Logger tests generic Chain with Logger type.
func TestChain_Logger(t *testing.T) {
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

	chained := Chain[Logger](decorator1, decorator2)
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

// TestChain_Tracer tests generic Chain with Tracer type.
func TestChain_Tracer(t *testing.T) {
	tracer := NewNoOpTracer()
	executionOrder := []string{}

	decorator1 := func(tr Tracer) Tracer {
		executionOrder = append(executionOrder, "tracer1")
		return tr
	}
	decorator2 := func(tr Tracer) Tracer {
		executionOrder = append(executionOrder, "tracer2")
		return tr
	}

	chained := Chain[Tracer](decorator1, decorator2)
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

// TestChain_Executor tests generic Chain with Executor type.
func TestChain_Executor(t *testing.T) {
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

	chained := Chain[Executor](decorator1, decorator2)
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

// TestChain_Empty tests Chain with no decorators.
func TestChain_Empty(t *testing.T) {
	logger := NewNoOpLogger()

	chained := Chain[Logger]()
	result := chained(logger)

	if result != logger {
		t.Error("Empty chain should return original logger")
	}
}

// TestChain_Single tests Chain with single decorator.
func TestChain_Single(t *testing.T) {
	logger := NewNoOpLogger()
	called := false

	decorator := func(l Logger) Logger {
		called = true
		return l
	}

	chained := Chain[Logger](decorator)
	result := chained(logger)

	if result == nil {
		t.Fatal("Chained decorator should return a logger")
	}
	if !called {
		t.Error("Decorator was not called")
	}
}

// mockUserComponent is a test helper for user-defined component decorator tests.
type mockUserComponent struct{}

func (m *mockUserComponent) Method() string {
	return "mock"
}

// TestChain_UserDefinedType tests Chain with user-defined component type.
func TestChain_UserDefinedType(t *testing.T) {
	// Define a simple user component type
	type UserComponent interface {
		Method() string
	}

	component := &mockUserComponent{}
	executionOrder := []string{}

	// User decorator 1
	decorator1 := func(c UserComponent) UserComponent {
		executionOrder = append(executionOrder, "user-dec1")
		return c
	}

	// User decorator 2
	decorator2 := func(c UserComponent) UserComponent {
		executionOrder = append(executionOrder, "user-dec2")
		return c
	}

	// Use generic Chain with user-defined type
	chained := Chain[UserComponent](decorator1, decorator2)
	result := chained(component)

	if result == nil {
		t.Fatal("Chained decorator should return a component")
	}

	if len(executionOrder) != 2 {
		t.Fatalf("Expected 2 decorators to execute, got %d", len(executionOrder))
	}
	if executionOrder[0] != "user-dec1" || executionOrder[1] != "user-dec2" {
		t.Errorf("Decorator execution order incorrect: %v", executionOrder)
	}
}

// TestDecoratorRegistry_Basic tests basic registry operations.
func TestDecoratorRegistry_Basic(t *testing.T) {
	registry := NewDecoratorRegistry[Logger]()

	if registry.Count() != 0 {
		t.Errorf("Expected empty registry, got count %d", registry.Count())
	}

	// Register decorators
	decorator1 := func(l Logger) Logger { return l }
	decorator2 := func(l Logger) Logger { return l }

	registry.Register(decorator1)
	if registry.Count() != 1 {
		t.Errorf("Expected count 1, got %d", registry.Count())
	}

	registry.Register(decorator2)
	if registry.Count() != 2 {
		t.Errorf("Expected count 2, got %d", registry.Count())
	}

	// Clear registry
	registry.Clear()
	if registry.Count() != 0 {
		t.Errorf("Expected count 0 after clear, got %d", registry.Count())
	}
}

// TestDecoratorRegistry_Apply tests applying registered decorators.
func TestDecoratorRegistry_Apply(t *testing.T) {
	logger := NewNoOpLogger()
	registry := NewDecoratorRegistry[Logger]()
	executionOrder := []string{}

	decorator1 := func(l Logger) Logger {
		executionOrder = append(executionOrder, "reg1")
		return l
	}
	decorator2 := func(l Logger) Logger {
		executionOrder = append(executionOrder, "reg2")
		return l
	}

	registry.Register(decorator1)
	registry.Register(decorator2)

	result := registry.Apply(logger)

	if result == nil {
		t.Fatal("Registry.Apply should return a logger")
	}

	if len(executionOrder) != 2 {
		t.Fatalf("Expected 2 decorators to execute, got %d", len(executionOrder))
	}
	if executionOrder[0] != "reg1" || executionOrder[1] != "reg2" {
		t.Errorf("Decorator execution order incorrect: %v", executionOrder)
	}
}

// TestDecoratorRegistry_ApplyEmpty tests applying with empty registry.
func TestDecoratorRegistry_ApplyEmpty(t *testing.T) {
	logger := NewNoOpLogger()
	registry := NewDecoratorRegistry[Logger]()

	result := registry.Apply(logger)

	if result != logger {
		t.Error("Empty registry should return original logger")
	}
}

// TestDecoratorRegistry_MultipleTypes tests registries for different types.
func TestDecoratorRegistry_MultipleTypes(t *testing.T) {
	// Logger registry
	loggerRegistry := NewDecoratorRegistry[Logger]()
	loggerRegistry.Register(func(l Logger) Logger { return l })

	// Tracer registry
	tracerRegistry := NewDecoratorRegistry[Tracer]()
	tracerRegistry.Register(func(tr Tracer) Tracer { return tr })

	// Executor registry
	executorRegistry := NewDecoratorRegistry[Executor]()
	executorRegistry.Register(func(e Executor) Executor { return e })

	if loggerRegistry.Count() != 1 {
		t.Errorf("Logger registry count should be 1, got %d", loggerRegistry.Count())
	}
	if tracerRegistry.Count() != 1 {
		t.Errorf("Tracer registry count should be 1, got %d", tracerRegistry.Count())
	}
	if executorRegistry.Count() != 1 {
		t.Errorf("Executor registry count should be 1, got %d", executorRegistry.Count())
	}
}

// TestTypeAliases tests that type aliases work correctly.
func TestTypeAliases(t *testing.T) {
	// Verify that type aliases are usable
	var loggerDec LoggerDecorator = func(l Logger) Logger { return l }
	var tracerDec TracerDecorator = func(tr Tracer) Tracer { return tr }
	var executorDec ExecutorDecorator = func(e Executor) Executor { return e }

	// Apply decorators
	logger := loggerDec(NewNoOpLogger())
	tracer := tracerDec(NewNoOpTracer())
	executor := executorDec(NewNoOpDatabase().Executor())

	if logger == nil {
		t.Error("LoggerDecorator should work")
	}
	if tracer == nil {
		t.Error("TracerDecorator should work")
	}
	if executor == nil {
		t.Error("ExecutorDecorator should work")
	}
}
