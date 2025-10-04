package main

import (
	"context"
	"fmt"

	"github.com/mapoio/hyperion"
	"github.com/mapoio/hyperion/adapter/otel"
	"github.com/mapoio/hyperion/adapter/viper"
	"github.com/mapoio/hyperion/adapter/zap"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	// Create config
	cfg, err := viper.NewProviderFromEnv()
	if err != nil {
		panic(err)
	}

	// Create logger
	logger, err := zap.NewZapLogger(cfg)
	if err != nil {
		panic(err)
	}

	// Create tracer
	tracer, err := otel.NewOtelTracer(cfg)
	if err != nil {
		panic(err)
	}

	// Create database and meter
	db := hyperion.NewNoOpDatabase()
	meter, err := otel.NewOtelMeter(cfg)
	if err != nil {
		panic(err)
	}

	// Create factory
	factory := hyperion.NewContextFactory(logger, tracer, db, meter)

	// Create hyperion context
	hctx := factory.New(context.Background())

	// Start a span via tracer
	hctx, span := tracer.Start(hctx, "test-operation")
	defer span.End()

	// Extract standard context
	stdCtx := context.Context(hctx)

	// Check if span is in context
	extractedSpan := trace.SpanFromContext(stdCtx)
	fmt.Printf("Span is recording: %v\n", extractedSpan.IsRecording())
	fmt.Printf("Span context valid: %v\n", extractedSpan.SpanContext().IsValid())

	if extractedSpan.SpanContext().IsValid() {
		fmt.Printf("TraceID: %s\n", extractedSpan.SpanContext().TraceID().String())
		fmt.Printf("SpanID: %s\n", extractedSpan.SpanContext().SpanID().String())
	}

	// Try logging
	fmt.Println("\nLogging test:")
	hctx.Logger().Info("test message", "key", "value")
}
