package otel

import (
	"fmt"

	"github.com/mapoio/hyperion"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// otelSpan wraps an OpenTelemetry span to implement hyperion.Span.
type otelSpan struct {
	span trace.Span
}

// End completes the span with optional end options.
func (s *otelSpan) End(options ...hyperion.SpanEndOption) {
	// OTel doesn't use end options in the same way, so we just end the span
	s.span.End()
}

// AddEvent adds an event to the span with optional event options.
func (s *otelSpan) AddEvent(name string, options ...hyperion.EventOption) {
	s.span.AddEvent(name)
}

// RecordError records an error on the span with optional event options.
func (s *otelSpan) RecordError(err error, options ...hyperion.EventOption) {
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

// SetAttributes sets attributes on the span.
func (s *otelSpan) SetAttributes(attributes ...hyperion.Attribute) {
	attrs := convertAttributes(attributes...)
	s.span.SetAttributes(attrs...)
}

// SpanContext returns the span's context information.
func (s *otelSpan) SpanContext() hyperion.SpanContext {
	sc := s.span.SpanContext()
	return &otelSpanContext{sc: sc}
}

// otelSpanContext wraps an OpenTelemetry SpanContext to implement hyperion.SpanContext.
type otelSpanContext struct {
	sc trace.SpanContext
}

// TraceID returns the trace ID as a string.
func (c *otelSpanContext) TraceID() string {
	return c.sc.TraceID().String()
}

// SpanID returns the span ID as a string.
func (c *otelSpanContext) SpanID() string {
	return c.sc.SpanID().String()
}

// IsValid returns whether this span context is valid.
func (c *otelSpanContext) IsValid() bool {
	return c.sc.IsValid()
}

// convertAttributes converts hyperion attributes to OTel attributes.
func convertAttributes(attrs ...hyperion.Attribute) []attribute.KeyValue {
	otelAttrs := make([]attribute.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		otelAttrs = append(otelAttrs, attribute.KeyValue{
			Key:   attribute.Key(attr.Key),
			Value: convertAttributeValue(attr.Value),
		})
	}
	return otelAttrs
}

// convertAttributeValue converts a hyperion attribute value to an OTel attribute value.
func convertAttributeValue(value any) attribute.Value {
	switch v := value.(type) {
	case bool:
		return attribute.BoolValue(v)
	case int:
		return attribute.IntValue(v)
	case int64:
		return attribute.Int64Value(v)
	case float64:
		return attribute.Float64Value(v)
	case string:
		return attribute.StringValue(v)
	case []bool:
		return attribute.BoolSliceValue(v)
	case []int:
		return attribute.IntSliceValue(v)
	case []int64:
		return attribute.Int64SliceValue(v)
	case []float64:
		return attribute.Float64SliceValue(v)
	case []string:
		return attribute.StringSliceValue(v)
	default:
		// Fallback to string representation
		return attribute.StringValue(fmt.Sprintf("%v", v))
	}
}
