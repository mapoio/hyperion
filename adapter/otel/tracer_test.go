package otel


import (
	"context"
	"testing"

	"github.com/mapoio/hyperion"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)


func TestOtelTracer_Start(t *testing.T) {
	// Create an in-memory span exporter for testing
	exporter := tracetest.NewInMemoryExporter()

	// Create a tracer provider with the test exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)

	tracer := &OtelTracer{
		tracer:   tp.Tracer("test"),
		provider: tp,
	}

	ctx := context.Background()

	// Test basic span creation
	t.Run("creates span successfully", func(t *testing.T) {
		newCtx, span := tracer.Start(wrapContext(ctx), "test-span")

		if span == nil {
			t.Fatal("expected non-nil span")
		}

		if newCtx == nil {
			t.Fatal("expected non-nil context")
		}

		// End the span
		span.End()

		// Force flush before shutdown
		if err := tp.ForceFlush(context.Background()); err != nil {
			t.Fatalf("failed to force flush: %v", err)
		}

		// Verify span was recorded
		spans := exporter.GetSpans()
		if len(spans) != 1 {
			t.Fatalf("expected 1 span, got %d", len(spans))
		}

		if spans[0].Name != "test-span" {
			t.Errorf("expected span name 'test-span', got %q", spans[0].Name)
		}

		// Shutdown
		if err := tp.Shutdown(context.Background()); err != nil {
			t.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	})
}

func TestOtelTracer_StartWithAttributes(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)

	tracer := &OtelTracer{
		tracer:   tp.Tracer("test"),
		provider: tp,
	}

	ctx := context.Background()

	t.Run("creates span with attributes", func(t *testing.T) {
		_, span := tracer.Start(wrapContext(ctx), "test-span-with-attrs")

		// Set attributes after creation
		span.SetAttributes(
			hyperion.String("key1", "value1"),
			hyperion.Int64("key2", 42),
			hyperion.Bool("key3", true),
		)

		span.End()

		// Force flush
		if err := tp.ForceFlush(context.Background()); err != nil {
			t.Fatalf("failed to force flush: %v", err)
		}

		spans := exporter.GetSpans()
		if len(spans) != 1 {
			t.Fatalf("expected 1 span, got %d", len(spans))
		}

		// Verify attributes
		attrs := spans[0].Attributes
		if len(attrs) != 3 {
			t.Errorf("expected 3 attributes, got %d", len(attrs))
		}

		// Shutdown
		if err := tp.Shutdown(context.Background()); err != nil {
			t.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	})
}

func TestOtelSpan_RecordError(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)

	tracer := &OtelTracer{
		tracer:   tp.Tracer("test"),
		provider: tp,
	}

	ctx := context.Background()

	t.Run("records error on span", func(t *testing.T) {
		_, span := tracer.Start(wrapContext(ctx), "test-error-span")

		// Record an error
		testErr := context.DeadlineExceeded
		span.RecordError(testErr)

		span.End()

		// Force flush
		if err := tp.ForceFlush(context.Background()); err != nil {
			t.Fatalf("failed to force flush: %v", err)
		}

		spans := exporter.GetSpans()
		if len(spans) != 1 {
			t.Fatalf("expected 1 span, got %d", len(spans))
		}

		// Verify error was recorded
		events := spans[0].Events
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		if events[0].Name != "exception" {
			t.Errorf("expected event name 'exception', got %q", events[0].Name)
		}

		// Shutdown
		if err := tp.Shutdown(context.Background()); err != nil {
			t.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	})
}

func TestOtelSpan_AddEvent(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)

	tracer := &OtelTracer{
		tracer:   tp.Tracer("test"),
		provider: tp,
	}

	ctx := context.Background()

	t.Run("adds event to span", func(t *testing.T) {
		_, span := tracer.Start(wrapContext(ctx), "test-event-span")

		// Add an event
		span.AddEvent("test-event")

		span.End()

		// Force flush
		if err := tp.ForceFlush(context.Background()); err != nil {
			t.Fatalf("failed to force flush: %v", err)
		}

		spans := exporter.GetSpans()
		if len(spans) != 1 {
			t.Fatalf("expected 1 span, got %d", len(spans))
		}

		// Verify event was added
		events := spans[0].Events
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		if events[0].Name != "test-event" {
			t.Errorf("expected event name 'test-event', got %q", events[0].Name)
		}

		// Shutdown
		if err := tp.Shutdown(context.Background()); err != nil {
			t.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	})
}

func TestOtelSpanContext(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)

	tracer := &OtelTracer{
		tracer:   tp.Tracer("test"),
		provider: tp,
	}

	ctx := context.Background()

	t.Run("returns valid span context", func(t *testing.T) {
		_, span := tracer.Start(wrapContext(ctx), "test-span-context")

		spanCtx := span.SpanContext()

		if spanCtx == nil {
			t.Fatal("expected non-nil span context")
		}

		if !spanCtx.IsValid() {
			t.Error("expected valid span context")
		}

		traceID := spanCtx.TraceID()
		if traceID == "" {
			t.Error("expected non-empty trace ID")
		}

		spanID := spanCtx.SpanID()
		if spanID == "" {
			t.Error("expected non-empty span ID")
		}

		span.End()

		if err := tp.Shutdown(context.Background()); err != nil {
			t.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	})
}

func TestOtelTracer_Shutdown(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)

	tracer := &OtelTracer{
		tracer:   tp.Tracer("test"),
		provider: tp,
	}

	t.Run("shutdown succeeds", func(t *testing.T) {
		err := tracer.Shutdown(context.Background())
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
