package hyperion

// noopTracer is a no-op implementation of Tracer interface.
type noopTracer struct{}

// NewNoOpTracer creates a new no-op Tracer implementation.
func NewNoOpTracer() Tracer {
	return &noopTracer{}
}

func (t *noopTracer) Start(ctx Context, spanName string, opts ...SpanOption) (Context, Span) {
	return ctx, &noopSpan{}
}

// noopSpan is a no-op implementation of Span interface.
type noopSpan struct{}

func (s *noopSpan) End(opts ...SpanEndOption)                  {}
func (s *noopSpan) SetAttributes(attrs ...Attribute)           {}
func (s *noopSpan) RecordError(err error, opts ...EventOption) {}
func (s *noopSpan) AddEvent(name string, opts ...EventOption)  {}
func (s *noopSpan) SpanContext() SpanContext                   { return &noopSpanContext{} }

// noopSpanContext is a no-op implementation of SpanContext interface.
type noopSpanContext struct{}

func (sc *noopSpanContext) TraceID() string { return "" }
func (sc *noopSpanContext) SpanID() string  { return "" }
func (sc *noopSpanContext) IsValid() bool   { return false }
