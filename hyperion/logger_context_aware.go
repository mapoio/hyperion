package hyperion

import "context"

// ContextAwareLogger is an optional interface that Logger implementations
// can implement to support automatic trace context injection.
//
// When hyperion.Context.Logger() is called, if the logger implements this
// interface, WithContext() will be called to bind the context, enabling
// automatic trace_id and span_id injection into logs.
//
// Example implementation in adapter/zap:
//
//	func (l *zapLogger) WithContext(ctx context.Context) Logger {
//	    return newContextAwareLogger(l, ctx)
//	}
type ContextAwareLogger interface {
	Logger

	// WithContext returns a new Logger bound to the given context.
	// The returned logger should automatically extract and inject trace
	// context (trace_id, span_id) from the context into all log entries.
	WithContext(ctx context.Context) Logger
}
