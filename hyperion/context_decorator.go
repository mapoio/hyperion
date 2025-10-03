package hyperion

// LoggerDecorator wraps a Logger to add additional behavior.
// This enables AOP-style cross-cutting concerns like prefixing, filtering, or routing.
//
// Example: Add prefix to all log messages
//
//	func AddPrefixDecorator(prefix string) LoggerDecorator {
//	    return func(logger Logger) Logger {
//	        return &prefixLogger{logger: logger, prefix: prefix}
//	    }
//	}
type LoggerDecorator func(Logger) Logger

// TracerDecorator wraps a Tracer to add additional behavior.
// This enables AOP-style tracing enhancements like adding default attributes.
//
// Example: Add service name to all spans
//
//	func AddServiceNameDecorator(serviceName string) TracerDecorator {
//	    return func(tracer Tracer) Tracer {
//	        return &serviceNameTracer{tracer: tracer, serviceName: serviceName}
//	    }
//	}
type TracerDecorator func(Tracer) Tracer

// ExecutorDecorator wraps an Executor to add additional behavior.
// This enables AOP-style database operations like query logging, metrics, or caching.
//
// Example: Log all queries
//
//	func QueryLoggingDecorator(logger Logger) ExecutorDecorator {
//	    return func(executor Executor) Executor {
//	        return &queryLoggingExecutor{executor: executor, logger: logger}
//	    }
//	}
type ExecutorDecorator func(Executor) Executor

// ChainLoggerDecorators composes multiple LoggerDecorators into one.
// Decorators are applied in the order they are provided (left to right).
//
// Example:
//
//	decorator := ChainLoggerDecorators(
//	    AddPrefixDecorator("[APP]"),
//	    FilterByLevelDecorator(InfoLevel),
//	)
//	logger = decorator(logger)
func ChainLoggerDecorators(decorators ...LoggerDecorator) LoggerDecorator {
	return func(logger Logger) Logger {
		for _, decorator := range decorators {
			logger = decorator(logger)
		}
		return logger
	}
}

// ChainTracerDecorators composes multiple TracerDecorators into one.
// Decorators are applied in the order they are provided (left to right).
func ChainTracerDecorators(decorators ...TracerDecorator) TracerDecorator {
	return func(tracer Tracer) Tracer {
		for _, decorator := range decorators {
			tracer = decorator(tracer)
		}
		return tracer
	}
}

// ChainExecutorDecorators composes multiple ExecutorDecorators into one.
// Decorators are applied in the order they are provided (left to right).
//
// Example:
//
//	decorator := ChainExecutorDecorators(
//	    QueryLoggingDecorator(logger),
//	    QueryMetricsDecorator(metrics),
//	)
//	executor = decorator(executor)
func ChainExecutorDecorators(decorators ...ExecutorDecorator) ExecutorDecorator {
	return func(executor Executor) Executor {
		for _, decorator := range decorators {
			executor = decorator(executor)
		}
		return executor
	}
}
