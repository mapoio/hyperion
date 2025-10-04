package zap

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestTraceExtractionFromContext(t *testing.T) {
	// Create a tracer
	tracer := otel.Tracer("test")

	// Start a span
	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "test-operation")
	defer span.End()

	// Try to extract trace context
	fields := extractTraceContext(ctx)

	t.Logf("Extracted fields: %v", fields)
	t.Logf("Span is recording: %v", span.IsRecording())
	t.Logf("Span context is valid: %v", span.SpanContext().IsValid())

	if len(fields) == 0 {
		t.Error("Expected trace fields, got none")
	}
}
