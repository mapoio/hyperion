package hyperion

import (
	"context"
	"time"
)

// Tracer is the distributed tracing abstraction.
// It follows OpenTelemetry semantics but doesn't depend on it.
type Tracer interface {
	// Start creates a new span and returns a new context with the span.
	Start(ctx context.Context, spanName string, opts ...SpanOption) (context.Context, Span)
}

// Span represents a single span in a trace.
type Span interface {
	// End completes the span.
	End(opts ...SpanEndOption)

	// SetAttributes sets attributes on the span.
	SetAttributes(attrs ...Attribute)

	// RecordError records an error on the span.
	RecordError(err error, opts ...EventOption)

	// AddEvent adds an event to the span.
	AddEvent(name string, opts ...EventOption)

	// SpanContext returns the span's context (trace ID, span ID, etc.)
	SpanContext() SpanContext
}

// SpanContext contains trace identification information.
type SpanContext interface {
	// TraceID returns the trace ID as a string.
	TraceID() string

	// SpanID returns the span ID as a string.
	SpanID() string

	// IsValid returns whether this span context is valid.
	IsValid() bool
}

// Attribute represents a key-value pair for span metadata.
type Attribute struct {
	Value any
	Key   string
}

// Helper functions for creating attributes

// String creates a string-valued attribute.
func String(key, value string) Attribute {
	return Attribute{Key: key, Value: value}
}

// Int creates an int-valued attribute.
func Int(key string, value int) Attribute {
	return Attribute{Key: key, Value: value}
}

// Int64 creates an int64-valued attribute.
func Int64(key string, value int64) Attribute {
	return Attribute{Key: key, Value: value}
}

// Float64 creates a float64-valued attribute.
func Float64(key string, value float64) Attribute {
	return Attribute{Key: key, Value: value}
}

// Bool creates a bool-valued attribute.
func Bool(key string, value bool) Attribute {
	return Attribute{Key: key, Value: value}
}

// SpanOption configures a span at start time.
type SpanOption interface {
	applySpanStart(*spanConfig)
}

type spanConfig struct {
	Timestamp  time.Time
	Attributes []Attribute
	SpanKind   SpanKind
}

// SpanKind represents the role of a span in a trace.
type SpanKind int

const (
	// SpanKindInternal indicates the span represents an internal operation.
	SpanKindInternal SpanKind = iota

	// SpanKindServer indicates the span covers server-side handling of a request.
	SpanKindServer

	// SpanKindClient indicates the span describes a request to a remote service.
	SpanKindClient

	// SpanKindProducer indicates the span describes the initiator of an asynchronous request.
	SpanKindProducer

	// SpanKindConsumer indicates the span describes the consumer of an asynchronous request.
	SpanKindConsumer
)

// WithAttributes returns a SpanOption that sets attributes on a span.
func WithAttributes(attrs ...Attribute) SpanOption {
	return spanOptionFunc(func(cfg *spanConfig) {
		cfg.Attributes = append(cfg.Attributes, attrs...)
	})
}

// WithSpanKind returns a SpanOption that sets the span kind.
func WithSpanKind(kind SpanKind) SpanOption {
	return spanOptionFunc(func(cfg *spanConfig) {
		cfg.SpanKind = kind
	})
}

// WithTimestamp returns a SpanOption that sets the span start time.
func WithTimestamp(t time.Time) SpanOption {
	return spanOptionFunc(func(cfg *spanConfig) {
		cfg.Timestamp = t
	})
}

type spanOptionFunc func(*spanConfig)

func (f spanOptionFunc) applySpanStart(cfg *spanConfig) {
	f(cfg)
}

// SpanEndOption configures a span at end time.
type SpanEndOption interface {
	applySpanEnd(*spanEndConfig)
}

type spanEndConfig struct {
	Timestamp time.Time
}

// WithEndTime returns a SpanEndOption that sets the span end time.
func WithEndTime(t time.Time) SpanEndOption {
	return spanEndOptionFunc(func(cfg *spanEndConfig) {
		cfg.Timestamp = t
	})
}

type spanEndOptionFunc func(*spanEndConfig)

func (f spanEndOptionFunc) applySpanEnd(cfg *spanEndConfig) {
	f(cfg)
}

// EventOption configures an event.
type EventOption interface {
	applyEvent(*eventConfig)
}

type eventConfig struct {
	Timestamp  time.Time
	Attributes []Attribute
}

// WithEventAttributes returns an EventOption that sets event attributes.
func WithEventAttributes(attrs ...Attribute) EventOption {
	return eventOptionFunc(func(cfg *eventConfig) {
		cfg.Attributes = append(cfg.Attributes, attrs...)
	})
}

// WithEventTimestamp returns an EventOption that sets the event timestamp.
func WithEventTimestamp(t time.Time) EventOption {
	return eventOptionFunc(func(cfg *eventConfig) {
		cfg.Timestamp = t
	})
}

type eventOptionFunc func(*eventConfig)

func (f eventOptionFunc) applyEvent(cfg *eventConfig) {
	f(cfg)
}
