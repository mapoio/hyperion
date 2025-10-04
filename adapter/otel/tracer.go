package otel

import (
	"context"

	"github.com/mapoio/hyperion"

	"go.opentelemetry.io/otel/trace"
)

// OtelTracer wraps an OpenTelemetry tracer to implement hyperion.Tracer.
// It is exported to allow applications to access the underlying TracerProvider
// for integrating OTel auto-instrumentation libraries.
//
//nolint:revive // Name is intentional to distinguish from hyperion.Tracer interface
type OtelTracer struct {
	tracer   trace.Tracer
	provider trace.TracerProvider
}

// TracerProvider returns the underlying OpenTelemetry TracerProvider.
// This allows applications to integrate OTel auto-instrumentation libraries
// (e.g., otelhttp, otelgrpc) that require access to the TracerProvider.
func (t *OtelTracer) TracerProvider() trace.TracerProvider {
	return t.provider
}

// Start creates a new span and returns a context with the span and the span itself.
func (t *OtelTracer) Start(hctx hyperion.Context, spanName string, opts ...hyperion.SpanOption) (hyperion.Context, hyperion.Span) {
	otelOpts := convertSpanOpts(opts...)
	// Extract the underlying context.Context from hyperion.Context
	// This is critical: OTel needs the standard context.Context to store span context
	stdCtx, span := t.tracer.Start(context.Context(hctx), spanName, otelOpts...)

	// Wrap the span
	wrappedSpan := &otelSpan{span: span}

	// Update context with both the new standard context and the span
	newHctx := hyperion.WithContext(hctx, stdCtx)
	newHctx = hyperion.WithSpan(newHctx, wrappedSpan)

	return newHctx, wrappedSpan
}

// Shutdown flushes any pending traces and shuts down the tracer provider.
func (t *OtelTracer) Shutdown(ctx context.Context) error {
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
