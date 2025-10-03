# Decorator Pattern Refactoring Design

## Objective
Refactor decorator implementation to use generics for better extensibility and type safety.

## Current Problems
1. Hardcoded decorator types for each component
2. Duplicated Chain functions (3x almost identical code)
3. Users cannot extend with custom component decorators
4. Difficult to add new component types

## Solution: Generic Decorator with Registry

### Core Generic Types

```go
// Decorator wraps a component to add additional behavior (AOP pattern).
// T can be any component type: Logger, Tracer, Executor, Cache, or user-defined.
type Decorator[T any] func(T) T

// Type aliases for framework components (preserve readability)
type (
    LoggerDecorator   = Decorator[Logger]
    TracerDecorator   = Decorator[Tracer]
    ExecutorDecorator = Decorator[Executor]
)
```

### Generic Chain Function

```go
// Chain composes multiple decorators into one.
// Decorators are applied in order (left to right).
//
// Example:
//
//	decorator := hyperion.Chain[hyperion.Logger](
//	    AddPrefixDecorator("[APP]"),
//	    FilterByLevelDecorator(hyperion.InfoLevel),
//	)
//	logger = decorator(logger)
func Chain[T any](decorators ...Decorator[T]) Decorator[T] {
    if len(decorators) == 0 {
        return func(t T) T { return t }
    }
    return func(component T) T {
        for _, decorator := range decorators {
            component = decorator(component)
        }
        return component
    }
}
```

### Decorator Registry

```go
// DecoratorRegistry manages a collection of decorators for a specific component type.
// It allows dynamic registration and batch application of decorators.
//
// Example:
//
//	registry := hyperion.NewDecoratorRegistry[hyperion.Logger]()
//	registry.Register(AddPrefixDecorator("[APP]"))
//	registry.Register(FilterByLevelDecorator(hyperion.InfoLevel))
//	logger = registry.Apply(logger)
type DecoratorRegistry[T any] struct {
    decorators []Decorator[T]
}

// NewDecoratorRegistry creates a new decorator registry for type T.
func NewDecoratorRegistry[T any]() *DecoratorRegistry[T] {
    return &DecoratorRegistry[T]{
        decorators: make([]Decorator[T], 0),
    }
}

// Register adds a decorator to the registry.
func (r *DecoratorRegistry[T]) Register(decorator Decorator[T]) {
    r.decorators = append(r.decorators, decorator)
}

// Apply applies all registered decorators to the component in registration order.
func (r *DecoratorRegistry[T]) Apply(component T) T {
    return Chain(r.decorators...)(component)
}

// Clear removes all registered decorators.
func (r *DecoratorRegistry[T]) Clear() {
    r.decorators = r.decorators[:0]
}

// Count returns the number of registered decorators.
func (r *DecoratorRegistry[T]) Count() int {
    return len(r.decorators)
}
```

## API Changes

### Breaking Changes
- Remove: `ChainLoggerDecorators()`
- Remove: `ChainTracerDecorators()`
- Remove: `ChainExecutorDecorators()`

### New APIs
- Add: `Chain[T any](decorators ...Decorator[T]) Decorator[T]`
- Add: `NewDecoratorRegistry[T any]() *DecoratorRegistry[T]`
- Add: `DecoratorRegistry[T].Register(decorator Decorator[T])`
- Add: `DecoratorRegistry[T].Apply(component T) T`
- Add: `DecoratorRegistry[T].Clear()`
- Add: `DecoratorRegistry[T].Count() int`

### Migration Guide

**Before (Story 2.3)**:
```go
decorator := ChainLoggerDecorators(
    AddPrefixDecorator("[APP]"),
    FilterByLevelDecorator(InfoLevel),
)
logger = decorator(logger)
```

**After (Refactored)**:
```go
// Option 1: Direct chain
decorator := Chain[Logger](
    AddPrefixDecorator("[APP]"),
    FilterByLevelDecorator(InfoLevel),
)
logger = decorator(logger)

// Option 2: Registry (for dynamic scenarios)
registry := NewDecoratorRegistry[Logger]()
registry.Register(AddPrefixDecorator("[APP]"))
registry.Register(FilterByLevelDecorator(InfoLevel))
logger = registry.Apply(logger)
```

## User Extension Example

Users can now decorate custom components:

```go
// User-defined Cache interface
type MyCache interface {
    Get(key string) ([]byte, error)
    Set(key string, value []byte) error
}

// Use hyperion.Decorator for custom type
type MyCacheDecorator = hyperion.Decorator[MyCache]

func MetricsDecorator(metrics Metrics) MyCacheDecorator {
    return func(cache MyCache) MyCache {
        return &metricsCache{cache: cache, metrics: metrics}
    }
}

// Use hyperion.Chain with custom type
cache = hyperion.Chain[MyCache](
    MetricsDecorator(metrics),
    LoggingDecorator(logger),
)(cache)
```

## Testing Strategy

1. Update existing decorator tests to use `Chain[T]`
2. Add registry tests:
   - Register and apply decorators
   - Clear and count operations
   - Empty registry behavior
3. Add user extension test (custom component type)

## Files to Modify

1. `hyperion/context_decorator.go` - Refactor to generic implementation
2. `hyperion/context_decorator_test.go` - Update tests
3. `hyperion/context_factory.go` - Update to use type aliases
4. `hyperion/context_factory_test.go` - Update tests
5. `hyperion/noop_test.go` - Update decorator examples

## Implementation Steps

1. ✅ Design review and approval
2. [ ] Implement generic Decorator and Chain
3. [ ] Implement DecoratorRegistry
4. [ ] Remove old Chain functions
5. [ ] Update ContextFactory to use new types
6. [ ] Update all tests
7. [ ] Update documentation
8. [ ] Verify all tests pass
9. [ ] Commit refactoring

## Success Criteria

- ✅ Single generic Decorator[T] definition
- ✅ Single Chain[T] implementation (DRY)
- ✅ Users can decorate custom components
- ✅ DecoratorRegistry supports dynamic registration
- ✅ All tests pass with >= 90% coverage
- ✅ Documentation includes user extension examples
