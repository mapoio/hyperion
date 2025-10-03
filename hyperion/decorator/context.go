package decorator

// Note: Type aliases for hyperion components (LoggerDecorator, TracerDecorator, etc.)
// are defined in github.com/mapoio/hyperion/context_decorator.go to avoid import cycles.

// DecoratorRegistry provides dynamic decorator management for a specific component type.
// It allows decorators to be registered and applied at runtime.
//
// Example:
//
//	registry := NewDecoratorRegistry[hyperion.Logger]()
//	registry.Register(AddPrefixDecorator("[APP]"))
//	registry.Register(FilterByLevelDecorator(hyperion.InfoLevel))
//	logger = registry.Apply(logger)
//
//nolint:revive // DecoratorRegistry is the intended name for clarity
type DecoratorRegistry[T any] struct {
	decorators []Decorator[T]
}

// NewDecoratorRegistry creates a new registry for decorators of type T.
func NewDecoratorRegistry[T any]() *DecoratorRegistry[T] {
	return &DecoratorRegistry[T]{
		decorators: make([]Decorator[T], 0),
	}
}

// Register adds a decorator to the registry.
// Decorators are applied in the order they are registered.
func (r *DecoratorRegistry[T]) Register(decorator Decorator[T]) {
	r.decorators = append(r.decorators, decorator)
}

// Apply applies all registered decorators to the given component.
// Returns the decorated component.
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
