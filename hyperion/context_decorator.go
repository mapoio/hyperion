package hyperion

// Decorator wraps a component to add additional behavior using the AOP pattern.
// T can be any component type: Logger, Tracer, Executor, Cache, or user-defined types.
//
// This generic design allows users to extend Hyperion with custom component decorators
// without modifying the framework.
//
// Example with built-in Logger:
//
//	func AddPrefixDecorator(prefix string) hyperion.Decorator[hyperion.Logger] {
//	    return func(logger hyperion.Logger) hyperion.Logger {
//	        return &prefixLogger{logger: logger, prefix: prefix}
//	    }
//	}
//
//	logger = hyperion.Chain[hyperion.Logger](
//	    AddPrefixDecorator("[APP]"),
//	    FilterByLevelDecorator(hyperion.InfoLevel),
//	)(logger)
//
// Example with user-defined component:
//
//	type MyCache interface {
//	    Get(key string) ([]byte, error)
//	    Set(key string, value []byte) error
//	}
//
//	func MetricsDecorator(metrics Metrics) hyperion.Decorator[MyCache] {
//	    return func(cache MyCache) MyCache {
//	        return &metricsCache{cache: cache, metrics: metrics}
//	    }
//	}
//
//	cache = hyperion.Chain[MyCache](
//	    MetricsDecorator(metrics),
//	    LoggingDecorator(logger),
//	)(cache)
type Decorator[T any] func(T) T

// Type aliases for framework components to improve API readability.
// Users can use these aliases or the generic Decorator[T] directly.
type (
	// LoggerDecorator wraps a Logger to add additional behavior.
	// Example: prefixing, filtering, or routing log messages.
	LoggerDecorator = Decorator[Logger]

	// TracerDecorator wraps a Tracer to add additional behavior.
	// Example: adding default span attributes or filtering spans.
	TracerDecorator = Decorator[Tracer]

	// ExecutorDecorator wraps an Executor to add additional behavior.
	// Example: query logging, metrics collection, or caching.
	ExecutorDecorator = Decorator[Executor]
)

// Chain composes multiple decorators into a single decorator.
// Decorators are applied in the order they are provided (left to right).
//
// If no decorators are provided, returns a no-op decorator that returns
// the component unchanged.
//
// Example with explicit type parameter:
//
//	decorator := hyperion.Chain[hyperion.Logger](
//	    AddPrefixDecorator("[APP]"),
//	    FilterByLevelDecorator(hyperion.InfoLevel),
//	)
//	logger = decorator(logger)
//
// Example with type inference (when type can be inferred from context):
//
//	var loggerDec hyperion.Decorator[hyperion.Logger]
//	loggerDec = hyperion.Chain(
//	    AddPrefixDecorator("[APP]"),
//	    FilterByLevelDecorator(hyperion.InfoLevel),
//	)
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

// DecoratorRegistry manages a collection of decorators for a specific component type.
// It allows dynamic registration and batch application of decorators, which is useful
// for plugin-based architectures or configuration-driven decorator composition.
//
// Example:
//
//	// Create registry for Logger decorators
//	registry := hyperion.NewDecoratorRegistry[hyperion.Logger]()
//
//	// Register decorators dynamically
//	if config.EnablePrefix {
//	    registry.Register(AddPrefixDecorator(config.Prefix))
//	}
//	if config.MinLevel != "" {
//	    registry.Register(FilterByLevelDecorator(config.MinLevel))
//	}
//
//	// Apply all registered decorators
//	logger = registry.Apply(logger)
//
// Thread-safety: DecoratorRegistry is NOT thread-safe. If you need to register
// decorators from multiple goroutines, you must use external synchronization.
type DecoratorRegistry[T any] struct {
	decorators []Decorator[T]
}

// NewDecoratorRegistry creates a new decorator registry for component type T.
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

// Apply applies all registered decorators to the component in registration order.
// If no decorators are registered, returns the component unchanged.
func (r *DecoratorRegistry[T]) Apply(component T) T {
	return Chain(r.decorators...)(component)
}

// Clear removes all registered decorators from the registry.
func (r *DecoratorRegistry[T]) Clear() {
	r.decorators = r.decorators[:0]
}

// Count returns the number of registered decorators.
func (r *DecoratorRegistry[T]) Count() int {
	return len(r.decorators)
}
