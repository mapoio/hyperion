package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// TestTracerLoggerIntegration tests that trace context is properly propagated to logs.
// This verifies that when using a Tracer to create spans, the trace_id and span_id
// are available for loggers to extract and include in log output.
func TestTracerLoggerIntegration(t *testing.T) {
	// Reset provider before test
	resetProviderForTesting()

	// Create an in-memory span exporter for testing
	exporter := tracetest.NewInMemoryExporter()

	// Create a tracer provider with the test exporter
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)

	tracer := &OtelTracer{
		tracer:   tp.Tracer("test-integration"),
		provider: tp,
	}

	t.Run("trace context is extractable from span", func(t *testing.T) {
		ctx := context.Background()

		// Start a span using the tracer
		newCtx, span := tracer.Start(wrapContext(ctx), "test-operation")

		// Verify span is not nil
		if span == nil {
			t.Fatal("span should not be nil")
		}

		// Extract trace context from the span
		spanCtx := span.SpanContext()

		// Verify trace context is valid
		if !spanCtx.IsValid() {
			t.Fatal("span context should be valid")
		}

		// Verify we can extract trace_id and span_id
		traceID := spanCtx.TraceID()
		if traceID == "" {
			t.Error("trace_id should not be empty")
		}

		spanID := spanCtx.SpanID()
		if spanID == "" {
			t.Error("span_id should not be empty")
		}

		// Verify the context contains the span (using OTel API)
		extractedSpan := oteltrace.SpanFromContext(newCtx)
		extractedSpanCtx := extractedSpan.SpanContext()

		// Compare trace and span IDs
		if extractedSpanCtx.TraceID().String() != traceID {
			t.Error("extracted span trace_id should match original")
		}
		if extractedSpanCtx.SpanID().String() != spanID {
			t.Error("extracted span span_id should match original")
		}

		span.End()

		// Force flush
		if err := tp.ForceFlush(context.Background()); err != nil {
			t.Fatalf("failed to force flush: %v", err)
		}

		// Verify span was exported
		spans := exporter.GetSpans()
		if len(spans) != 1 {
			t.Fatalf("expected 1 span, got %d", len(spans))
		}

		if spans[0].Name != "test-operation" {
			t.Errorf("expected span name 'test-operation', got %q", spans[0].Name)
		}

		// Shutdown
		if err := tp.Shutdown(context.Background()); err != nil {
			t.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	})
}

// TestTraceContextExtraction tests extracting trace context for logging purposes.
// This simulates what a logger would do to inject trace_id and span_id into logs.
func TestTraceContextExtraction(t *testing.T) {
	// Reset provider before test
	resetProviderForTesting()

	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)

	tracer := &OtelTracer{
		tracer:   tp.Tracer("test-extraction"),
		provider: tp,
	}

	t.Run("extracts trace context for logging", func(t *testing.T) {
		ctx := context.Background()

		// Start a span
		newCtx, span := tracer.Start(wrapContext(ctx), "logged-operation")

		// Simulate logger extracting trace context using OTel API
		extractedSpan := oteltrace.SpanFromContext(newCtx)
		spanCtx := extractedSpan.SpanContext()

		// Verify span context is valid
		if !spanCtx.IsValid() {
			t.Fatal("span context should be valid")
		}

		// Create a simulated log entry with trace context
		logEntry := map[string]string{
			"message":  "test log message",
			"trace_id": spanCtx.TraceID().String(),
			"span_id":  spanCtx.SpanID().String(),
		}

		// Verify log entry contains trace context
		if logEntry["trace_id"] == "" {
			t.Error("log entry should contain trace_id")
		}

		if logEntry["span_id"] == "" {
			t.Error("log entry should contain span_id")
		}

		// Verify log entry can be serialized to JSON (for structured logging)
		jsonData, err := json.Marshal(logEntry)
		if err != nil {
			t.Fatalf("failed to marshal log entry: %v", err)
		}

		// Verify JSON contains trace context
		if !bytes.Contains(jsonData, []byte("trace_id")) {
			t.Error("JSON log should contain trace_id field")
		}

		if !bytes.Contains(jsonData, []byte("span_id")) {
			t.Error("JSON log should contain span_id field")
		}

		span.End()

		// Force flush and shutdown
		if err := tp.ForceFlush(context.Background()); err != nil {
			t.Fatalf("failed to force flush: %v", err)
		}

		if err := tp.Shutdown(context.Background()); err != nil {
			t.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	})
}

// TestMultipleSpansTraceContext tests that parent-child span relationships
// maintain consistent trace_id while having different span_ids.
func TestMultipleSpansTraceContext(t *testing.T) {
	// Reset provider before test
	resetProviderForTesting()

	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)

	tracer := &OtelTracer{
		tracer:   tp.Tracer("test-multi-span"),
		provider: tp,
	}

	t.Run("child spans share parent trace_id", func(t *testing.T) {
		ctx := context.Background()

		// Start parent span
		parentCtx, parentSpan := tracer.Start(wrapContext(ctx), "parent-operation")
		parentSpanCtx := parentSpan.SpanContext()

		// Start child span
		_, childSpan := tracer.Start(parentCtx, "child-operation")
		childSpanCtx := childSpan.SpanContext()

		// Verify both spans share the same trace_id
		if parentSpanCtx.TraceID() != childSpanCtx.TraceID() {
			t.Error("parent and child spans should share the same trace_id")
		}

		// Verify spans have different span_ids
		if parentSpanCtx.SpanID() == childSpanCtx.SpanID() {
			t.Error("parent and child spans should have different span_ids")
		}

		// End spans
		childSpan.End()
		parentSpan.End()

		// Force flush
		if err := tp.ForceFlush(context.Background()); err != nil {
			t.Fatalf("failed to force flush: %v", err)
		}

		// Verify both spans were exported
		spans := exporter.GetSpans()
		if len(spans) != 2 {
			t.Fatalf("expected 2 spans, got %d", len(spans))
		}

		// Shutdown
		if err := tp.Shutdown(context.Background()); err != nil {
			t.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	})
}
