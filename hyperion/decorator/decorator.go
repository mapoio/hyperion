package decorator

// Decorator wraps a component to add additional behavior using the AOP pattern.
// T can be any component type: Logger, Tracer, Executor, Cache, or user-defined types.
//
// This generic design allows users to extend Hyperion with custom component decorators
// without modifying the framework.
//
// Example with built-in Logger:
//
//	func AddPrefixDecorator(prefix string) decorator.Decorator[hyperion.Logger] {
//	    return func(logger hyperion.Logger) hyperion.Logger {
//	        return &prefixLogger{logger: logger, prefix: prefix}
//	    }
//	}
//
//	logger = decorator.Chain[hyperion.Logger](
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
//	func MetricsDecorator(metrics Metrics) decorator.Decorator[MyCache] {
//	    return func(cache MyCache) MyCache {
//	        return &metricsCache{cache: cache, metrics: metrics}
//	    }
//	}
//
//	cache = decorator.Chain[MyCache](
//	    MetricsDecorator(metrics),
//	    LoggingDecorator(logger),
//	)(cache)
type Decorator[T any] func(T) T

// Chain composes multiple decorators into a single decorator.
// Decorators are applied in the order they are provided (left to right).
//
// If no decorators are provided, returns a no-op decorator that returns
// the component unchanged.
//
// Example with explicit type parameter:
//
//	decorator := decorator.Chain[hyperion.Logger](
//	    AddPrefixDecorator("[APP]"),
//	    FilterByLevelDecorator(hyperion.InfoLevel),
//	)
//	logger = decorator(logger)
//
// Example with type inference (when type can be inferred from context):
//
//	var loggerDec decorator.Decorator[hyperion.Logger]
//	loggerDec = decorator.Chain(
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
