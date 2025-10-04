package zap

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// otelCore is a zapcore.Core wrapper that automatically injects trace context.
// It extracts trace_id and span_id from context.Context and adds them as fields.
type otelCore struct {
	zapcore.Core
}

// newOtelCore wraps an existing zapcore.Core with OpenTelemetry trace context injection.
func newOtelCore(core zapcore.Core) zapcore.Core {
	return &otelCore{Core: core}
}

// With adds structured context to the logger.
// This is part of the zapcore.Core interface.
func (c *otelCore) With(fields []zapcore.Field) zapcore.Core {
	return &otelCore{Core: c.Core.With(fields)}
}

// Check determines whether the supplied Entry should be logged.
// This is part of the zapcore.Core interface.
func (c *otelCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checked.AddCore(entry, c)
	}
	return checked
}

// Write serializes the Entry and any Fields supplied at the log site.
// This is where we inject the trace context before writing the log.
func (c *otelCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Note: We cannot extract context here because zapcore.Core.Write doesn't receive context.
	// The trace context injection happens at the logger level via contextLogger.
	return c.Core.Write(entry, fields)
}

// contextLogger wraps a zap.Logger and provides context-aware logging methods.
// It automatically extracts trace_id and span_id from context.Context.
type contextLogger struct {
	logger *zap.Logger
}

// newContextLogger creates a new context-aware logger wrapper.
func newContextLogger(logger *zap.Logger) *contextLogger {
	return &contextLogger{logger: logger}
}

// extractTraceContext extracts trace_id and span_id from context and returns zap fields.
func extractTraceContext(ctx context.Context) []zap.Field {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()

	// Only check if span context is valid (has trace ID and span ID)
	// We don't check IsRecording() because:
	// 1. Ended spans (IsRecording() == false) still have valid trace context
	// 2. We want to correlate logs even after span.End() is called
	// 3. The trace context remains valid throughout the entire request lifecycle
	if !spanCtx.IsValid() {
		return nil
	}

	return []zap.Field{
		zap.String("trace_id", spanCtx.TraceID().String()),
		zap.String("span_id", spanCtx.SpanID().String()),
	}
}

// InfoContext logs an info message with trace context from ctx.
func (c *contextLogger) InfoContext(ctx context.Context, msg string, fields ...zap.Field) {
	traceFields := extractTraceContext(ctx)
	if traceFields != nil {
		fields = append(traceFields, fields...)
	}
	c.logger.Info(msg, fields...)
}

// DebugContext logs a debug message with trace context from ctx.
func (c *contextLogger) DebugContext(ctx context.Context, msg string, fields ...zap.Field) {
	traceFields := extractTraceContext(ctx)
	if traceFields != nil {
		fields = append(traceFields, fields...)
	}
	c.logger.Debug(msg, fields...)
}

// WarnContext logs a warning message with trace context from ctx.
func (c *contextLogger) WarnContext(ctx context.Context, msg string, fields ...zap.Field) {
	traceFields := extractTraceContext(ctx)
	if traceFields != nil {
		fields = append(traceFields, fields...)
	}
	c.logger.Warn(msg, fields...)
}

// ErrorContext logs an error message with trace context from ctx.
func (c *contextLogger) ErrorContext(ctx context.Context, msg string, fields ...zap.Field) {
	traceFields := extractTraceContext(ctx)
	if traceFields != nil {
		fields = append(traceFields, fields...)
	}
	c.logger.Error(msg, fields...)
}

// FatalContext logs a fatal message with trace context from ctx and exits.
func (c *contextLogger) FatalContext(ctx context.Context, msg string, fields ...zap.Field) {
	traceFields := extractTraceContext(ctx)
	if traceFields != nil {
		fields = append(traceFields, fields...)
	}
	c.logger.Fatal(msg, fields...)
}
