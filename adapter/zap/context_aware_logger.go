package zap

import (
	"context"

	"github.com/mapoio/hyperion"
)

// contextAwareLogger wraps zapLogger and automatically injects trace context.
// It extracts the underlying context.Context from hyperion.Context when available.
type contextAwareLogger struct {
	zapLogger *zapLogger
	stdCtx    context.Context // The underlying context.Context for trace extraction
}

// newContextAwareLogger creates a context-aware logger that automatically
// extracts trace context from the embedded context.Context.
func newContextAwareLogger(ctx context.Context, zapLogger *zapLogger) hyperion.Logger {
	return &contextAwareLogger{
		zapLogger: zapLogger,
		stdCtx:    ctx,
	}
}

// Debug logs a debug message with trace context automatically injected.
func (c *contextAwareLogger) Debug(msg string, fields ...any) {
	zapFields := convertToZapFields(fields...)
	c.zapLogger.contextLogger.DebugContext(c.stdCtx, msg, zapFields...)
}

// Info logs an info message with trace context automatically injected.
func (c *contextAwareLogger) Info(msg string, fields ...any) {
	zapFields := convertToZapFields(fields...)
	c.zapLogger.contextLogger.InfoContext(c.stdCtx, msg, zapFields...)
}

// Warn logs a warning message with trace context automatically injected.
func (c *contextAwareLogger) Warn(msg string, fields ...any) {
	zapFields := convertToZapFields(fields...)
	c.zapLogger.contextLogger.WarnContext(c.stdCtx, msg, zapFields...)
}

// Error logs an error message with trace context automatically injected.
func (c *contextAwareLogger) Error(msg string, fields ...any) {
	zapFields := convertToZapFields(fields...)
	c.zapLogger.contextLogger.ErrorContext(c.stdCtx, msg, zapFields...)
}

// Fatal logs a fatal message with trace context automatically injected and exits.
func (c *contextAwareLogger) Fatal(msg string, fields ...any) {
	zapFields := convertToZapFields(fields...)
	c.zapLogger.contextLogger.FatalContext(c.stdCtx, msg, zapFields...)
}

// With creates a child logger with additional fields.
func (c *contextAwareLogger) With(fields ...any) hyperion.Logger {
	childLogger, ok := c.zapLogger.With(fields...).(*zapLogger)
	if !ok {
		// This should never happen since zapLogger.With() always returns *zapLogger
		return c
	}
	return newContextAwareLogger(c.stdCtx, childLogger)
}

// WithError creates a child logger with an error field.
func (c *contextAwareLogger) WithError(err error) hyperion.Logger {
	return c.With("error", err)
}

// SetLevel changes the log level dynamically.
func (c *contextAwareLogger) SetLevel(level hyperion.LogLevel) {
	c.zapLogger.SetLevel(level)
}

// GetLevel returns the current log level.
func (c *contextAwareLogger) GetLevel() hyperion.LogLevel {
	return c.zapLogger.GetLevel()
}

// Sync flushes any buffered log entries.
func (c *contextAwareLogger) Sync() error {
	return c.zapLogger.Sync()
}
