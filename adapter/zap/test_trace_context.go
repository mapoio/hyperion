package zap

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestTraceContextRoundtrip(t *testing.T) {
	// Parse trace ID and span ID
	traceIDStr := "4ab3828f4f2bf47f24fe3b23b5df8d71"
	spanIDStr := "1e823dd17fcdec4c"

	var traceID trace.TraceID
	hex.Decode(traceID[:], []byte(traceIDStr))

	var spanID trace.SpanID
	hex.Decode(spanID[:], []byte(spanIDStr))

	// Create span context
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	fmt.Printf("Original SpanContext: %+v\n", spanCtx)
	fmt.Printf("  TraceID: %s\n", spanCtx.TraceID().String())
	fmt.Printf("  SpanID: %s\n", spanCtx.SpanID().String())
	fmt.Printf("  TraceFlags: %d\n", spanCtx.TraceFlags())

	// Inject into context
	ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)

	// Extract from context
	extractedSpanCtx := trace.SpanContextFromContext(ctx)

	fmt.Printf("\nExtracted SpanContext: %+v\n", extractedSpanCtx)
	fmt.Printf("  TraceID: %s\n", extractedSpanCtx.TraceID().String())
	fmt.Printf("  SpanID: %s\n", extractedSpanCtx.SpanID().String())
	fmt.Printf("  TraceFlags: %d\n", extractedSpanCtx.TraceFlags())

	// Verify
	if extractedSpanCtx.TraceID() != traceID {
		t.Errorf("TraceID mismatch: got %s, want %s", extractedSpanCtx.TraceID(), traceID)
	}
	if extractedSpanCtx.SpanID() != spanID {
		t.Errorf("SpanID mismatch: got %s, want %s", extractedSpanCtx.SpanID(), spanID)
	}
}
