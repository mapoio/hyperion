package otel

import (
	"context"

	"github.com/mapoio/hyperion"

	"go.opentelemetry.io/otel/trace"
)

// otelTracer wraps an OpenTelemetry tracer to implement hyperion.Tracer.
type otelTracer struct {
	tracer   trace.Tracer
	provider trace.TracerProvider
}

// Start creates a new span and returns a context with the span and the span itself.
func (t *otelTracer) Start(hctx hyperion.Context, spanName string, opts ...hyperion.SpanOption) (hyperion.Context, hyperion.Span) {
	otelOpts := convertSpanOpts(opts...)
	stdCtx, span := t.tracer.Start(hctx, spanName, otelOpts...)
	// Wrap the new standard context back into hyperion.Context
	newHctx := hyperion.WithContext(hctx, stdCtx)
	return newHctx, &otelSpan{span: span}
}

// Shutdown flushes any pending traces and shuts down the tracer provider.
func (t *otelTracer) Shutdown(ctx context.Context) error {
	if sp, ok := t.provider.(interface{ Shutdown(context.Context) error }); ok {
		return sp.Shutdown(ctx)
	}
	return nil
}

// convertSpanOpts converts hyperion span options to OTel span start options.
func convertSpanOpts(opts ...hyperion.SpanOption) []trace.SpanStartOption {
	// For now, we don't convert options since hyperion's option methods are private
	// This is a known limitation that will be addressed in a future update
	// TODO: Add option conversion when hyperion exposes option values
	return nil
}
