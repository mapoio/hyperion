package hyperion

import "github.com/mapoio/hyperion/decorator"

// Re-export decorator types for backward compatibility.
// These aliases allow existing code to continue using hyperion.Decorator[T]
// while the implementation lives in the decorator package.
type (
	// Decorator wraps a component to add additional behavior using the AOP pattern.
	// See decorator.Decorator for full documentation.
	Decorator[T any] = decorator.Decorator[T]

	// DecoratorRegistry provides dynamic decorator management.
	// See decorator.DecoratorRegistry for full documentation.
	DecoratorRegistry[T any] = decorator.DecoratorRegistry[T]

	// Type aliases for framework components to improve API readability.
	// Example: prefixing, filtering, or routing log messages.
	LoggerDecorator = Decorator[Logger]

	// TracerDecorator wraps a Tracer to add additional behavior.
	// Example: adding default span attributes or filtering spans.
	TracerDecorator = Decorator[Tracer]

	// ExecutorDecorator wraps an Executor to add additional behavior.
	// Example: query logging, metrics, or read/write splitting.
	ExecutorDecorator = Decorator[Executor]

	// CacheDecorator wraps a Cache to add additional behavior.
	// Example: metrics, logging, or multi-tier caching.
	CacheDecorator = Decorator[Cache]
)

// Chain composes multiple decorators into a single decorator.
// This is a convenience re-export of decorator.Chain.
func Chain[T any](decorators ...Decorator[T]) Decorator[T] {
	return decorator.Chain(decorators...)
}

// NewDecoratorRegistry creates a new decorator registry.
// This is a convenience re-export of decorator.NewDecoratorRegistry.
func NewDecoratorRegistry[T any]() *DecoratorRegistry[T] {
	return decorator.NewDecoratorRegistry[T]()
}
