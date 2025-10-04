package main

import (
	"context"
	"fmt"

	"github.com/mapoio/hyperion"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	// Create simple OTel tracer
	otelTracer := otel.Tracer("test")
	
	// Test 1: Direct OTel usage
	fmt.Println("=== Test 1: Direct OTel ===")
	ctx1 := context.Background()
	ctx1, span1 := otelTracer.Start(ctx1, "direct-operation")
	defer span1.End()
	
	extractedSpan1 := trace.SpanFromContext(ctx1)
	fmt.Printf("Direct - Span is recording: %v\n", extractedSpan1.IsRecording())
	fmt.Printf("Direct - Span context valid: %v\n", extractedSpan1.SpanContext().IsValid())
	
	// Test 2: Through hyperion.Context
	fmt.Println("\n=== Test 2: Through hyperion.Context ===")
	logger := hyperion.NewNoOpLogger()
	db := hyperion.NewNoOpDatabase().Executor()
	meter := hyperion.NewNoOpMeter()
	tracer := hyperion.NewNoOpTracer()
	
	hctx := hyperion.New(context.Background(), logger, db, tracer, meter)
	fmt.Printf("hyperion.Context type: %T\n", hctx)
	
	// Extract standard context
	stdCtx := context.Context(hctx)
	fmt.Printf("Std context type: %T\n", stdCtx)
	
	// Start span with standard context extracted from hyperion.Context
	stdCtx2, span2 := otelTracer.Start(stdCtx, "hyperion-operation")
	defer span2.End()
	
	extractedSpan2 := trace.SpanFromContext(stdCtx2)
	fmt.Printf("Hyperion - Span is recording: %v\n", extractedSpan2.IsRecording())
	fmt.Printf("Hyperion - Span context valid: %v\n", extractedSpan2.SpanContext().IsValid())
}
